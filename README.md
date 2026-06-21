# IncTUI: Terminal UI for Incus

IncTUI provies a TUI on top of the local Incus API (via unix socket).

Current features:

* Display status, live CPU usage, live memory usage
* Start/Stop containers

[![asciicast](https://asciinema.org/a/MWp6F8po15TZU2e5.svg)](https://asciinema.org/a/acCBHs8LGylTIQ1L)

## Incus (Linux containers) on MacOS

I recommend installing Incus with `Colima`:

```bash
brew install incus 
brew install colima 
colima start --runtime=incus
```
## Installation 

### Clone, build and run
```bash
 git clone https://github.com/MixTapeSoftware/inctui.git \
 cd inctui \
 go build . \
 ./inctui
```

## Developing 

Run with live debugger to see bubbletea lifecyle events:

`export DEBUG=true; go run .`
