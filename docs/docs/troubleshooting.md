# Troubleshooting

## "Could not find the jenkins command"

The app can't resolve `JENKINS_COMMAND` on its `PATH`. If `jenkins` works in your terminal, install location is the issue: terminal sessions load `~/.zshrc`, Finder-launched apps don't. The app already adds the Homebrew locations itself, so this usually means Jenkins isn't installed:

```sh
brew install jenkins
```

Otherwise set `JENKINS_COMMAND` to the full path of your launcher. The error message shows the exact `PATH` that was searched.

## "Port is already in use by another service"

Something non-Jenkins is listening on the port. Stop it, or point the app elsewhere:

```sh
JENKINS_PORT=9090 open "build/bin/Jenkins Desktop.app"
```

If that "something" is actually Jenkins (started by you or a previous session), the app attaches to it instead — no action needed.

## "Jenkins exited unexpectedly"

The child process died. Open **View logs** on the error screen — the tail of Jenkins's output is there. Typical causes: an incompatible Java version or a broken Jenkins home directory. The same output goes to the app's stdout when launched from a terminal.

## Splash screen never finishes

First boots (plugin installs, slow disks) can take minutes — the splash shows elapsed seconds while Jenkins answers 503. Use **View logs** to watch progress. If Jenkins is genuinely stuck, quit the app (which stops its child) and try again.

## Leftover Jenkins after a crash

Normal quits always stop the owned Jenkins. Only an uncatchable kill (`kill -9`) or a crash can orphan it; the next launch will attach to the leftover rather than duplicate it. To clean up manually:

```sh
pkill -f jenkins.war
```

## Window shows a browser connection error

Jenkins died between the readiness check and the handoff. Relaunch the app; if it repeats, check the logs as above.
