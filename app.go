package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Jenkins lifecycle states reported to the frontend.
const (
	JenkinsStarting = "starting"
	JenkinsReady    = "ready"
	JenkinsError    = "error"
)

const maxLogLines = 200

// debugSplash locks the app on the splash screen for UI debugging.
// While true, Jenkins is never started and the frontend never redirects.
// Set to false for normal behavior.
const debugSplash = true

// App owns the Jenkins child process for the lifetime of the desktop app.
type App struct {
	ctx context.Context

	httpClient *http.Client

	mu      sync.Mutex
	status  string
	message string
	url     string
	pid     int
	owned   bool // false when we attached to an already-running Jenkins

	cmd    *exec.Cmd
	doneCh chan struct{} // closed when the child process exits
	stopCh chan struct{} // closed on app shutdown
	stopWg sync.WaitGroup

	logs []string
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{
		status:  JenkinsStarting,
		message: "Starting Jenkins...",
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

// startup is called when the app starts. It launches Jenkins and owns it.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.stopCh = make(chan struct{})

	if debugSplash {
		log.Print("debugSplash is on: staying on the splash screen, Jenkins will not start.")
		a.setState(JenkinsStarting, "Splash preview — set debugSplash to false for normal startup.", 0)
		return
	}

	command := envOr("JENKINS_COMMAND", "jenkins")
	port := envOr("JENKINS_PORT", "8080")
	a.url = "http://localhost:" + port

	args := []string{"--httpPort=" + port}
	if extra := os.Getenv("JENKINS_ARGS"); extra != "" {
		args = append(args, strings.Fields(extra)...)
	}

	// If something Jenkins-shaped is already listening, attach to it
	// instead of spawning a second instance.
	if probed := a.probe(); probed.isJenkins {
		a.mu.Lock()
		a.owned = false
		a.mu.Unlock()
		log.Printf("Jenkins already running at %s, attaching (will not kill on exit)", a.url)
		if probed.ready {
			a.setState(JenkinsReady, "Jenkins is already running.", 0)
		} else {
			a.setState(JenkinsStarting, fmt.Sprintf("Jenkins is starting (HTTP %d)...", probed.code), 0)
		}
		a.watchReadiness()
		return
	} else if probed.listening {
		a.setState(JenkinsError, fmt.Sprintf("Port %s is already in use by another service. Stop it or set JENKINS_PORT to a free port.", port), 0)
		return
	}

	path, err := exec.LookPath(command)
	if err != nil {
		a.setState(JenkinsError, fmt.Sprintf("Could not find the %q command. Install Jenkins and make sure it is on your PATH.", command), 0)
		return
	}

	cmd := exec.Command(path, args...)
	setProcessGroup(cmd) // so we can signal the whole group on shutdown
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		a.setState(JenkinsError, fmt.Sprintf("Failed to start Jenkins: %v", err), 0)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		a.setState(JenkinsError, fmt.Sprintf("Failed to start Jenkins: %v", err), 0)
		return
	}

	if err := cmd.Start(); err != nil {
		a.setState(JenkinsError, fmt.Sprintf("Failed to start Jenkins: %v", err), 0)
		return
	}

	a.mu.Lock()
	a.cmd = cmd
	a.pid = cmd.Process.Pid
	a.owned = true
	a.doneCh = make(chan struct{})
	a.mu.Unlock()

	log.Printf("Jenkins started (pid %d): %s %s", cmd.Process.Pid, path, strings.Join(args, " "))
	a.setState(JenkinsStarting, "Jenkins is starting, this usually takes a few seconds...", cmd.Process.Pid)

	// Stream child output into the in-memory log (and our own stdout).
	a.stopWg.Add(1)
	go func() {
		defer a.stopWg.Done()
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		for scanner.Scan() {
			a.appendLog(scanner.Text())
		}
	}()
	a.stopWg.Add(1)
	go func() {
		defer a.stopWg.Done()
		scanner := bufio.NewScanner(stderr)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		for scanner.Scan() {
			a.appendLog(scanner.Text())
		}
	}()

	// Notice unexpected exits (e.g. bad Java version, port conflict).
	go func() {
		err := cmd.Wait()
		a.mu.Lock()
		owned := a.owned
		status := a.status
		a.mu.Unlock()
		close(a.doneCh)
		select {
		case <-a.stopCh:
			return // we are shutting down, exit was expected
		default:
		}
		if owned && status != JenkinsError {
			msg := "Jenkins exited unexpectedly."
			if err != nil {
				msg = fmt.Sprintf("Jenkins exited unexpectedly: %v", err)
			}
			a.setState(JenkinsError, msg+" See the logs below.", cmd.Process.Pid)
		}
	}()

	a.watchReadiness()
}

// shutdown is called when the app is closing. It stops the Jenkins process
// we own so nothing is left behind. Attached (foreign) instances are left alone.
func (a *App) shutdown(ctx context.Context) {
	select {
	case <-a.stopCh:
		return
	default:
		close(a.stopCh)
	}

	a.mu.Lock()
	cmd := a.cmd
	owned := a.owned
	a.mu.Unlock()

	if !owned || cmd == nil || cmd.Process == nil {
		return
	}

	pid := cmd.Process.Pid
	log.Printf("Stopping Jenkins (pid %d)...", pid)
	if err := signalProcessGroup(pid); err != nil {
		log.Printf("Failed to signal Jenkins process: %v", err)
	}

	select {
	case <-a.doneCh:
		log.Print("Jenkins stopped cleanly.")
	case <-time.After(15 * time.Second):
		log.Print("Jenkins did not stop in time, killing...")
		_ = killProcessGroup(pid)
		select {
		case <-a.doneCh:
		case <-time.After(5 * time.Second):
			log.Print("Jenkins process did not die, giving up.")
		}
	}
	a.stopWg.Wait()
}

// JenkinsURL returns the URL of the Jenkins web UI.
func (a *App) JenkinsURL() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.url
}

// JenkinsState is polled by the splash screen. It reports the Jenkins
// lifecycle state: starting, ready or error.
func (a *App) JenkinsState() map[string]interface{} {
	a.mu.Lock()
	defer a.mu.Unlock()
	return map[string]interface{}{
		"status":  a.status,
		"message": a.message,
		"url":     a.url,
		"pid":     a.pid,
		"owned":   a.owned,
	}
}

// JenkinsLogs returns the tail of the Jenkins process output.
func (a *App) JenkinsLogs() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return strings.Join(a.logs, "\n")
}

func (a *App) setState(status, message string, pid int) {
	a.mu.Lock()
	a.status = status
	a.message = message
	if pid != 0 {
		a.pid = pid
	}
	a.mu.Unlock()
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "jenkins:state", status)
		if status == JenkinsReady {
			runtime.EventsEmit(a.ctx, "jenkins:ready", a.JenkinsURL())
		}
	}
}

func (a *App) appendLog(line string) {
	log.Printf("[jenkins] %s", line)
	a.mu.Lock()
	a.logs = append(a.logs, line)
	if len(a.logs) > maxLogLines {
		a.logs = a.logs[len(a.logs)-maxLogLines:]
	}
	a.mu.Unlock()
}

type probeResult struct {
	listening bool
	isJenkins bool
	ready     bool
	code      int
}

// probe checks whether our Jenkins URL is already serving something and
// whether that something looks like Jenkins (/login answers 200 when ready,
// 503 while it is still booting).
func (a *App) probe() probeResult {
	var res probeResult
	resp, err := a.httpClient.Get(a.url + "/login")
	if err != nil {
		return res
	}
	defer resp.Body.Close()
	res.listening = true
	res.code = resp.StatusCode
	if resp.Header.Get("X-Jenkins") != "" || resp.Header.Get("X-Hudson") != "" {
		res.isJenkins = true
		res.ready = resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusForbidden
	}
	return res
}

// watchReadiness polls /login until Jenkins answers 200, then flips to ready.
func (a *App) watchReadiness() {
	a.stopWg.Add(1)
	go func() {
		defer a.stopWg.Done()
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		start := time.Now()
		for {
			select {
			case <-a.stopCh:
				return
			case <-ticker.C:
				probed := a.probe()
				if probed.isJenkins && probed.ready {
					a.mu.Lock()
					pid := a.pid
					a.mu.Unlock()
					a.setState(JenkinsReady, "Jenkins is up and running.", pid)
					return
				}
				if probed.isJenkins {
					a.setState(JenkinsStarting, fmt.Sprintf("Jenkins is starting... (%ds)", int(time.Since(start).Seconds())), 0)
					continue
				}
				// If we spawned Jenkins ourselves, a missing port just means
				// it is not listening yet. If we attached, losing the port
				// means it went away.
				a.mu.Lock()
				owned := a.owned
				a.mu.Unlock()
				if !owned {
					a.setState(JenkinsError, "The Jenkins instance we attached to stopped responding.", 0)
					return
				}
			}
		}
	}()
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
