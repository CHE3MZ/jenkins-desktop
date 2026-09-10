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
    setTimeout(() => window.location.replace(state.value.url), 500)
    return
  }

  if (state.value.status === 'error' && showLogs.value) {
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
})
</script>

<template>
  <main class="splash">
    <div class="card">
      <h1>Jenkins Desktop</h1>
      <p class="url">{{ state.url }}</p>

      <div v-if="state.status !== 'error'" class="status">
        <div class="hbar" aria-hidden="true"><div class="hbar-fill"></div></div>
        <p>{{ redirected ? 'Opening Jenkins...' : state.message }}</p>
      </div>

      <div v-else class="error">
        <p class="error-title">Jenkins failed to start</p>
        <p>{{ state.message }}</p>
        <div class="actions">
          <button class="btn" @click="retry">Retry</button>
          <button class="btn secondary" @click="toggleLogs">
            {{ showLogs ? 'Hide logs' : 'Show logs' }}
          </button>
        </div>
        <pre v-if="showLogs" class="logs">{{ logs || '(no output yet)' }}</pre>
      </div>
    </div>
  </main>
</template>

<style scoped>
.splash {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  box-sizing: border-box;
}

.card {
  max-width: 400px;
  width: 100%;
  background: #ffffff;
  border: 1px solid #e3e6ea;
  border-radius: 12px;
  box-shadow: 1px 2px 3px rgba(15, 30, 45, 0.2);
  padding: 36px 32px;
}

h1 {
  margin: 0 0 4px;
  font-size: 30px;
  font-weight: 700;
  font-family: Georgia, 'Times New Roman', serif;
  color: #11181f;
}

.url {
  margin: 0 0 24px;
  color: #5b6b7c;
  font-size: 14px;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  font-weight: 700;
}

.status p {
  margin: 20px 0 0;
  color: #333e4c;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  font-weight: 700;
}

.hbar {
  height: 18px;
  margin: 4px auto 0;
  background: #e9edf2;
  border: 3px solid #ffffff;
  border-radius: 6px;
  box-shadow: 1px 2px 3px rgba(15, 30, 45, 0.2);
  overflow: hidden;
  position: relative;
}

.hbar-fill {
  position: absolute;
  top: 0;
  left: 0;
  height: 100%;
  width: 32%;
  background: #2563eb;
  border-radius: 3px;
  animation: hslide 1.2s ease-in-out infinite;
}

@keyframes hslide {
  from {
    transform: translateX(-110%);
  }
  to {
    transform: translateX(320%);
  }
}

.error-title {
  font-weight: 700;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  color: #b3261e;
}

.error p {
  color: #333e4c;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  font-weight: 700;
}

.actions {
  display: flex;
  gap: 12px;
  justify-content: center;
  margin: 20px 0;
}

.btn {
  border: none;
  border-radius: 6px;
  padding: 8px 20px;
  font-size: 14px;
  font-weight: 700;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  cursor: pointer;
  background: #11181f;
  color: #fff;
}

.btn.secondary {
  background: #fff;
  color: #11181f;
  border: 1px solid #11181f;
}

.logs {
  text-align: left;
  background: #f1f3f5;
  color: #333e4c;
  border: 1px solid #e3e6ea;
  border-radius: 8px;
  padding: 12px;
  max-height: 220px;
  overflow: auto;
  font-size: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
