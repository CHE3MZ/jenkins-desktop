# Configuration

Everything is configured through environment variables. Defaults target a standard Homebrew Jenkins on macOS.

| Variable | Default | Meaning |
|---|---|---|
| `JENKINS_COMMAND` | `jenkins` | Binary to run (or a full path) |
| `JENKINS_PORT` | `8080` | HTTP port, passed as `--httpPort` |
| `JENKINS_LISTEN_ADDRESS` | `127.0.0.1` | Interface Jenkins binds to |
| `JENKINS_ARGS` | _(empty)_ | Extra args appended (split on whitespace) |

Examples:

```sh
# Run Jenkins on another port
JENKINS_PORT=9090 open "build/bin/Jenkins Desktop.app"

# Allow LAN access (default is this machine only)
JENKINS_LISTEN_ADDRESS=0.0.0.0 open "build/bin/Jenkins Desktop.app"
```

`JENKINS_PORT` must be a number between 1 and 65535, otherwise the app shows an error instead of starting anything.

## Startup rules

On launch the app probes `http://localhost:<port>/login`:

- **Nothing listening** → it spawns `jenkins --httpPort=<port> --httpListenAddress=<listen>` and owns that process.
- **Jenkins answering** (detected via its `X-Jenkins`/`X-Hudson` headers, not just an open port) → it attaches. The splash waits for it to finish booting (HTTP 503 means "still starting") and the app will not stop it on exit.
- **Something else answering** → error screen suggesting you free the port or set `JENKINS_PORT`.
- **Our child dies but a Jenkins is answering** (e.g. two app instances raced at startup) → the app attaches to the survivor instead of failing.

## Windows notes

The backend compiles for Windows, but there is usually no `jenkins` command on `PATH` there. Point the app at your launcher explicitly:

```bat
set JENKINS_COMMAND=java
set JENKINS_ARGS=-jar C:\path\to\jenkins.war
```

A WebView2 runtime is also required on Windows (standard for all Wails apps).
