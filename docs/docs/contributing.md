# Contributing

All contributions are welcome — code, documentation, testing, and bug reports. The repository lives at [CHE3MZ/jenkins-desktop](https://github.com/CHE3MZ/jenkins-desktop); please open an issue there before large changes so design can be agreed up front.

## Development setup

- Go (see `go.mod`), [Bun](https://bun.sh), [Wails v2](https://wails.io)
- `wails dev` for live development with hot reload
- `./scripts/build-macos.sh` for packaged builds
- Set `const debugSplash = true` in `app.go` to iterate on the splash screen without starting Jenkins (remember to set it back)

Every push to `master` runs the `build-macos` workflow, and successful builds are kept as artifacts — a green run is the minimum bar for a change.

## Conventions

- Go code must be `gofmt`-clean and pass `go vet` plus `golangci-lint` with zero issues.
- Keep the splash screen visually consistent with upstream Jenkins (design tokens and components are copied from `reference/`). Don't touch `docs/theme/` — the docs theme is fixed.
- Don't commit build output (`build/bin`, `frontend/dist`), `node_modules`, or the `reference/` upstream copy — all are git-ignored.

## Current needs

These are the areas where help matters most, roughly ordered by impact:

### 1. Proper Windows support

The backend compiles for Windows (`GOOS=windows go build` passes) and `JENKINS_COMMAND`/`JENKINS_ARGS` already allow pointing at a Windows launcher, but it has never been tested on a real Windows machine. Needed: verify startup/attach/shutdown against a local Jenkins (service install and `java -jar jenkins.war`), confirm WebView2 behavior for the localhost handoff, and figure out the installer story.

### 2. More Jenkins install sources

Startup currently assumes a `jenkins` executable on `PATH` (plus hardcoded Homebrew fallback dirs). It should detect more layouts automatically:

- **Linux:** distro packages (`apt`/`dnf` installs), the standalone `jenkins.war`
- **Windows:** Winget, Chocolatey, Scoop installs, Windows service detection
- **macOS:** MacPorts and manual installs beyond Homebrew
- Possibly: a Jenkins running in Docker (port probe already handles the attach half; discovery is the missing piece)

The natural shape is extending the `ensureGuiPath` idea per-OS: a list of known install locations probed before falling back to `PATH`.

### 3. Linux support

Same story as Windows: expected to mostly work (process-group code is shared Unix), but untested — needs dependency docs (WebKitGTK), a smoke test, and likely small fixes.

### 4. Code signing and notarization

Releases are currently ad-hoc signed, which means Gatekeeper warnings on other people's Macs. Needed: an Apple Developer ID workflow (certificates as CI secrets, `notarytool` stapling) wired into `release-macos`.

### 5. Tests

There are none yet. Valuable starting points: unit tests for the probe/state machine in `app.go` (port validation, attach-vs-spawn decisions with a fake HTTP server) and component tests for the splash screen states.

### 6. Update mechanism

Right now updating means downloading a new DMG from Releases. Options worth exploring: an in-app update check against the GitHub Releases API, or publishing via Homebrew Cask.
