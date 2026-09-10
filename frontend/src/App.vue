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
        <div class="spinner" aria-hidden="true"></div>
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
  max-width: 560px;
  width: 100%;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 12px;
  padding: 40px 36px;
}

h1 {
  margin: 0 0 4px;
  font-size: 28px;
  font-weight: 700;
}

.url {
  margin: 0 0 24px;
  color: #9fb3c8;
  font-size: 14px;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
}

.status p {
  margin: 16px 0 0;
  color: #d7e3f0;
}

.spinner {
  width: 36px;
  height: 36px;
  margin: 0 auto;
  border-radius: 50%;
  border: 3px solid rgba(255, 255, 255, 0.2);
  border-top-color: #fff;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.error-title {
  font-weight: 700;
  color: #ff9d9d;
}

.error p {
  color: #d7e3f0;
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
  cursor: pointer;
  background: #fff;
  color: #1b2636;
}

.btn.secondary {
  background: transparent;
  color: #fff;
  border: 1px solid rgba(255, 255, 255, 0.4);
}

.logs {
  text-align: left;
  background: rgba(0, 0, 0, 0.45);
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
