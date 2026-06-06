<script setup lang="ts">
import type { Series } from '../types'

defineProps<{
    seriesList: Series[]
}>()

const emit = defineEmits<{
    (e: 'delete', id: number): void
    (e: 'select', series: Series): void
}>()

function handleDelete(id: number | undefined) {
    if (id === undefined) return
    emit('delete', id)
}

function getImageUrl(series: Series): string {
    if (!series.thumbnail) return ''
    return `/api/assets?path=${encodeURIComponent(series.path + '/' + series.thumbnail)}`
}
</script>
<template>
    <div class="series-grid">
        <div 
            v-for="s in seriesList" 
            :key="s.id" 
            @click="$emit('select', s)"
            class="series-item-card"
        >
            <div class="card-cover-wrapper">
                <img 
                    v-if="s.thumbnail" 
                    :src="getImageUrl(s)" 
                    :alt="s.title" 
                    class="card-image"
                    loading="lazy"
                />
                <div v-else class="card-image-fallback">
                    <span>No Cover</span>
                </div>
            </div>

            <div class="card-details">
                <div class="card-main-content">
                    <div class="card-top-row">
                        <h3 class="series-card-title" :title="s.title">{{ s.title }}</h3>
                        <button 
                            @click.stop="handleDelete(s.id)" 
                            class="remove-button"
                        >
                            Remove
                        </button>
                    </div>
                    <p class="series-card-status">Status: {{ s.status }}</p>
                </div>
                <code class="series-card-path" :title="s.path">{{ s.path }}</code>
            </div>
        </div>
    </div>
</template>

<style scoped>
.series-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: 1.25rem;
    width: 100%;
    box-sizing: border-box;
}

@media (min-width: 640px) {
    .series-grid {
        grid-template-columns: repeat(2, 1fr);
    }
}

@media (min-width: 1024px) {
    .series-grid {
        grid-template-columns: repeat(3, 1fr);
    }
}

.series-item-card {
    background-color: #1e293b;
    cursor: pointer;
    border-radius: 0.75rem;
    border: 1px solid #334155;
    display: flex;
    flex-direction: row;
    height: 8.5rem;
    transition: background-color 0.2s, border-color 0.2s;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
    overflow: hidden;
}

.series-item-card:hover {
    background-color: rgba(30, 41, 59, 0.8);
    border-color: rgba(96, 165, 250, 0.4);
}

.card-cover-wrapper {
    width: 6rem;
    height: 100%;
    flex-shrink: 0;
    background-color: #0f172a;
    border-right: 1px solid #334155;
    position: relative;
}

.card-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.card-image-fallback {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.75rem;
    color: #475569;
    font-weight: 700;
    text-transform: uppercase;
}

.card-details {
    flex: 1;
    min-width: 0;
    padding: 0.875rem;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
}

.card-main-content {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
}

.card-top-row {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 0.5rem;
    width: 100%;
}

.series-card-title {
    font-weight: 700;
    color: #60a5fa;
    margin: 0;
    font-size: 1.05rem;
    line-height: 1.3;
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    white-space: normal;
}

.remove-button {
    font-size: 0.75rem;
    color: #f87171;
    background: none;
    border: none;
    cursor: pointer;
    padding: 0;
    margin-top: 0.15rem;
    transition: color 0.2s;
    flex-shrink: 0;
}

.remove-button:hover {
    color: #fca5a5;
}

.series-card-status {
    font-size: 0.75rem;
    color: #94a3b8;
    margin: 0;
}

.series-card-path {
    font-size: 0.70rem;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    color: #cbd5e1;
    background-color: #0f172a;
    padding: 0.25rem 0.5rem;
    border-radius: 0.25rem;
    display: block;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    width: 100%;
    box-sizing: border-box;
    border: 1px solid #1e293b;
}
</style>
