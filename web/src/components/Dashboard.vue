<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { Series } from '../types'
import SeriesList from './SeriesList.vue'
import SeriesDetail from './SeriesDetail.vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const series = ref<Series[]>([])
const loading = ref(true)

async function fetchSeries() {
    loading.value = true
    try {
        const res = await fetch('/api/series')
        if (!res.ok) throw new Error('fetch fail')
        const data = await res.json() ?? [];
        series.value = data
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

function handleSelect(s: Series) {
    router.push({ name: "series-detail", params: { id: s.id } })
}

onMounted(fetchSeries)
</script>

<template>
    <div class="dashboard-wrapper">
        <div class="dashboard-header">
            <h2 class="header-title">Library</h2>
            <button @click="fetchSeries" class="refresh-button">Refresh</button>
        </div>

        <div v-if="loading" class="info-message">Loading library...</div>

        <template v-else>
            <SeriesList v-if="series.length > 0" :seriesList="series" @delete="removeSeries" @select="handleSelect" />
            <div v-else class="info-message status-empty">
                Library empty. Search and import series.
            </div>
        </template>
    </div>
</template>

<style scoped>
.dashboard-wrapper {
    width: 100%;
    max-width: 56rem;
}

.dashboard-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
}

.header-title {
    font-size: 1.5rem;
    font-weight: 700;
    margin: 0;
}

.refresh-button {
    font-size: 0.875rem;
    background-color: #1e293b;
    color: #ffffff;
    padding: 0.25rem 0.75rem;
    border-radius: 0.25rem;
    border: 1px solid #334155;
    cursor: pointer;
    transition: background-color 0.2s, border-color 0.2s;
}

.refresh-button:hover {
    background-color: #334155;
    border-color: #475569;
}

.info-message {
    color: #64748b;
}

.status-empty {
    font-style: italic;
}
</style>
