package main

import (
	"bytes"
	_ "embed"
	"errors"
	"flag"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nikolaydubina/watchhttp/args"
	"github.com/nikolaydubina/watchhttp/htmldelta"
)

const doc = `
Run command periodically and expose latest STDOUT as HTTP endpoint

Examples:
$ watchhttp -t 1s -p 9000 -- ls -la
$ watchhttp vmstat
$ watchhttp tail /var/log/system.log
$ watchhttp -json -- cat myfile.json
$ watchhttp -p 9000 -json -- kubectl get pod mypod -o=json
$ watchhttp -p 9000 -- kubectl get pod mypod -o=yaml
$ watchhttp curl ...
$ watchhttp -json -- /bin/sh -c 'curl ... | jq'

Command options:
`

func main() {
	flag.Usage = func() {
		flag.CommandLine.Output().Write([]byte(doc))
		flag.PrintDefaults()
	}

	cmdargs, hasFlags := args.GetCommandFromArgs(os.Args[1:])

	var (
		port            int           = 9000
		interval        time.Duration = time.Second
		contentTypeJSON bool          = false
		contentTypeYAML bool          = false
		isDelta         bool          = false
	)

	if hasFlags {
		flag.IntVar(&port, "p", port, "port")
		flag.DurationVar(&interval, "t", interval, `interval to execute command (units: ns, us, µs, ms, s, m, h, d, w, y)`)
		flag.BoolVar(&contentTypeJSON, "json", contentTypeJSON, "set Content-Type: application/json")
		flag.BoolVar(&contentTypeYAML, "yaml", contentTypeYAML, "set Content-Type: application/yaml")
		flag.BoolVar(&isDelta, "d", isDelta, "show animated HTML delta difference (only JSON)")
		flag.Parse()
	}

	if len(cmdargs) == 0 {
		log.Fatal("missing command")
	}
	if contentTypeJSON && contentTypeYAML {
		log.Fatal("either json or yaml can be set as true")
	}

	log.Printf("serving at port=%d with interval=%v latest STDOUT of command: %v\n", port, interval, strings.Join(cmdargs, " "))

	commandRunner := &CommandRunner{
		ticker:     time.NewTicker(interval),
		lastStdout: bytes.NewBuffer(nil),
		mtx:        &sync.RWMutex{},
		cmd:        cmdargs,
	}
	go commandRunner.Run()

	h := ForwardHandler{
		provider: commandRunner,
		interval: interval,
	}

	if isDelta {
		var r interface {
			FromTo(r io.Reader, w io.Writer) error
		}
		title := html.EscapeString(strings.Join(cmdargs, " "))
		switch {
		case contentTypeJSON:
			r = htmldelta.NewJSONRenderer(title)
		case contentTypeYAML:
			r = htmldelta.NewYAMLRenderer(title)
		}
		h.provider = &RenderBridge{
			provider: commandRunner,
			renderer: r,
			raw:      bytes.NewBuffer(nil),
			out:      bytes.NewBuffer(nil),
			mtx:      &sync.RWMutex{},
		}
	}

	switch {
	case isDelta:
		h.contentType = "text/html; charset=utf-8"
	case contentTypeJSON:
		h.contentType = "application/json"
	case contentTypeYAML:
		h.contentType = "text/yaml"
	}

	http.HandleFunc("/", h.handleRequest)
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

// ForwardHandler will call Payload from wrapped class and serve it in response
type ForwardHandler struct {
	contentType string
	interval    time.Duration
	provider    interface {
		io.WriterTo
		LastUpdatedAt() time.Time
	}
}

func (s ForwardHandler) handleRequest(w http.ResponseWriter, req *http.Request) {
	if s.contentType != "" {
		w.Header().Set("Content-Type", s.contentType)
	}
	w.Header().Set("Last-Modified", s.provider.LastUpdatedAt().UTC().Format(http.TimeFormat))
	w.Header().Set("Refresh", strconv.Itoa(int(s.interval.Seconds())))
	if _, err := s.provider.WriteTo(w); err != nil {
		log.Printf("error: %s", err)
	}
}

// RenderBridge passes data from raw provider to renderer.
// Caches renderer output, in case renderer is not idempotent.
type RenderBridge struct {
	renderer interface {
		FromTo(r io.Reader, w io.Writer) error
	}
	provider interface {
		LastUpdatedAt() time.Time
		WriteTo(w io.Writer) (written int64, err error)
	}
	raw *bytes.Buffer
	out *bytes.Buffer
	ts  time.Time
	mtx *sync.RWMutex
}

func (s *RenderBridge) refresh() error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.ts = s.provider.LastUpdatedAt()

	// copy data, in case provider will modify data
	s.raw.Reset()
	if _, err := s.provider.WriteTo(s.raw); err != nil {
		return err
	}
	s.out.Reset()
	if err := s.renderer.FromTo(s.raw, s.out); err != nil {
		return err
	}

	return nil
}

func (s *RenderBridge) WriteTo(w io.Writer) (written int64, err error) {
	if s.LastUpdatedAt().Before(s.provider.LastUpdatedAt()) {
		if err := s.refresh(); err != nil {
			return 0, err
		}
	}

	s.mtx.RLock()
	defer s.mtx.RUnlock()
	n, err := w.Write(s.out.Bytes())
	return int64(n), err
}

func (s *RenderBridge) LastUpdatedAt() time.Time {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.ts
}

// CommandRunner on interval and store last STDOUT
type CommandRunner struct {
	ticker     *time.Ticker
	cmd        []string
	lastStdout *bytes.Buffer
	ts         time.Time
	mtx        *sync.RWMutex
}

func (s *CommandRunner) LastUpdatedAt() time.Time {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.ts
}

func (s *CommandRunner) WriteTo(w io.Writer) (written int64, err error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	n, err := w.Write(s.lastStdout.Bytes())
	return int64(n), err
}

func (s *CommandRunner) Run() {
	for range s.ticker.C {
		cmd := exec.Command(s.cmd[0], s.cmd[1:]...)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}

		if err := cmd.Start(); err != nil {
			log.Fatal(err)
		}

		s.mtx.Lock()

		s.ts = time.Now()
		s.lastStdout.Reset()
		if _, err := s.lastStdout.ReadFrom(stdout); err != nil && !errors.Is(err, io.EOF) {
			log.Fatal(err)
		}

		s.mtx.Unlock()

		if err := cmd.Wait(); err != nil {
			log.Fatal(err)
		}
	}
}
