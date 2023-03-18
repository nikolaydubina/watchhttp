package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
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

const doc string = `
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
		fmt.Fprint(flag.CommandLine.Output(), doc)
		flag.PrintDefaults()
	}

	cmdargs, hasFlags := args.GetCommandFromArgs(os.Args[1:])

	var (
		port            int           = 9000
		interval        time.Duration = time.Second
		contentTypeJSON bool          = false
		isDelta         bool          = false
	)

	if hasFlags {
		flag.IntVar(&port, "p", port, "port")
		flag.DurationVar(&interval, "t", interval, `interval to execute command (units: ns, us, µs, ms, s, m, h, d, w, y)`)
		flag.BoolVar(&contentTypeJSON, "json", contentTypeJSON, "set Content-Type: application/json")
		flag.BoolVar(&isDelta, "d", isDelta, "show animated HTML delta difference (only JSON)")
		flag.Parse()
	}

	if len(cmdargs) == 0 {
		log.Fatal("missing command")
	}

	log.Printf("serving at port=%d with interval=%v latest STDOUT of command: %v\n", port, interval, strings.Join(cmdargs, " "))

	cmdrunner := &CmdRunner{
		ticker:     time.NewTicker(interval),
		lastStdout: bytes.NewBuffer(nil),
		mtx:        &sync.RWMutex{},
		cmd:        cmdargs,
	}
	go cmdrunner.Run()

	h := ForwardHandler{
		provider: cmdrunner,
		interval: interval,
	}

	if isDelta && contentTypeJSON {
		h.provider = &JSONHTMLRenderBridge{
			provider: cmdrunner,
			renderer: htmldelta.NewJSONRenderer(html.EscapeString(strings.Join(cmdargs, " "))),
			b:        &bytes.Buffer{},
			mtx:      &sync.Mutex{},
		}
	}
	if isDelta {
		h.contentType = "text/html; charset=utf-8"
	} else if contentTypeJSON {
		h.contentType = "application/json"
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
	w.Header().Set("Refresh", fmt.Sprintf("%.0f", s.interval.Seconds()))
	if _, err := s.provider.WriteTo(w); err != nil {
		log.Printf("error: %s", err)
	}
}

// JSONHTMLRenderBridge passes data from raw JSON provider to HTML JSON delta renderer and write output to destination.
// It caches rendered delta HTML JSON because delta HTML JSON renderer is not idempotent.
type JSONHTMLRenderBridge struct {
	renderer *htmldelta.JSONRenderer
	provider *CmdRunner
	b        *bytes.Buffer
	ts       time.Time
	mtx      *sync.Mutex
}

func (s *JSONHTMLRenderBridge) WriteTo(w io.Writer) (written int64, err error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if ts := s.provider.LastUpdatedAt(); ts.After(s.ts) {
		s.ts = ts
		s.b.Reset()
		s.renderer.From(bytes.NewReader(s.provider.LastStdout())).WriteTo(s.b)
	}
	// to not drain buffer accessing its bytes
	return io.Copy(w, bytes.NewReader(s.b.Bytes()))
}

func (s *JSONHTMLRenderBridge) LastUpdatedAt() time.Time { return s.ts }

// CmdRunner runs command on interval and stores last STDOUT in buffer
type CmdRunner struct {
	ticker     *time.Ticker
	cmd        []string
	lastStdout *bytes.Buffer
	ts         time.Time
	mtx        *sync.RWMutex
}

func (s *CmdRunner) LastUpdatedAt() time.Time { return s.ts }

func (s *CmdRunner) LastStdout() []byte {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.lastStdout.Bytes()
}

func (s *CmdRunner) WriteTo(w io.Writer) (written int64, err error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	// to not drain buffer accessing its bytes
	return io.Copy(w, bytes.NewReader(s.lastStdout.Bytes()))
}

func (s *CmdRunner) Run() {
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

		s.lastStdout.Reset()

		s.ts = time.Now()
		if _, err := io.Copy(s.lastStdout, stdout); err != nil {
			log.Fatal(err)
		}

		s.mtx.Unlock()

		if err := cmd.Wait(); err != nil {
			log.Fatal(err)
		}
	}
}
