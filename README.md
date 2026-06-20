# IncTUI: An Incus TUI for Developers

IncTUI provies a TUI on top of the local Incus API (via unix socket).

## Incus (Linux containers) on MacOS

I recommend installing Incus with `Colima`:

```bash
brew install incus 
brew install colima 
colima start --runtime=incus
```
### Set `INCUS_SOCKET` path

Put this in your env config:
`INCUS_SOCKET=~/.colima/default/incus.sock`

If `inctui` fails to start, check `~/.config/incus/config.yml` and replace the above with the path after `unix://`

## Installation 

### Clone, build and run
```bash
 git clone https://github.com/MixTapeSoftware/inctui.git \
 cd inctui \
 go build . \
 ./inctui
```
