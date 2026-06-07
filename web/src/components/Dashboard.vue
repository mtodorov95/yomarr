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
        <header class="action-bar">
            <div class="action-group">
                <button 
                    @click="runLibraryScan" 
                    :disabled="scanning || loading" 
                    class="action-item-btn primary"
                >
                    <span class="btn-icon">🔄</span>
                    <span class="btn-text">{{ scanning ? 'Scanning...' : 'Update Library' }}</span>
                </button>
                
                <button 
                    @click="fetchSeries" 
                    :disabled="loading" 
                    class="action-item-btn"
                >
                    <span class="btn-icon">🔁</span>
                    <span class="btn-text">Refresh Status</span>
                </button>
            </div>
            
            <div class="view-group-placeholder">
                <span class="series-count-badge" v-if="series.length > 0">
                    Total Items: {{ series.length }}
                </span>
            </div>
        </header>

        <div class="content-panel">
            <div v-if="loading" class="info-message loading-state">
                <div class="spinner"></div>
                <span>Syncing Database Profile Layers...</span>
            </div>

            <template v-else>
                <SeriesList 
                    v-if="series.length > 0" 
                    :seriesList="series" 
                    @delete="triggerDeleteConfirmation" 
                    @select="handleSelect" 
                />
                <div v-else class="info-message empty-state">
                    <span class="empty-icon">📂</span>
                    <h3>Your Library is Completely Empty</h3>
                    <p>Get started by running a manual index scan or adding new track targets.</p>
                    <RouterLink to="/add" class="empty-cta">Add a Title Target</RouterLink>
                </div>
            </template>
        </div>

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
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
}

.action-bar {
    background-color: #1e293b;
    border: 1px solid #334155;
    border-radius: 0.5rem;
    padding: 0.5rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}

.action-group {
    display: flex;
    gap: 0.25rem;
}

.action-item-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background: transparent;
    color: #cbd5e1;
    border: 1px solid transparent;
    padding: 0.5rem 1rem;
    font-size: 0.875rem;
    font-weight: 600;
    border-radius: 0.375rem;
    cursor: pointer;
    transition: all 0.15s ease;
}

.action-item-btn:hover:not(:disabled) {
    background-color: #334155;
    color: #ffffff;
    border-color: #475569;
}

.action-item-btn.primary {
    color: #38bdf8;
}

.action-item-btn.primary:hover:not(:disabled) {
    background-color: #0c4a6e;
    border-color: #0369a1;
}

.action-item-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
}

.series-count-badge {
    font-size: 0.75rem;
    background-color: #0f172a;
    color: #94a3b8;
    padding: 0.35rem 0.75rem;
    border-radius: 9999px;
    font-weight: 600;
    border: 1px solid #334155;
    margin-right: 0.5rem;
}

.content-panel {
    width: 100%;
}

.info-message {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 5rem 2rem;
    color: #64748b;
}

.loading-state {
    gap: 1rem;
    font-size: 0.95rem;
    font-weight: 500;
}

.spinner {
    width: 2rem;
    height: 2rem;
    border: 3px solid #334155;
    border-top-color: #38bdf8;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
}

@keyframes spin {
    to { transform: rotate(360deg); }
}

.empty-state {
    background-color: #1e293b;
    border: 2px dashed #334155;
    border-radius: 0.75rem;
}

.empty-icon {
    font-size: 3rem;
    margin-bottom: 1rem;
}

.empty-state h3 {
    margin: 0 0 0.5rem 0;
    color: #f1f5f9;
    font-size: 1.25rem;
}

.empty-state p {
    margin: 0 0 1.5rem 0;
    max-width: 24rem;
    font-size: 0.95rem;
    line-height: 1.5;
}

.empty-cta {
    background-color: #2563eb;
    color: #ffffff;
    text-decoration: none;
    padding: 0.625rem 1.25rem;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    font-weight: 600;
    transition: background-color 0.15s;
}

.empty-cta:hover {
    background-color: #3b82f6;
}
</style>
