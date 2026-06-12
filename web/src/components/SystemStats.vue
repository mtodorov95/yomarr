<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface SystemStats {
  total_series: number;
  downloaded_chapters: number;
  missing_chapters: number;
  size_on_disk_bytes: number;
}

const stats = ref<SystemStats | null>(null)
const loading = ref(true)

function formatBytes(bytes: number, decimals: number = 1): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const dm = decimals < 0 ? 0 : decimals
  const sizes = ['B', 'KiB', 'MiB', 'GiB', 'TiB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i]
}

async function fetchStats() {
  try {
    const res = await fetch('/api/system/stats')
    if (res.ok) {
      stats.value = await res.json()
    }
  } catch (e) {
    console.error('Failed fetching footer metrics:', e)
  } finally {
    loading.value = false
  }
}

onMounted(fetchStats)
</script>

<template>
  <footer class="system-dashboard-footer">
    <div class="legend-panel">
      <div class="legend-item">
        <span class="indicator-box blue"></span>
        <span class="legend-text">Ongoing</span>
      </div>
      <div class="legend-item">
        <span class="indicator-box green"></span>
        <span class="legend-text">Completed</span>
      </div>
      <div class="legend-item">
        <span class="indicator-box purple"></span>
        <span class="legend-text">Downloading</span>
      </div>
      <div class="legend-item">
        <span class="indicator-box orange"></span>
        <span class="legend-text">Hiatus</span>
      </div>
      <div class="legend-item">
        <span class="indicator-box deep-red"></span>
        <span class="legend-text">Unmonitored</span>
      </div>
    </div>

    <div class="metrics-panel">
      <div class="metric-group">
        <div class="metric-row">
          <span class="label">Series:</span>
          <span class="value">{{ loading || !stats ? '-' : stats.total_series }}</span>
        </div>
      </div>

      <div class="metric-group">
        <div class="metric-row">
          <span class="label">Chapters:</span>
          {{ loading || !stats ? '-' : (stats.downloaded_chapters + stats.missing_chapters) }}
        </div>
        <div class="metric-row sub-row">
          <span class="label">Missing:</span>
          <span class="value text-warning">{{ loading || !stats ? '-' : stats.missing_chapters }}</span>
        </div>
      </div>

      <div class="metric-group total-size-group">
        <div class="metric-row">
          <span class="label">Total File Size:</span>
          <span class="valueHighlight">
            {{ loading || !stats ? '-' : formatBytes(stats.size_on_disk_bytes) }}
          </span>
        </div>
      </div>
    </div>
  </footer>
</template>

<style scoped>
.system-dashboard-footer {
  background-color: #0f172a;
  border-top: 1px solid #1e293b;
  padding: 1rem 2rem;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  font-family: system-ui, -apple-system, sans-serif;
  color: #94a3b8;
  font-size: 0.85rem;
  width: 100%;
  box-sizing: border-box;
}

.legend-panel {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.indicator-box {
  width: 14px;
  height: 14px;
  border-radius: 3px;
  display: inline-block;
}

.indicator-box.blue   { background-color: #2563eb; }
.indicator-box.green  { background-color: #16a34a; }
.indicator-box.purple { background-color: #9333ea; }
.indicator-box.orange   { background-color: #f97316; }
.indicator-box.deep-red { background-color: #7f1d1d; }

.legend-text {
  color: #94a3b8;
}

.metrics-panel {
  display: flex;
  gap: 3.5rem;
}

.metric-group {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.metric-row {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  min-width: 110px;
}

.metric-row.sub-row {
  font-size: 0.8rem;
  opacity: 0.8;
}

.label {
  color: #64748b;
  text-align: left;
}

.value {
  color: #e2e8f0;
  font-weight: 600;
  text-align: right;
}

.valueHighlight {
  color: #f8fafc;
  font-weight: 700;
}

.text-warning {
  color: #f59e0b;
}

@media (max-width: 768px) {
  .system-dashboard-footer {
    flex-direction: column;
    gap: 1.5rem;
    padding: 1rem;
  }
  .metrics-panel {
    gap: 1.5rem;
    flex-wrap: wrap;
  }
}
</style>
