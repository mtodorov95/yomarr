<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { Series } from '../types'
import SeriesList from './SeriesList.vue'
import { useRouter } from 'vue-router'
import { useToast } from '@/composables/useToast'
import ConfirmationModal from './ConfirmationModal.vue'

const router = useRouter()
const toast = useToast()

const series = ref<Series[]>([])
const loading = ref(true)
const scanning = ref(false)

const isModalOpen = ref(false)
const seriesToDelete = ref<number | null>(null)

async function fetchSeries() {
    loading.value = true
    try {
        const res = await fetch('/api/series')
        if (!res.ok) throw new Error('fetch fail')
        const data = await res.json() ?? [];
        series.value = data
    } catch (e) {
        console.error(e)
        toast.error("Failed fetching series")
    } finally {
        loading.value = false
    }
}

function triggerDeleteConfirmation(id: number) {
    seriesToDelete.value = id
    isModalOpen.value = true
}

async function removeSeries() {
    if (seriesToDelete.value === null) return

    const id = seriesToDelete.value
    isModalOpen.value = false
    seriesToDelete.value = null

    try {
        const res = await fetch(`/api/series?id=${id}`, { method: 'DELETE' })
        if (res.ok) {
            series.value = series.value.filter(s => s.id !== id)
            toast.success("Series removed")
        } else {
            toast.error('Failed to remove series')
        }
    } catch (e) {
        console.error(e)
        toast.error('Failed to remove series')
    }
}

function closeModal() {
    isModalOpen.value = false
    seriesToDelete.value = null
}

function handleSelect(s: Series) {
    router.push({ name: "series-detail", params: { id: s.id } })
}

async function runLibraryScan() {
    scanning.value = true
    try {
        const res = await fetch("/api/library/scan", {method: "POST"})
        if (res.ok) {
            toast.success("Manual library scan initiated in background")
            await fetchSeries()
        } else {
            toast.error("Failed to start library scan")
        }
    } catch (e) {
        console.error(e)
        toast.error("Network error trying to start library scan")
    } finally {
        scanning.value = false;
    }
}

onMounted(fetchSeries)
</script>

<template>
    <div class="dashboard-wrapper">
        <div class="dashboard-header">
            <h2 class="header-title">Library</h2>
            <div class="header-actions">
                <button 
                    @click="runLibraryScan" 
                    :disabled="scanning || loading" 
                    class="scan-button"
                >
                    {{ scanning ? 'Scanning...' : 'Scan Library' }}
                </button>
                <button 
                    @click="fetchSeries" 
                    :disabled="loading" 
                    class="refresh-button"
                >
                    Refresh
                </button>
            </div>
        </div>

        <div v-if="loading" class="info-message">Loading library...</div>

        <template v-else>
            <SeriesList 
            v-if="series.length > 0" 
            :seriesList="series" 
            @delete="triggerDeleteConfirmation" 
            @select="handleSelect" />
            <div v-else class="info-message status-empty">
                Library empty. Search and import series.
            </div>
        </template>

        <ConfirmationModal 
            :isOpen="isModalOpen"
            title="Remove Series"
            message="Are you sure you want to remove this series from the database? This action will drop tracker mappings."
            @close="closeModal"
            @confirm="removeSeries"
        />
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

.header-actions {
    display: flex;
    gap: 0.5rem;
}

.refresh-button .scan-button{
    font-size: 0.875rem;
    background-color: #1e293b;
    color: #ffffff;
    padding: 0.25rem 0.75rem;
    border-radius: 0.25rem;
    border: 1px solid #334155;
    cursor: pointer;
    transition: background-color 0.2s, border-color 0.2s;
}

.refresh-button {
    background-color: #1e293b;
    border: 1px solid #334155;
}

.refresh-button:hover:not(:disabled) {
    background-color: #334155;
    border-color: #475569;
}

.scan-button {
    background-color: #2563eb;
    border: 1px solid rgba(59, 130, 246, 0.2);
}

.scan-button:hover:not(:disabled) {
    background-color: #3b82f6;
}

.refresh-button:disabled, .scan-button:disabled {
    background-color: #1e293b;
    color: #64748b;
    border-color: transparent;
    cursor: not-allowed;
}

.info-message {
    color: #64748b;
}

.status-empty {
    font-style: italic;
}
</style>
