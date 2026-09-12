## What does this do?

<!-- A short, friendly description of the change and why it's needed. Link any related issues like "Fixes #12". -->

## Type of change

<!-- Tick the ones that apply. -->

- [ ] Bug fix
- [ ] New feature
- [ ] UI change (please attach a screenshot or screen recording)
- [ ] Docs / website
- [ ] Build, CI, or packaging
- [ ] Something else: ___

## How was it tested?

<!-- The `build-macos` workflow runs automatically, but it only checks that things compile. -->

- [ ] `wails build` succeeds (or CI is green)
- [ ] Launched the app: Jenkins starts and the window hands over to it
- [ ] Quit the app: the owned Jenkins process stops, no orphans (`pgrep -f jenkins.war` is empty)
- [ ] Attach case still works (launch with Jenkins already running; quitting leaves it alone)
- [ ] `gofmt`, `go vet`, and `golangci-lint` are clean (Go changes)

## Checklist

- [ ] `debugSplash` is `false` in `app.go` (it must never ship as `true`)
- [ ] No new `<p>` elements wrapping dynamic text in the splash screen (they silently fail to render — use `<div>`/`<span>`)
- [ ] Docs updated if behavior changed
- [ ] No build output, credentials, or local-only files included
