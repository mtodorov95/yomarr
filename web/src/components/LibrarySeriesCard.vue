<script setup lang="ts">
import { useToast } from '@/composables/useToast';
import type { Series } from '../types'
import { computed, ref } from 'vue';

const props = defineProps<{
    series: Series
}>()

const toast = useToast()
const searching = ref(false)

const emit = defineEmits<{
    (e: 'delete', id: number): void
    (e: 'select', series: Series): void
}>()

function handleDelete() {
    if (props.series.id === undefined) return
    emit('delete', props.series.id)
}

async function handleSearchMissing() {
    if (!props.series.id) return
    searching.value = true
    try {
        const res = await fetch(`/api/series/search-missing?series_id=${props.series.id}`, {
            method: 'POST'
        })
        if (!res.ok) throw new Error('Search failed')
        toast.info(`Search indexer triggered for: ${props.series.title}`)
    } catch (e) {
        console.error(e)
        toast.error(`Failed to search missing chapters for ${props.series.title}`)
    } finally {
        searching.value = false
    }
}

function getCompletionPercentage(): number {
    if (!props.series.total_chapters || props.series.total_chapters <= 0) return 100
    
    const downloaded = props.series.downloaded_count || 0
    const percentage = Math.round((downloaded / props.series.total_chapters) * 100)
    
    return Math.min(Math.max(percentage, 0), 100)
}

function getImageUrl(): string {
    if (!props.series.thumbnail) return ''
    return `/api/assets?path=${encodeURIComponent(props.series.path + '/' + props.series.thumbnail)}`
}

const truncatedTitle = computed(() => {
    const limit = 18
    if (props.series.title.length <= limit) return props.series.title;
    return props.series.title.slice(0, limit) + '...';
})
</script>

<template>
    <div @click="emit('select', series)" class="series-poster-card" :class="{ 'is-unmonitored': !series.monitored }">
        <div class="poster-frame">
            <div 
                class="monitoring-ribbon" 
                :class="series.monitored ? 'monitored' : 'unmonitored'"
                :title="series.monitored ? 'Monitored' : 'Unmonitored'"
            >
            </div>

            <img 
                v-if="series.thumbnail" 
                :src="getImageUrl()" 
                :alt="series.title" 
                class="poster-image"
                loading="lazy"
            />
            <div v-else class="poster-fallback">
                <span class="fallback-icon">📖</span>
                <span>No Cover Loaded</span>
            </div>

            <div class="poster-hover-overlay">
                <div class="overlay-actions-row">
                    <button 
                        @click.stop="handleSearchMissing" 
                        :disabled="searching"
                        class="overlay-action-btn search"
                        :class="{ 'is-spinning': searching }"
                        title="Search Missing Chapters"
                    >
                            🔍
                    </button>
                    <button 
                        @click.stop="handleDelete" 
                        class="overlay-action-btn delete"
                        title="Remove Series"
                    >
                        🗑️
                    </button>
                </div>
            </div>

            <div 
                class="monitoring-progress-wrapper"
                :title="`Status: ${series.status} | Progress: ${getCompletionPercentage()}%`"
            >
                <div class="progress-track-bg"></div>
                <div 
                    class="progress-fill-bar"
                    :class="series.downloading ? 'status-downloading' : { 
                        'status-ongoing': series.status?.toLowerCase() === 'ongoing',
                        'status-completed': series.status?.toLowerCase() === 'completed',
                        'status-hiatus': series.status?.toLowerCase() === 'hiatus'
                    }"
                    :style="{ width: series.downloading ? '100%' : getCompletionPercentage() + '%' }"
                ></div>
                    </div>
                </div>

        <div class="poster-meta">
            <h3 class="series-title-text" :title="series.title">{{ truncatedTitle }}</h3>
        </div>
    </div>
</template>

<style scoped>
.series-poster-card {
    display: flex;
    flex-direction: column;
    cursor: pointer;
    background-color: transparent;
    transition: transform 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

.series-poster-card.is-unmonitored {
    opacity: 0.75;
}

.series-poster-card.is-unmonitored:hover {
    opacity: 1;
}

.series-poster-card:hover {
    transform: translateY(-4px);
}

.poster-frame {
    position: relative;
    width: 100%;
    aspect-ratio: 2 / 3;
    background-color: #1e293b;
    border-radius: 0.375rem;
    border: 1px solid #334155;
    overflow: hidden;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.2), 0 2px 4px -1px rgba(0, 0, 0, 0.1);
    transition: border-color 0.2s;
}

.series-poster-card:hover .poster-frame {
    border-color: #475569;
    box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.4), 0 4px 6px -2px rgba(0, 0, 0, 0.2);
}

.monitoring-ribbon {
    position: absolute;
    top: 0;
    left: 0;
    z-index: 10;
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-bottom-right-radius: 0.375rem;
    box-shadow: 1px 1px 4px rgba(0, 0, 0, 0.3);
}

.monitoring-ribbon.monitored {
    background-color: #0ea5e9;
}

.monitoring-ribbon.unmonitored {
    background-color: #7f1d1d;
}

.poster-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.poster-fallback {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    font-size: 0.75rem;
    color: #475569;
    font-weight: 700;
    text-transform: uppercase;
    background-color: #0f172a;
}

.fallback-icon {
    font-size: 1.5rem;
}

.poster-hover-overlay {
    position: absolute;
    inset: 0;
    background: linear-gradient(to bottom, rgba(15, 23, 42, 0.1), rgba(15, 23, 42, 0.85));
    display: flex;
    align-items: flex-end;
    justify-content: center;
    padding-bottom: 1rem;
    opacity: 0;
    pointer-events: none;
    transition: opacity 0.2s ease-in-out;
}

.series-poster-card:hover .poster-hover-overlay {
    opacity: 1;
    pointer-events: auto;
}

.overlay-actions-row {
    display: flex;
    gap: 0.5rem;
    width: 100%;
    padding: 0 0.75rem;
    justify-content: center;
}

.overlay-action-btn {
    font-size: 0.75rem;
    font-weight: 700;
    border: 1px solid transparent;
    padding: 0.4rem 0.75rem;
    border-radius: 0.25rem;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    transition: all 0.15s;
    flex: 1;
    max-width: 50px;
}

.overlay-action-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

.overlay-action-btn.search {
    background-color: #1e293b;
    color: #38bdf8;
    border-color: #0c4a6e;
}

.overlay-action-btn.search:hover:not(:disabled) {
    background-color: #2563eb;
    color: #ffffff;
    border-color: #3b82f6;
}

.overlay-action-btn.delete {
    background-color: #1e293b;
    color: #f87171;
    border-color: #451a03;
}

.overlay-action-btn.delete:hover {
    background-color: #991b1b;
    color: #ffffff;
    border-color: #ef4444;
}

.overlay-action-btn.delete:hover:not(:disabled) {
    background-color: #991b1b;
    color: #ffffff;
    border-color: #ef4444;
}

.overlay-action-btn.search.is-spinning {
    animation: action-spin-clockwise 1s linear infinite;
}

@keyframes action-spin-clockwise {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
}

.monitoring-progress-wrapper {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 6px;
    background-color: #020617;
    overflow: hidden;
}

.progress-track-bg {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    background-color: #334155;
    opacity: 0.25;
}

.progress-fill-bar {
    position: absolute;
    top: 0;
    left: 0;
    bottom: 0;
    height: 100%;
    transition: width 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

.progress-fill-bar.status-ongoing { 
    background-color: #2563eb; 
}

.progress-fill-bar.status-completed { 
    background-color: #16a34a;
}

.progress-fill-bar.status-hiatus { 
    background-color: #f97316;
}

.progress-fill-bar.status-downloading { 
    background-color: #9333ea;
}

.progress-fill-bar.status-unmonitored { 
    background-color: #7f1d1d;
}

.poster-meta {
    padding: 0.5rem 0.25rem 0 0.25rem;
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
    align-items: center;
}

.series-title-text {
    font-weight: 700;
    color: #f1f5f9;
    margin: 0;
    font-size: 0.9rem;
    line-height: 1.25;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.series-poster-card:hover .series-title-text {
    color: #38bdf8;
}
</style>
