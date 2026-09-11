<script lang="ts" setup>
import { onMounted, onUnmounted, ref } from 'vue'
import { JenkinsLogs, JenkinsState } from '../wailsjs/go/main/App'

interface JenkinsStatus {
  status: 'starting' | 'ready' | 'error'
  message: string
  url: string
  pid: number
  owned: boolean
}

const state = ref<JenkinsStatus>({
  status: 'starting',
  message: 'Starting Jenkins...',
  url: 'http://localhost:8080',
  pid: 0,
  owned: true,
})
const logs = ref('')
const showLogs = ref(false)
const redirected = ref(false)

let timer: number | undefined
let redirectTimer: number | undefined

async function refreshLogs() {
  try {
    logs.value = await JenkinsLogs()
  } catch {
    // backend not reachable yet, keep old logs
  }
}

async function poll() {
  try {
    state.value = (await JenkinsState()) as JenkinsStatus
  } catch {
    // backend not reachable yet, keep polling
    return
  }

  if (state.value.status === 'ready' && !redirected.value) {
    redirected.value = true
    if (timer !== undefined) window.clearInterval(timer)
    // Hand the whole window over to the Jenkins web UI. Jenkins sends
    // X-Frame-Options: sameorigin, so it cannot live inside an iframe —
    // a top-level navigation displays it natively instead.
    redirectTimer = window.setTimeout(async () => {
      // Re-check right before navigating: Jenkins may have died since
      // the last poll, and a blind redirect would land on an error page.
      try {
        const fresh = (await JenkinsState()) as JenkinsStatus
        if (fresh.status === 'ready') {
          window.location.replace(fresh.url)
          return
        }
      } catch {
        // backend unreachable, fall through and resume polling
      }
      redirected.value = false
      poll()
      timer = window.setInterval(poll, 750)
    }, 500)
    return
  }

  if (showLogs.value) {
    await refreshLogs()
  }
}

function retry() {
  redirected.value = false
  poll()
}

async function toggleLogs() {
  showLogs.value = !showLogs.value
  if (showLogs.value) await refreshLogs()
}

onMounted(() => {
  poll()
  timer = window.setInterval(poll, 750)
})

onUnmounted(() => {
  if (timer !== undefined) window.clearInterval(timer)
  if (redirectTimer !== undefined) window.clearTimeout(redirectTimer)
})
</script>

<template>
  <main class="simple-page">
    <div class="app-jenkins-booting">
      <div class="logo">
        <img src="./assets/images/jenkins-logo.svg" alt="Jenkins logo" />
      </div>
      <div class="brand">Jenkins Desktop</div>

      <template v-if="state.status !== 'error'">
        <h1 class="loading">
          <p class="jenkins-spinner">{{ redirected ? 'Opening Jenkins...' : state.message }}</p>
        </h1>
        <div
          class="app-progress-bar app-progress-bar--unknown app-progress-bar--animate"
          aria-hidden="true"
        >
          <span></span><span></span>
        </div>
        <div class="restarting">{{ state.url }}</div>
        <button class="jenkins-button logs-toggle" @click="toggleLogs">
          {{ showLogs ? 'Hide logs' : 'View logs' }}
        </button>
        <pre v-if="showLogs" class="logs">{{ logs || '(no output yet)' }}</pre>
      </template>

      <template v-else>
        <h1 class="loading">
          <p class="error-title">Jenkins failed to start</p>
        </h1>
        <div class="restarting">{{ state.message }}</div>
        <div class="actions">
          <button class="jenkins-button jenkins-button--primary" @click="retry">Retry</button>
          <button class="jenkins-button" @click="toggleLogs">
            {{ showLogs ? 'Hide logs' : 'Show logs' }}
          </button>
        </div>
        <pre v-if="showLogs" class="logs">{{ logs || '(no output yet)' }}</pre>
      </template>
    </div>
  </main>
</template>

<style scoped>
/* Jenkins design tokens (light theme), copied from the Jenkins repo:
   reference/src/main/scss/abstracts/_theme.scss */
.app-jenkins-booting {
  --background: #fefefe;
  --text-color: #010010;
  --text-color-secondary: #6382ad;
  --accent-color: #0067f3;
  --error-color: #f80000;
  --font-family-sans:
    system-ui, "Segoe UI", roboto, "Noto Sans", oxygen, ubuntu, cantarell,
    "Fira Sans", "Droid Sans", "Helvetica Neue", arial, sans-serif;
  --font-bold-weight: 450;
  --form-input-border-radius: 0.625rem;
}

.simple-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: var(--background);
  font-family: var(--font-family-sans);
}

.simple-page .logo {
  text-align: center;
}

.simple-page .logo > img {
  height: 140px;
}

.brand {
  font-family: Georgia, 'Times New Roman', serif;
  font-weight: 700;
  font-size: 1.75rem;
  line-height: 1.2;
  color: var(--text-color);
  text-align: center;
  margin: 0;
}

/* Mirrors .app-jenkins-booting in reference/src/main/scss/simple-page.scss */
.app-jenkins-booting {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  align-items: center;
  max-width: 600px;
  width: 100%;
  padding: 0 24px;
  box-sizing: border-box;
  animation: fade-in-jenkins-booting 0.4s both 0.2s;
  /* Optical compensation to visually center content */
  margin-top: -3rem;
}

@keyframes fade-in-jenkins-booting {
  from {
    scale: 97.5%;
    opacity: 0;
    filter: blur(1px);
  }
}

.loading {
  margin: 1rem 0 0;
  font-size: 24px;
  line-height: normal;
  text-align: center;
  color: var(--text-color);
}

.loading p {
  font-weight: var(--font-bold-weight);
  font-size: 1.125rem;
}

/* Mirrors .jenkins-spinner in
   reference/src/main/scss/components/_spinner.scss */
.jenkins-spinner {
  position: relative;
  display: inline-flex;
  align-items: center;
  margin: 0;
}

.jenkins-spinner::before,
.jenkins-spinner::after {
  content: "";
  display: inline-block;
  width: 1em;
  height: 1em;
  border-radius: 100%;
  border: 0.14em solid currentColor;
}

.jenkins-spinner::before {
  position: relative;
  margin-right: 0.55em;
  opacity: 0.3;
  border-color: var(--text-color-secondary);
}

.jenkins-spinner::after {
  position: absolute;
  top: 50%;
  left: 0;
  translate: 0 -50%;
  clip-path: inset(0 0 50% 50%);
  animation: loading-spinner 1s infinite linear;
}

@keyframes loading-spinner {
  to {
    transform: rotate(360deg);
  }
}

/* Mirrors .app-progress-bar (+ --unknown, --animate) in
   reference/src/main/scss/components/_progress-bar.scss */
.app-progress-bar {
  --color: var(--accent-color);
  height: 12px;
  width: 50%;
  padding: 2px;
  border-radius: 6px;
  box-shadow: 0px 0px 2px rgba(200,200,200,0.65);
  box-sizing: border-box;
  /* Jenkins uses --text-color-secondary at 25% here; bumped to 45% so the
     full-width track stays visible against the page background. */
  background-color: rgba(99, 130, 173, 0.4);
  background-image: linear-gradient(
    -45deg,
    var(--background) 25%,
    transparent 25%,
    transparent 50%,
    var(--background) 50%,
    var(--background) 75%,
    transparent 75%,
    transparent
  );
  background-size: 25px 25px;
  animation: progress-bar-stripes 4s linear infinite;
  overflow: hidden;
  position: relative;
}

.app-progress-bar > span {
  position: absolute;
  top: 2px;
  bottom: 2px;
  left: 0;
  width: 30%;
  background-color: var(--color);
  box-shadow: 1px 1px 3px rgba(0,0,0,0.05);
  display: block;
  border-radius: 4px;
  animation: progress-bar-slide 1.4s ease-in-out infinite;
}

/* Second segment chases the first so the bar is never empty mid-cycle */
.app-progress-bar > span + span {
  animation-delay: -0.7s;
}

@keyframes progress-bar-stripes {
  0% {
    background-position: 0 0;
  }

  100% {
    background-position: 25px 25px;
  }
}

@keyframes progress-bar-slide {
  from {
    transform: translateX(-110%);
  }

  to {
    transform: translateX(350%);
  }
}

.restarting {
  color: var(--text-color-secondary);
  text-align: center;
  font-size: 0.875rem;
  margin: 0;
}

.error-title {
  color: var(--error-color);
}

.actions {
  display: flex;
  gap: 0.75rem;
  justify-content: center;
}

/* Mirrors .jenkins-button (+ --primary) in
   reference/src/main/scss/components/_buttons.scss */
.jenkins-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin: 0;
  padding: 0.5rem 1rem;
  min-height: 2.375rem;
  box-sizing: border-box;
  font-family: var(--font-family-sans);
  font-size: 0.875rem;
  white-space: nowrap;
  cursor: pointer;
  color: var(--text-color);
  background: rgba(99, 130, 173, 0.12);
  border: 1px solid rgba(99, 130, 173, 0.35);
  border-radius: var(--form-input-border-radius);
}

.jenkins-button--primary {
  background: var(--accent-color);
  border-color: var(--accent-color);
  color: var(--background);
}

.logs {
  text-align: left;
  width: 100%;
  box-sizing: border-box;
  background: rgba(99, 130, 173, 0.08);
  color: var(--text-color);
  border: 1px solid rgba(99, 130, 173, 0.35);
  border-radius: var(--form-input-border-radius);
  padding: 12px;
  margin: 0;
  max-height: 220px;
  overflow: auto;
  font-size: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
}

.logs-toggle {
  min-height: 0;
  padding: 0.25rem 0.75rem;
  font-size: 0.75rem;
}
</style>
