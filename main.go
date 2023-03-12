package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nikolaydubina/watchhttp/internal/args"
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
	)

	if hasFlags {
		flag.IntVar(&port, "p", port, "port")
		flag.DurationVar(&interval, "t", interval, `interval to execute command (units: ns, us, µs, ms, s, m, h, d, w, y)`)
		flag.BoolVar(&contentTypeJSON, "json", contentTypeJSON, "set Content-Type: application/json")
		flag.Parse()
	}

	if len(cmdargs) == 0 {
		log.Fatal("missing command")
	}

	log.Printf("serving at port=%d with interval=%v latest STDOUT of command: %v\n", port, interval, strings.Join(cmdargs, " "))

	runner := CmdRunner{
		ticker:     time.NewTicker(interval),
		lastStdOut: bytes.NewBuffer(nil),
		mtx:        &sync.RWMutex{},
		cmd:        cmdargs,
	}
	go runner.Run()

	runnerHandler := ForwardHandler{
		Provider: &runner,
	}
	if contentTypeJSON {
		runnerHandler.ContentType = "application/json"
	}

	http.HandleFunc("/", runnerHandler.handleRequest)
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

// ForwardHandler will call Payload from wrapped class and serve it in response
type ForwardHandler struct {
	ContentType string
	Provider    interface {
		WritePayload(writer io.Writer) (int64, error)
	}
}

func (s ForwardHandler) handleRequest(w http.ResponseWriter, req *http.Request) {
	if s.ContentType != "" {
		w.Header().Set("Content-Type", s.ContentType)
	}
	if _, err := s.Provider.WritePayload(w); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		log.Printf("error: %s", err)
		return
	}
}

// CmdRunner runs command on interval and stores last STDOUT in buffer
type CmdRunner struct {
	ticker     *time.Ticker
	cmd        []string
	lastStdOut *bytes.Buffer
	mtx        *sync.RWMutex
}

func (s *CmdRunner) WritePayload(writer io.Writer) (bytesWritten int64, err error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return io.Copy(writer, bytes.NewReader(s.lastStdOut.Bytes()))
}

func (s *CmdRunner) Run() {
	for {
		<-s.ticker.C

		cmd := exec.Command(s.cmd[0], s.cmd[1:]...)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}

		if err := cmd.Start(); err != nil {
			log.Fatal(err)
		}

		s.mtx.Lock()

		s.lastStdOut.Reset()

		if _, err := io.Copy(s.lastStdOut, stdout); err != nil {
			log.Fatal(err)
		}

		s.mtx.Unlock()

		if err := cmd.Wait(); err != nil {
			log.Fatal(err)
		}
	}
}
