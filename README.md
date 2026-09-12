<img src="assets/jenkins_logo.png" alt="Jenkins logo" width="192" />

# Jenkins Desktop

### __A simple desktop wrapper for Jenkins.__

Jenkins Desktop is a simple desktop wrapper for Jenkins that is currently compatible with Homebrew installations of Jenkins on macOS.

#### __Check out the [official documentation](https://che3mz.github.io/jenkins-desktop/) to get started.__ 

<img src="docs/docs/images/screenshot.png" alt="screenshot" width="900" />

## What it does

The application upon launch will attempt to run the "jenkins" command and open up the localhost URL of Jenkins and display it to the user. If Jenkins is already running, it will skip the first part (trying to run the Jenkins command).

It is heavily recommended that you already have Jenkins running in the background if you're planning to use this application for a long duration, as closing the application will also kill the Jenkins process IF Jenkins Desktop was the one that launched it, however, if you were the one to have launched it, then Jenkins will continue running and only the desktop app will close. This is crucial if you're working on long builds, etc., as closing Jenkins prematurely could cause issues.

brew can already handle this easily; just run:

```
brew services start jenkins
```

in your terminal and you won't have to worry about Jenkins (probably), as it will keep on running in the background and will be managed by the brew services service.
you could also run the "jenkins" command directly in your terminal, but that would mean that you'd have to keep the terminal tab open in order to keep Jenkins running, as closing the tab would also kill the Jenkins process.

## Contributing 

Check out the **[Contribution guide in the docs](docs/docs/contributing.md)** *or* **[In the website](https://che3mz.github.io/jenkins-desktop/contributing/)**.

**Any and all contributions are welcome! 😁**

## License

The __Jenkins Desktop__ project itself is licensed under the __[MIT license](LICENSE)__

See the licensing for the __Jenkins Project__ __[Here](https://github.com/jenkinsci/jenkins/blob/master/LICENSE.txt)__.
