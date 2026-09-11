# How It Works

## Backend lifecycle (`app.go`)

1. **Locate Jenkins.** The app first ensures the Homebrew binary locations are on its own `PATH` (Finder-launched apps don't inherit the shell `PATH`), then resolves `JENKINS_COMMAND` via `exec.LookPath`.
2. **Probe the port.** `GET http://localhost:<port>/login` decides between spawning a new Jenkins, attaching to a running one, or showing an error — see [Configuration](configuration.md#startup-rules).
3. **Own the process.** The child runs in its own process group so shutdown signals reach the whole tree. Stdout/stderr stream into a 500-line in-memory ring buffer exposed to the UI.
4. **Wait for readiness.** Jenkins answers HTTP 503 while booting and 200 once ready. The backend polls every 500ms and reports `starting` / `ready` / `error` states, also emitting `jenkins:state` / `jenkins:ready` events.
5. **Shut down.** On app close, Wails calls `OnShutdown`: SIGTERM to the process group, up to 15s grace for Jenkins's own orderly shutdown, then SIGKILL. Attached (foreign) instances are never signalled.

### Finding Jenkins

`wails dev` inherits the terminal `PATH`, but a `.app` opened from Finder gets the system default (`/usr/bin:/bin:/usr/sbin:/sbin`), where Homebrew is invisible. That is why the same code can work in dev and report "command not found" once packaged. The fix is local to the process: `ensureGuiPath()` prepends `/opt/homebrew/bin` and `/usr/local/bin` when they exist. It modifies nothing on the system — process environment only.

## Frontend handoff (`frontend/src/App.vue`)

The UI is a splash screen that polls `JenkinsState()` and, once ready, hands the entire window to Jenkins with `window.location.replace(url)`.

An `<iframe>` embed is not possible: Jenkins sends `X-Frame-Options: sameorigin`, which blocks framing from the Wails frontend origin. A top-level navigation has no such restriction and displays Jenkins natively.

The splash styling mirrors Jenkins's own boot page (`HudsonIsLoading`) and components (`.jenkins-spinner`, `.app-progress-bar`) from the upstream repo, using the same light-theme tokens.
