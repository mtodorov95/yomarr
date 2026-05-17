<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { Series } from '../types'

const series = ref<Series[]>([])
const loading = ref(true)

async function fetchSeries() {
  loading.value = true
  try {
    const res = await fetch('/api/series')
    if (!res.ok) throw new Error('fetch fail')
    series.value = await res.json()
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(fetchSeries)
</script>

<template>
  <div class="w-full max-w-4xl">
    <div class="flex justify-between items-center mb-6">
      <h2 class="text-2xl font-bold">Library</h2>
      <button @click="fetchSeries" class="text-sm bg-slate-800 px-3 py-1 rounded border border-slate-700">Refresh</button>
    </div>

    <div v-if="loading" class="text-slate-500">Loading library...</div>
    
    <div v-else-if="series.length > 0" class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div v-for="s in series" :key="s.id" class="p-4 bg-slate-800 rounded-xl border border-slate-700">
        <h3 class="font-bold text-blue-400">{{ s.title }}</h3>
        <p class="text-xs text-slate-400 mb-2">Status: {{ s.status }}</p>
        <code class="text-xs bg-black p-1 rounded block truncate">{{ s.path }}</code>
      </div>
    </div>

    <div v-else class="text-slate-500 italic">Library empty. Search and import series.</div>
  </div>
</template>
