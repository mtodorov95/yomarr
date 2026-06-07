<script setup lang="ts">
import type { Series } from '../types'

const props = defineProps<{
    series: Series
}>()

const emit = defineEmits<{
    (e: 'delete', id: number): void
    (e: 'select', series: Series): void
}>()

function handleDelete() {
    if (props.series.id === undefined) return
    emit('delete', props.series.id)
}

function getImageUrl(): string {
    if (!props.series.thumbnail) return ''
    return `/api/assets?path=${encodeURIComponent(props.series.path + '/' + props.series.thumbnail)}`
}
</script>

<template>
    <div @click="emit('select', series)" class="series-poster-card">
        <div class="poster-frame">
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
                <button 
                    @click.stop="handleDelete" 
                    class="overlay-action-btn delete"
                    title="Remove Target Series"
                >
                    🗑️ Remove
                </button>
            </div>

            <div 
                class="monitoring-ribbon" 
                :class="{ 
                    'status-ongoing': series.status?.toLowerCase() === 'ongoing',
                    'status-completed': series.status?.toLowerCase() === 'completed'
                }"
                :title="'Tracking Status: ' + series.status"
            ></div>
        </div>

        <div class="poster-meta">
            <h3 class="series-title-text" :title="series.title">{{ series.title }}</h3>
            <span class="series-subtitle-path" :title="series.path">{{ series.path }}</span>
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

.overlay-action-btn {
    font-size: 0.75rem;
    font-weight: 700;
    border: 1px solid transparent;
    padding: 0.4rem 0.75rem;
    border-radius: 0.25rem;
    cursor: pointer;
    transition: all 0.15s;
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

.monitoring-ribbon {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 5px;
    background-color: #475569;
}

.monitoring-ribbon.status-ongoing {
    background-color: #15803d;
    box-shadow: 0 -2px 8px rgba(21, 128, 61, 0.4);
}

.monitoring-ribbon.status-completed {
    background-color: #1d4ed8;
}

.poster-meta {
    padding: 0.5rem 0.25rem 0 0.25rem;
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
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

.series-subtitle-path {
    font-size: 0.7rem;
    color: #64748b;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
</style>
