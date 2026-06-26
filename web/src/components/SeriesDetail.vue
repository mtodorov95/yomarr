<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { RouterLink } from 'vue-router'
import type { Series, VolumeCover } from '../types'
import { useToast } from '@/composables/useToast'
import SeriesCovers from './SeriesCovers.vue'
import ChapterTable from './ChapterTable.vue'
import { getAssetUrl } from '@/utils/utils'

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

const toast = useToast()
const series = ref<Series | null>(null)
const groupedChapters = ref<GroupedChapter[]>([])
const loading = ref(true)
const searching = ref(false)
const refreshing = ref(false)
const updatingStatus = ref(false)

const activeTab = ref<'chapters' | 'covers'>('chapters')

watch(() => props.id, () => { activeTab.value = 'chapters' })

async function loadPageData() {
    loading.value = true
    try {
        const seriesId = Number(props.id)

        const seriesRes = await fetch(`/api/series?id=${seriesId}`)
        if (seriesRes.ok) {
            series.value = await seriesRes.json() ?? null
        }

        const chaptersRes = await fetch(`/api/chapters?series_id=${seriesId}`)
        if (!chaptersRes.ok) throw new Error('Failed to fetch chapters')
        groupedChapters.value = await chaptersRes.json() ?? []
    } catch (e) {
        console.error(e)
        toast.error("Failed to fetch library manifest records")
    } finally {
        loading.value = false
    }
}

async function refreshMetadata() {
    if (!series.value?.id) return
    refreshing.value = true
    try {
        const res = await fetch(`/api/series?id=${series.value.id}&action=refresh`, {
            method: 'POST'
        })
        if (!res.ok) throw new Error('Refresh request failed')

        const updatedData = await res.json()
        if (updatedData) {
            series.value = updatedData
        }

        toast.success('Metadata and art synced from upstream providers!')
    } catch (e) {
        console.error(e)
        toast.error('Failed to refresh metadata')
    } finally {
        refreshing.value = false
    }
}

async function searchMissingChapters() {
    if (!series.value?.id) return
    searching.value = true
    try {
        const res = await fetch(`/api/series/search-missing?series_id=${series.value.id}`, {
            method: 'POST'
        })
        if (!res.ok) throw new Error('Search request failed')
        toast.info('Search indexer triggered in background')
    } catch (e) {
        console.error(e)
        toast.error('Failed to wake tracking search target')
    } finally {
        searching.value = false
    }
}

async function updateSeriesCovers(updatedCovers: VolumeCover[], updatedThumbnail: string) {
    if (!series.value) return

    const previousCovers = [...(series.value.historical_covers ?? [])]
    const previousThumbnail = series.value.thumbnail

    series.value.historical_covers = updatedCovers
    series.value.thumbnail = updatedThumbnail

    try {
        const res = await fetch('/api/series', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(series.value)
        })

        if (!res.ok) throw new Error('Failed updating series covers records')
        toast.info('Covers updated successfully')
    } catch (e) {
        console.error(e)
        toast.error('Failed to sync cover updates')
        if (series.value) {
            series.value.historical_covers = previousCovers
            series.value.thumbnail = previousThumbnail
        }
    }
}

function handlePromoteCover(coverFile: string) {
    if (!series.value) return
    updateSeriesCovers(series.value.historical_covers ?? [], coverFile)
}

async function handleRemoveCover(coverFile: string) {
    if (!series.value) return

    try {
        const seriesId = series.value.id
        const res = await fetch(`/api/series?id=${seriesId}&cover=${encodeURIComponent(coverFile)}`, {
            method: 'DELETE'
        })

        if (!res.ok) throw new Error('Delete execution rejected')

        series.value = await res.json()
        toast.info('Cover wiped from storage disk')
    } catch (e) {
        console.error(e)
        toast.error('Failed to eliminate cover artwork')
    }
}

async function toggleMonitorStatus() {
    if (!series.value || updatingStatus.value) return

    updatingStatus.value = true
    const previousStatus = series.value.monitored
    series.value.monitored = !previousStatus

    try {
        const res = await fetch('/api/series', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(series.value)
        })

        if (!res.ok) throw new Error('Status modification update rejected by database layer')

        if (!series.value.monitored) {
            toast.info(`Stopped tracking automation loops for: ${series.value.title}`)
        } else {
            toast.success(`Resumed missing chapter tracking automation for: ${series.value.title}`)
        }
    } catch (e) {
        console.error(e)
        toast.error('Failed to update monitoring status')
        if (series.value) {
            series.value.monitored = previousStatus
        }
    } finally {
        updatingStatus.value = false
    }
}

onMounted(loadPageData)
</script>

<template>
    <div class="detail-container">
        <div v-if="loading && !series" class="loading-state">
            Querying Database ...
        </div>

        <div v-else-if="!series" class="empty-state">
            <p class="empty-title">Series Record Hidden</p>
            <p class="empty-subtitle">The backend database failed to return a match for this ID reference.</p>
            <RouterLink to="/" class="back-button-link" style="margin-top: 1rem;">
                Back to Library
            </RouterLink>
        </div>

        <template v-else>
            <div class="arr-series-banner">
                <div class="banner-poster-container">
                    <img v-if="series.thumbnail" :src="getAssetUrl(series.thumbnail, series.path)" :alt="series.title"
                        class="banner-poster-img" />
                    <div v-else class="banner-poster-fallback">No Cover</div>
                </div>

                <div class="banner-content-info">
                    <div class="banner-actions-line">
                        <RouterLink to="/" class="back-button-link">
                            ← Back to Library
                        </RouterLink>

                        <div class="banner-actions-wrapper">
                            <button @click="refreshMetadata" :disabled="refreshing || searching"
                                class="arr-action-btn refresh-accent">
                                <span>{{ refreshing ? '⏳' : '🔄' }}</span>
                                {{ refreshing ? 'Syncing...' : 'Refresh Metadata' }}
                            </button>

                            <button @click="searchMissingChapters" :disabled="searching || refreshing"
                                class="arr-action-btn search-accent">
                                <span>{{ searching ? '⏳' : '🔍' }}</span>
                                {{ searching ? 'Searching missing...' : 'Search Missing' }}
                            </button>
                        </div>
                    </div>

                    <div class="series-identity-block">
                        <div class="series-headline-row">
                            <button @click="toggleMonitorStatus" class="monitor-toggle-btn"
                                :class="{ 'is-unmonitored': !series.monitored }"
                                :disabled="updatingStatus"
                                :title="!series.monitored ? 'Click to Start Monitoring' : 'Click to Stop Monitoring'"
                                aria-label="Toggle series monitoring tracking status state">
                                <svg v-if="series.monitored"
                                    xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor"
                                    class="monitor-icon">
                                    <path
                                        d="M5.625 1.5c-1.036 0-1.875.84-1.875 1.875v17.25c0 1.235 1.34 1.96 2.343 1.249l4.773-3.385a1.25 1.25 0 0 1 1.458 0l4.773 3.385c1.003.711 2.343-.014 2.343-1.249V3.375c0-1.036-.84-1.875-1.875-1.875H5.625Z" />
                                </svg>
                                <svg v-else xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"
                                    stroke-width="2.5" stroke="currentColor" class="monitor-icon">
                                    <path stroke-linecap="round" stroke-linejoin="round"
                                        d="M17.593 3.322c1.1.154 1.907 1.1 1.907 2.185V21L12 17.25 4.5 21V5.507c0-1.086.807-2.03 1.907-2.185a48.507 48.507 0 0 1 11.186 0Z" />
                                </svg>
                            </button>
                            <h1 class="main-series-headline">{{ series.title }}</h1>
                        </div>

                        <div v-if="series.author || series.artist" class="author-label-line">
                            <span v-if="series.author">
                                by <span class="author-name-highlight">{{ series.author }}</span>
                            </span>
                            <span v-if="series.author && series.artist" class="credit-separator"> · </span>
                            <span v-if="series.artist">
                                art by <span class="author-name-highlight">{{ series.artist }}</span>
                            </span>
                        </div>

                        <div class="arr-metadata-pill-row">
                            <span class="arr-pill storage-path-pill">
                                📁 {{ series.path }}
                            </span>
                            <span class="arr-pill status-pill" :class="series.status?.toLowerCase()">
                                {{ series.status || 'Unknown State' }}
                            </span>
                            <template v-if="series.genres">
                                <span
                                    v-for="genre in (Array.isArray(series.genres) ? series.genres : String(series.genres).split(','))"
                                    :key="genre" class="arr-pill genre-pill">
                                    {{ genre.trim() }}
                                </span>
                            </template>
                        </div>

                        <p class="series-plot-synopsis">
                            {{ series.description || 'No descriptive structural profile summary details synchronized or indexed inside database fields yet.' }}
                        </p>
                    </div>

                    <div class="arr-sub-tab-navigation">
                        <button @click="activeTab = 'chapters'"
                            :class="['navigation-tab-btn', activeTab === 'chapters' ? 'is-active' : '']">
                            Chapters ({{ groupedChapters.length }})
                        </button>
                        <button @click="activeTab = 'covers'"
                            :class="['navigation-tab-btn', activeTab === 'covers' ? 'is-active' : '']">
                            Covers ({{ series.historical_covers?.length ?? 0 }})
                        </button>
                    </div>
                </div>
            </div>

            <div v-if="activeTab === 'chapters'" class="tab-pane-view">
                <ChapterTable :chapters="groupedChapters" />
            </div>

            <SeriesCovers v-else-if="activeTab === 'covers'" :covers="series.historical_covers"
                :currentThumbnail="series.thumbnail" :seriesPath="series.path" @promote="handlePromoteCover"
                @remove="handleRemoveCover" />
        </template>
    </div>
</template>

<style scoped>
@keyframes fadeIn {
    from {
        opacity: 0;
        transform: translateY(2px);
    }

    to {
        opacity: 1;
        transform: translateY(0);
    }
}

.detail-container {
    width: 100%;
    animation: fadeIn 0.15s cubic-bezier(0.16, 1, 0.3, 1);
}

.arr-series-banner {
    display: flex;
    flex-direction: column;
    gap: 2rem;
    background-color: #111827;
    border: 1px solid #1f2937;
    padding: 1.5rem;
    border-radius: 0.5rem;
    margin-bottom: 2rem;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.3);
}

@media (min-width: 768px) {
    .arr-series-banner {
        flex-direction: row;
        align-items: flex-start;
    }
}

.banner-poster-container {
    width: 11.5rem;
    aspect-ratio: 2 / 3;
    flex-shrink: 0;
    border-radius: 0.25rem;
    overflow: hidden;
    background-color: #030712;
    border: 1px solid #374151;
    box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.5);
    margin: 0 auto;
}

@media (min-width: 768px) {
    .banner-poster-container {
        margin: 0;
    }
}

.banner-poster-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.banner-poster-fallback {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #4b5563;
    font-size: 0.875rem;
    font-weight: 700;
}

.banner-content-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
}

.banner-actions-line {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    margin-bottom: 1.25rem;
    gap: 1rem;
}

.back-button-link {
    text-decoration: none;
    background-color: #1f2937;
    border: 1px solid #374151;
    padding: 0.4rem 0.85rem;
    border-radius: 0.25rem;
    font-size: 0.75rem;
    font-weight: 700;
    color: #f3f4f6;
    transition: background-color 0.15s;
}

.back-button-link:hover {
    background-color: #4b5563;
}

.banner-actions-wrapper {
    display: inline-flex;
    gap: 0.5rem;
}

.arr-action-btn {
    background-color: #2563eb;
    border: 1px solid transparent;
    padding: 0.4rem 0.85rem;
    border-radius: 0.25rem;
    font-size: 0.75rem;
    font-weight: 700;
    color: #ffffff;
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    cursor: pointer;
    transition: background-color 0.15s;
}

.arr-action-btn.search-accent {
    background-color: #2563eb;
}

.arr-action-btn.search-accent:hover:not(:disabled) {
    background-color: #3b82f6;
}

.arr-action-btn.refresh-accent {
    background-color: #4b5563;
    border: 1px solid #6b7280;
}

.arr-action-btn.refresh-accent:hover:not(:disabled) {
    background-color: #374151;
}

.arr-action-btn:disabled {
    background-color: #374151;
    color: #9ca3af;
    cursor: not-allowed;
}

.series-identity-block {
    display: flex;
    flex-direction: column;
    margin-bottom: 1.5rem;
}

.main-series-headline {
    font-size: 2rem;
    font-weight: 400;
    color: #ffffff;
    margin: 0;
    letter-spacing: -0.01em;
    line-height: 1.15;
}

.series-headline-row {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-top: 0.25rem;
}

.monitor-toggle-btn {
    background: transparent;
    border: none;
    padding: 0.25rem;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    color: #38bdf8;
    border-radius: 0.375rem;
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.monitor-toggle-btn:hover:not(:disabled) {
    background-color: #1f2937;
    transform: scale(1.1);
}

.monitor-toggle-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

.monitor-icon {
    width: 1.75rem;
    height: 1.75rem;
    filter: drop-shadow(0 0 4px rgba(56, 189, 248, 0.2));
}

.monitor-toggle-btn.is-unmonitored {
    color: #ef4444;
}

.monitor-toggle-btn.is-unmonitored .monitor-icon {
    filter: drop-shadow(0 0 4px rgba(239, 68, 68, 0.1));
}

.author-label-line {
    font-size: 0.9rem;
    color: #9ca3af;
    margin-top: 0.25rem;
}

.credit-separator {
    color: #4b5563;
    margin: 0 0.25rem;
    font-weight: bold;
}

.author-name-highlight {
    color: #38bdf8;
    font-weight: 600;
}

.arr-metadata-pill-row {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4rem;
    margin-top: 0.75rem;
    margin-bottom: 0.85rem;
}

.arr-pill {
    font-size: 0.75rem;
    color: #d1d5db;
    background-color: #1f2937;
    padding: 0.2rem 0.5rem;
    border-radius: 0.25rem;
    font-weight: 500;
    display: inline-flex;
    align-items: center;
}

.storage-path-pill {
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    background-color: #111827;
    border: 1px solid #1f2937;
}

.status-pill.ongoing {
    color: #60a5fa;
    background-color: rgba(29, 78, 216, 0.2);
}

.status-pill.completed {
    color: #4ade80;
    background-color: rgba(21, 128, 61, 0.2);
}

.status-pill.hiatus {
    color: #fdba74;
    background-color: rgba(194, 65, 12, 0.2);
    border: 1px solid rgba(194, 65, 12, 0.4);
}

.status-pill.cancelled {
    color: #f87171;
    background-color: rgba(127, 29, 29, 0.3);
    border: 1px solid rgba(185, 28, 28, 0.4);
}

.genre-pill {
    background-color: #374151;
    color: #e5e7eb;
}

.series-plot-synopsis {
    font-size: 0.85rem;
    line-height: 1.5;
    color: #9ca3af;
    margin: 0;
    max-width: 48rem;
}

.arr-sub-tab-navigation {
    display: flex;
    gap: 1.25rem;
    border-bottom: 1px solid #1f2937;
    margin-top: auto;
}

.navigation-tab-btn {
    background: none;
    border: none;
    color: #9ca3af;
    font-size: 0.85rem;
    font-weight: 600;
    padding: 0.5rem 0.1rem;
    cursor: pointer;
    position: relative;
    transition: color 0.15s;
}

.navigation-tab-btn:hover {
    color: #ffffff;
}

.navigation-tab-btn.is-active {
    color: #38bdf8 !important;
}

.navigation-tab-btn.is-active::after {
    content: '';
    position: absolute;
    bottom: -1px;
    left: 0;
    right: 0;
    height: 2px;
    background-color: #38bdf8;
}

.tab-pane-view {
    width: 100%;
}

.loading-state {
    color: #9ca3af;
    font-style: italic;
    padding: 3rem 0;
    text-align: center;
}

.empty-state {
    color: #9ca3af;
    text-align: center;
    padding: 5rem 1rem;
    display: flex;
    flex-direction: column;
    align-items: center;
}

.empty-title {
    font-size: 1.25rem;
    font-weight: 700;
    margin: 0 0 0.5rem 0;
    color: #ffffff;
}

.empty-subtitle {
    font-size: 0.875rem;
    color: #9ca3af;
    margin: 0;
}
</style>
