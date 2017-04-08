# picam
Raspberry Pi Network Camera

Requires go and the go-tool esc from mjibson
```
go install -v github.com/mjibson/esc
```

Compile an arm version for a Pi with
```
go generate && GOARCH=arm go build
```
