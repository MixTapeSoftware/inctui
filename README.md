# IncTUI: An Incus TUI for Developers

IncTUI provies a TUI on top of the local Incus API (via unix socket).

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
