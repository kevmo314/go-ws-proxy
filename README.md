# kevmo314/go-ws-proxy

A tiny websocket proxy to redirect stdin/stdout to a websocket.

Useful for environments where a websocket library is not readily available.

## Usage

```sh
./go-ws-proxy <destination> [--origin origin] [--protocol protocol]
```
