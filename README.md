# watchhttp

Execute command on timeout and serve its latest STDOUT at HTTP endpoint

```bash
$ watchhttp -t 1s -p 9000 --path /home -txt -- ls -la
$ watchhttp -t 5s -p 9000 --path /home -json -- cat myfile.json
$ watchhttp -t 5s -p 9000 --path /home -json -- kubectl get pod mypod -o=json
$ watchhttp -t 5s -p 9000 --path /home -json -- curl ...
$ watchhttp -t 5s -p 9000 --path /home -json -- /bin/sh -c 'curl ... | jq'
$ watchhttp -t 5s -p 9000 --path /home -- kubectl get pod mypod
$ watchhttp -t 5s -p 9000 --path /home -- graph
$ watchhttp kubectl get pod mypod
```
print one line what it does when starts to STDERR

### TODO

Rich Difference embedded?

HTML + reload
https://www.w3schools.com/jsref/met_loc_reload.asp

### Existing Tools

- as of 2023-03-12, [awesome-go](http://github.com/avelino/awesome-go) does not mention tools that can do this
- `netcat` can not do this

### Alternative: File Server + Bash

Similar effect can be achieved with file server and bash, albeit headers would not be set.

```go
package main

import "net/http"

func main() { http.ListenAndServe(":9000", http.FileServer(http.Dir("."))) }
```
```bash
$ while sleep 5; do <something> > <file> ; done
```

### Paths Not Taken

> Expose STDIN as HTTP endpoint

The problem is how to differentate separate responses?
Is empty line valid separator of responses?
There are clearly interesting eaxmples with UNIX pipes and `tail -f`, logs, WebSockets.
However, that may need separate tool.

> Stream `top`, k9s, [datadash](https://github.com/keithknott26/datadash) to browser in HTML

First, those tools re-render terminal output at their own interval.
It would take some effort, but effectivelly this is emulating terminal rendering in browser.
Issue of converting terminal escaped ASCII to HTML output colors.
As next step, you would want to also pass to STDIN through browser too.
Overall, this is separate problem of exposing terminal throuhg browser.
