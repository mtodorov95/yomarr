<script setup lang="ts">
import { computed, ref } from 'vue'
import type { Series } from '../types'

const props = defineProps<{
    item: Series
}>()

const emit = defineEmits<{
    (e: 'import', series: Series): void
}>()

const localPath = ref(props.item.localPath || `/Manga/${props.item.title}`)

const coverSrc = computed(() => {
    if (!props.item.thumbnail) return ''
    
    return `/api/proxy-cover?url=${encodeURIComponent(props.item.thumbnail)}`
})
</script>

<template>
    <div class="card-container">
        <div class="card-content">
            <div class="thumbnail-wrapper">
                <img 
                    v-if="coverSrc" 
                    :src="coverSrc" 
                    :alt="item.title" 
                    class="series-thumbnail"
                    loading="lazy"
                />
                <div v-else class="thumbnail-fallback">
                    <span>No Image</span>
                </div>
            </div>

            <div class="main-details">
                <div class="card-header">
                    <div class="meta-section">
                        <p class="series-title">{{ item.title }}</p>
                        <p class="series-status">Status: {{ item.status }}</p>
                    </div>
                </div>

                <div class="action-row">
                    <input 
                        v-model="localPath" 
                        type="text" 
                        placeholder="Storage path..."
                        class="path-input" 
                    />
                    <button 
                        @click="emit('import', { ...item, localPath })"
                        class="import-button"
                    >
                        Import
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.card-container {
    padding: 1rem;
    background-color: #0f172a;
    border-radius: 0.5rem;
    border: 1px solid #334155;
}

.card-content {
    display: flex;
    gap: 1rem;
    align-items: flex-start;
}

.thumbnail-wrapper {
    width: 4.5rem;
    height: 6.5rem;
    flex-shrink: 0;
    border-radius: 0.375rem;
    overflow: hidden;
    background-color: #1e293b;
    border: 1px solid #475569;
}

.series-thumbnail {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.thumbnail-fallback {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.65rem;
    color: #64748b;
    text-transform: uppercase;
    font-weight: 700;
}

.main-details {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    min-width: 0;
}

.card-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
}

.meta-section {
    display: flex;
    flex-direction: column;
}

.series-title {
    font-weight: 700;
    color: #ffffff;
    font-size: 1.125rem;
    margin: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.series-status {
    font-size: 0.75rem;
    color: #94a3b8;
    margin: 0;
    margin-top: 0.25rem;
}

.action-row {
    display: flex;
    gap: 0.5rem;
    margin-top: auto;
}

.path-input {
    flex: 1;
    font-size: 0.875rem;
    background-color: #1e293b;
    border: 1px solid #475569;
    border-radius: 0.375rem;
    padding: 0.375rem 0.75rem;
    color: #ffffff;
    transition: border-color 0.2s;
}

.path-input:focus {
    outline: none;
    border-color: #3b82f6;
}

.import-button {
    background-color: #16a34a;
    color: #ffffff;
    font-size: 0.875rem;
    font-weight: 700;
    padding: 0.375rem 1rem;
    border-radius: 0.375rem;
    border: none;
    cursor: pointer;
    transition: background-color 0.2s;
    white-space: nowrap;
}

.import-button:hover {
    background-color: #22c55e;
}
</style>
