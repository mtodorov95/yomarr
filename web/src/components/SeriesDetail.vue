<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { Series, Chapter } from '../types'
import { useToast } from '@/composables/useToast';

interface LanguageVariant {
    id: number
    status: string
    language: string
    file_path?: string
}

interface GroupedChapter {
    number: number
    volume: number | null
    variants: LanguageVariant[]
}

const props = defineProps<{
    id: string
}>()

const toast = useToast();
const series = ref<Series | null>(null)
const groupedChapters = ref<GroupedChapter[]>([])
const loading = ref(true)
const searching = ref(false)

const activeTab = ref<'chapters' | 'covers'>('chapters')

async function loadPageData() {
    loading.value = true
    try {
        const seriesId = Number(props.id)
        
        const seriesRes = await fetch('/api/series')
        if (seriesRes.ok) {
            const allSeries: Series[] = await seriesRes.json() ?? []
            series.value = allSeries.find(s => s.id === seriesId) || null
        }

        const chaptersRes = await fetch(`/api/chapters?series_id=${seriesId}`)
        if (!chaptersRes.ok) throw new Error('Failed to fetch chapters')
        groupedChapters.value = await chaptersRes.json() ?? [];
    } catch (e) {
        console.error(e)
        toast.error("Faield to fetch chapters")
    } finally {
        loading.value = false
    }
}

async function searchMissingChapters() {
  if (!series.value?.id) return
  try {
    const res = await fetch(`/api/series/search-missing?series_id=${series.value.id}`, {
      method: 'POST'
    })
    if (!res.ok) throw new Error('Search request failed')
    toast.info('Search started in background')
  } catch (e) {
    console.error(e)
    toast.error('Search trigger failed')
  }
}

function getAssetUrl(filename: string | undefined): string {
    if (!series.value || !filename) return ''
    return `/api/assets?path=${encodeURIComponent(series.value.path + '/' + filename)}`
}

onMounted(loadPageData)
</script>

<template>
    <div class="detail-container">
        <div v-if="loading && !series" class="loading-state">
            Querying Chapter Database Manifest...
        </div>

        <div v-else-if="!series" class="empty-state">
            <p class="empty-title">Series Not Found</p>
            <p class="empty-subtitle">The requested series record could not be located in your database.</p>
            <RouterLink to="/" class="back-button" style="display: inline-flex; margin-top: 1rem;">
                Back to Library
            </RouterLink>
        </div>

        <template v-else>
            <div class="series-profile-card">
                <div class="profile-poster-frame">
                    <img 
                        v-if="series.thumbnail" 
                        :src="getAssetUrl(series.thumbnail)" 
                        :alt="series.title" 
                        class="profile-poster"
                    />
                    <div v-else class="poster-fallback">No Cover</div>
                </div>

                <div class="profile-details">
                    <div class="detail-header">
                        <div class="meta-block">
                            <RouterLink to="/" class="back-button">
                                <span>←</span> Back to Library
                            </RouterLink>
                            <div>
                                <h2 class="series-title">{{ series.title }}</h2>
                                <p class="path-text">
                                    Storage Path: <code class="path-code">{{ series.path }}</code>
                                </p>
                            </div>
                        </div>

                        <div>
                            <button 
                                @click="searchMissingChapters()"
                                :disabled="searching"
                                class="search-missing-button"
                            >
                                <span v-if="searching" class="spinner">⏳</span>
                                <span v-else>🔍</span>
                                {{ searching ? 'Searching Nyaa...' : 'Search Missing' }}
                            </button>
                        </div>
                    </div>

                    <div class="tab-row-container">
                        <button 
                            @click="activeTab = 'chapters'" 
                            :class="['tab-toggle-item', activeTab === 'chapters' ? 'tab-active' : '']"
                        >
                            Chapters Manifest ({{ groupedChapters.length }})
                        </button>
                        <button 
                            @click="activeTab = 'covers'" 
                            :class="['tab-toggle-item', activeTab === 'covers' ? 'tab-active' : '']"
                        >
                            Historical Covers ({{ series.historical_covers?.length ?? 0 }})
                        </button>
                    </div>
                </div>
            </div>

            <div v-if="activeTab === 'chapters'">
                <div v-if="groupedChapters && groupedChapters.length > 0" class="table-card">
                    <div class="table-responsive">
                        <table class="manifest-table">
                            <thead>
                                <tr class="table-header-row">
                                    <th class="table-th">Chapter Number</th>
                                    <th class="table-th">Volume</th>
                                    <th class="table-th">Available Releases / Language Status</th>
                                </tr>
                            </thead>
                            <tbody class="table-body">
                                <tr v-for="ch in groupedChapters" :key="ch.number" class="table-row">
                                    <td class="table-td td-chapter">
                                        Chapter {{ ch.number }}
                                    </td>
                                    <td class="table-td td-volume">
                                        {{ ch.volume !== null && ch.volume !== undefined ? `Vol. ${ch.volume}` : '—' }}
                                    </td>
                                    <td class="table-td">
                                        <div class="variant-badge-stack">
                                            <div 
                                                v-for="v in ch.variants" 
                                                :key="v.id"
                                                class="variant-wrapper"
                                                :title="v.file_path || 'No file registered'"
                                            >
                                                <span :class="['status-badge', v.status === 'Downloaded' ? 'badge-downloaded' : 'badge-missing']">
                                                    <span class="lang-token">{{ v.language }}:</span> {{ v.status }}
                                                </span>
                                            </div>
                                        </div>
                                    </td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </div>

                <div v-else class="empty-state">
                    <p class="empty-title">Manifest Unpopulated</p>
                    <p class="empty-subtitle">MangaDex feed contains no records or synchronization hasn't run yet.</p>
                </div>
            </div>

            <div v-if="activeTab === 'covers'">
                <div v-if="series.historical_covers && series.historical_covers.length > 0" class="covers-archive-grid">
                    <div 
                        v-for="(coverFile, index) in series.historical_covers" 
                        :key="index" 
                        class="archive-cover-card"
                    >
                        <div class="archive-image-frame">
                            <img 
                                :src="getAssetUrl(coverFile)" 
                                alt="Volume Artwork Variant" 
                                class="archive-image" 
                                loading="lazy"
                            />
                        </div>
                        <div class="archive-meta-tag">
                            Variant {{ index + 1 }}
                        </div>
                    </div>
                </div>

                <div v-else class="empty-state">
                    <p class="empty-title">No Historical Variants Saved</p>
                    <p class="empty-subtitle">This series tracker only has its primary thumbnail record downloaded locally.</p>
                </div>
            </div>
        </template>
    </div>
</template>

<style scoped>
@keyframes fadeIn {
    from { opacity: 0; transform: translateY(4px); }
    to { opacity: 1; transform: translateY(0); }
}

@keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
}

.detail-container {
    width: 100%;
    max-width: 56rem;
    animation: fadeIn 0.2s ease-out;
}

.series-profile-card {
    display: flex;
    gap: 1.5rem;
    background-color: #1e293b;
    border: 1px solid #334155;
    padding: 1.25rem;
    border-radius: 1rem;
    margin-bottom: 1.5rem;
    box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.3);
}

.profile-poster-frame {
    width: 7.5rem;
    height: 11rem;
    flex-shrink: 0;
    border-radius: 0.5rem;
    overflow: hidden;
    background-color: #0f172a;
    border: 1px solid #475569;
}

.profile-poster {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.poster-fallback {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #475569;
    font-size: 0.875rem;
    font-weight: 700;
}

.profile-details {
    flex: 1;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    min-width: 0;
}

.detail-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 1rem;
    width: 100%;
}

.meta-block {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
}

.tab-row-container {
    display: flex;
    gap: 0.5rem;
    border-bottom: 1px solid #334155;
    margin-top: 1rem;
    padding-bottom: 0.125rem;
}

.tab-toggle-item {
    background: none;
    border: none;
    color: #94a3b8;
    font-size: 0.875rem;
    font-weight: 700;
    padding: 0.5rem 1rem;
    cursor: pointer;
    border-radius: 0.375rem 0.375rem 0 0;
    transition: color 0.2s, background-color 0.2s;
}

.tab-toggle-item:hover {
    color: #f8fafc;
    background-color: rgba(15, 23, 42, 0.3);
}

.tab-active {
    color: #60a5fa !important;
    background-color: #0f172a !important;
    border: 1px solid #334155;
    border-bottom-color: #0f172a;
    position: relative;
    margin-bottom: -2px;
}

.covers-archive-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
}

@media (min-width: 480px) {
    .covers-archive-grid {
        grid-template-columns: repeat(3, 1fr);
    }
}

@media (min-width: 640px) {
    .covers-archive-grid {
        grid-template-columns: repeat(4, 1fr);
    }
}

.archive-cover-card {
    background-color: #1e293b;
    border: 1px solid #334155;
    border-radius: 0.5rem;
    padding: 0.5rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    align-items: center;
}

.archive-image-frame {
    width: 100%;
    aspect-ratio: 2 / 3;
    overflow: hidden;
    border-radius: 0.375rem;
    background-color: #0f172a;
}

.archive-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.archive-meta-tag {
    font-size: 0.75rem;
    color: #94a3b8;
    font-weight: 600;
}

.back-button {
    text-decoration: none;
    background-color: #0f172a;
    border: 1px solid #334155;
    padding: 0.375rem 0.75rem;
    border-radius: 0.5rem;
    font-size: 0.75rem;
    font-weight: 700;
    color: #ffffff;
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    cursor: pointer;
    align-self: flex-start;
    transition: background-color 0.2s;
}

.back-button:hover {
    background-color: #334155;
}

.series-title {
    font-size: 1.5rem;
    font-weight: 900;
    color: #60a5fa;
    letter-spacing: -0.025em;
    margin: 0;
}

.path-text {
    font-size: 0.75rem;
    color: #94a3b8;
    margin-top: 0.125rem;
    margin-bottom: 0;
}

.path-code {
    color: #cbd5e1;
    background-color: rgba(0, 0, 0, 0.4);
    padding: 0.125rem 0.375rem;
    border-radius: 0.25rem;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.search-missing-button {
    background-color: #2563eb;
    border: 1px solid rgba(59, 130, 246, 0.3);
    padding: 0.5rem 1rem;
    border-radius: 0.75rem;
    font-size: 0.875rem;
    font-weight: 700;
    color: #ffffff;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    cursor: pointer;
    box-shadow: 0 10px 15px -3px rgba(59, 130, 246, 0.1);
    transition: background-color 0.2s;
}

.search-missing-button:hover:not(:disabled) {
    background-color: #3b82f6;
}

.search-missing-button:disabled {
    background-color: #334155;
    color: #64748b;
    border-color: transparent;
    cursor: not-allowed;
    box-shadow: none;
}

.spinner {
    display: inline-block;
    animation: spin 1s linear infinite;
    font-size: 0.75rem;
}

.loading-state {
    color: #64748b;
    font-style: italic;
    padding: 2rem 0;
    text-align: center;
}

.table-card {
    background-color: #0f172a;
    border: 1px solid #334155;
    border-radius: 1rem;
    overflow: hidden;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
}

.table-responsive {
    overflow-x: auto;
}

.manifest-table {
    width: 100%;
    text-align: left;
    border-collapse: collapse;
}

.table-header-row {
    border-bottom: 1px solid #334155;
    background-color: rgba(15, 23, 42, 0.4);
    color: #94a3b8;
    font-size: 0.75rem;
    text-transform: uppercase;
    font-weight: 900;
    letter-spacing: 0.05em;
}

.table-th {
    padding: 1rem;
}

.table-body {
    color: #e2e8f0;
    font-size: 0.875rem;
    font-weight: 500;
}

.table-body :deep(tr:not(:last-child)) {
    border-bottom: 1px solid rgba(51, 65, 85, 0.5);
}

.table-row {
    transition: background-color 0.2s;
}

.table-row:hover {
    background-color: rgba(51, 65, 85, 0.2);
}

.table-td {
    padding: 1rem;
}

.td-chapter {
    font-weight: 700;
    color: #f8fafc;
}

.td-volume {
    color: #94a3b8;
    font-weight: 400;
}

.variant-badge-stack {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    align-items: center;
}

.variant-wrapper {
    display: inline-flex;
}

.status-badge {
    padding: 0.25rem 0.75rem;
    font-size: 0.75rem;
    font-weight: 700;
    border-radius: 0.5rem;
    border: 1px solid transparent;
    letter-spacing: 0.025em;
    text-transform: uppercase;
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
}

.lang-token {
    opacity: 0.8;
    font-weight: 900;
}

.badge-downloaded {
    background-color: rgba(34, 197, 94, 0.1);
    color: #4ade80;
    border-color: rgba(34, 197, 94, 0.2);
}

.badge-missing {
    background-color: rgba(239, 68, 68, 0.1);
    color: #f87171;
    border-color: rgba(239, 68, 68, 0.2);
}

.empty-state {
    color: #94a3b8;
    text-align: center;
    padding: 3rem 0;
    background-color: rgba(30, 41, 59, 0.4);
    border: 1px solid #1e293b;
    border-style: dashed;
    border-radius: 1rem;
}

.empty-title {
    font-size: 1.125rem;
    font-weight: 700;
    margin-top: 0;
    margin-bottom: 0.25rem;
}

.empty-subtitle {
    font-size: 0.75rem;
    color: #64748b;
    margin: 0;
}
</style>
