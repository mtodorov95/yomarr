<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { Series } from '../types'
import SeriesList from './SeriesList.vue'
import SeriesDetail from './SeriesDetail.vue'

const series = ref<Series[]>([])
const selectedSeries = ref<Series | null>(null)
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

async function removeSeries(id: number) {
    if (!confirm('Are you sure you want to remove this series?')) return

    try {
        const res = await fetch(`/api/series?id=${id}`, { method: 'DELETE' })
        if (res.ok) {
            series.value = series.value.filter(s => s.id !== id)
        } else {
            alert('Failed to remove series')
        }
    } catch (e) {
        console.error(e)
    }
}

onMounted(fetchSeries)
</script>

<template>
    <div class="w-full max-w-4xl">
        <template v-if="selectedSeries">
            <SeriesDetail :series="selectedSeries" @back="selectedSeries = null" />
        </template>

        <template v-else>
        <div class="flex justify-between items-center mb-6">
            <h2 class="text-2xl font-bold">Library</h2>
            <button @click="fetchSeries"
                class="text-sm bg-slate-800 px-3 py-1 rounded border border-slate-700">Refresh</button>
        </div>

        <div v-if="loading" class="text-slate-500">Loading library...</div>

        <template v-else>
            <SeriesList 
                v-if="series.length > 0" 
                :seriesList="series" 
                @delete="removeSeries" 
                @select="s => selectedSeries = s"
                />
            <div v-else class="text-slate-500 italic">Library empty. Search and import series.</div>
        </template>
        </template>
    </div>
</template>
