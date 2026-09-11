# Getting Started

## Requirements

- macOS
- Jenkins installed via Homebrew (this also installs a compatible Java):

```sh
brew install jenkins
```

Verify it works from a terminal:

```sh
jenkins --version
```

## Install the app

Build it from source (see [Building from Source](building.md)):

```sh
./scripts/build-macos.sh
```

Then open `build/bin/Jenkins Desktop.app`.

## First launch

1. The app opens a splash screen and starts Jenkins. First boot takes a few seconds.
2. Jenkins asks to be unlocked. Find the initial admin password in the app via **View logs** on the splash screen, or read it directly:

```sh
cat ~/.jenkins/secrets/initialAdminPassword
```

3. Complete the Jenkins setup wizard in the window. From then on the app opens straight into Jenkins.

## Daily use

Just open the app. If Jenkins is already running (started by you or by a previous session that was killed), the app attaches to it instead of launching a duplicate.

### Long builds

Closing the app stops the Jenkins it started. If you run builds that outlast your desktop session, run Jenkins independently so the app only ever attaches:

```sh
brew services start jenkins
```

With Jenkins managed by `brew services`, it keeps running in the background whether the app is open or not, and closing the window can never interrupt a build. (Running `jenkins` directly in a terminal works too, but the terminal tab must stay open.)
