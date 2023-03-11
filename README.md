# watchhttp

Execute command on timeout and serve its latest STDOUT at HTTP endpoint

```bash
$ watchhttp -t 1s -p 9000 --path /home -txt -- ls -la
$ watchhttp -t 5s -p 9000 --path /home -json -- cat asdf
```

TODO

Rich Difference embedded?

### Alternative

Similar effect can be achieved with following, albeit headers would not be set.

Start file server
```go
package main

import "net/http"

func main() { http.ListenAndServe(":9000", http.FileServer(http.Dir("."))) }
```
Write output to file on timeout
```bash
$ while sleep 1; do <something> > <file> ; done
```

### Existing Tools

- as of 2023-03-12, go-awesome does not mention tools that can do this
- `netcat` can not do this
