<script setup lang="ts">
import { computed, ref } from 'vue'
import type { Series } from '../types'

const props = defineProps<{
    item: Series
    inLibrary: boolean
}>()

const emit = defineEmits<{
    (e: 'import', payload: { series: Series, onStart: () => void, onEnd: () => void }): void
}>()

const localPath = ref(props.item.localPath || `/Manga/${props.item.title}`)
const importing = ref(false)

const coverSrc = computed(() => {
    if (!props.item.thumbnail) return ''
    return `/api/proxy-cover?url=${encodeURIComponent(props.item.thumbnail)}`
})

function triggerImport() {
    if (props.inLibrary || importing.value) return
    
    emit('import', {
        series: { ...props.item, localPath: localPath.value },
        onStart: () => { importing.value = true },
        onEnd: () => { importing.value = false }
    })
}
</script>

<template>
    <div class="card-container" :class="{ 'in-library-card': inLibrary }">
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
                        <p class="series-status">Status: <span class="status-badge">{{ item.status }}</span></p>
                    </div>
                    
                    <div v-if="inLibrary" class="library-badge">
                        In Library
                    </div>
                </div>

                <div v-if="!inLibrary" class="action-row">
                    <input 
                        v-model="localPath" 
                        :disabled="importing"
                        type="text" 
                        placeholder="Storage path..."
                        class="path-input" 
                    />
                    <button 
                        @click="triggerImport"
                        :disabled="importing"
                        class="import-button"
                    >
                        {{ importing ? 'Importing...' : 'Import' }}
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
    transition: border-color 0.2s, background-color 0.2s;
    width: 100%;
    box-sizing: border-box;
}

.in-library-card {
    border-color: #1e293b;
    background-color: #111827;
    opacity: 0.85;
}

.card-content {
    display: flex;
    gap: 1rem;
    align-items: center;
    width: 100%;
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
    gap: 1rem;
    width: 100%;
}

.meta-section {
    display: flex;
    flex-direction: column;
    min-width: 0;
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

.status-badge {
    color: #cbd5e1;
    font-weight: 500;
}

.library-badge {
    background-color: #1e293b;
    color: #3b82f6;
    border: 1px solid #2563eb;
    font-size: 0.75rem;
    font-weight: 700;
    padding: 0.25rem 0.75rem;
    border-radius: 2rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    white-space: nowrap;
}

.action-row {
    display: flex;
    gap: 0.5rem;
    align-items: center;
    width: 100%;
}

.path-input {
    flex: 1;
    font-size: 0.875rem;
    background-color: #1e293b;
    border: 1px solid #475569;
    border-radius: 0.375rem;
    padding: 0.5rem 0.75rem;
    color: #ffffff;
    transition: border-color 0.2s, opacity 0.2s;
    min-width: 0;
}

.path-input:focus {
    outline: none;
    border-color: #3b82f6;
}

.path-input:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

.import-button {
    background-color: #16a34a;
    color: #ffffff;
    font-size: 0.875rem;
    font-weight: 700;
    padding: 0.5rem 1.25rem;
    border-radius: 0.375rem;
    border: none;
    cursor: pointer;
    transition: background-color 0.2s, opacity 0.2s;
    white-space: nowrap;
    min-width: 6.5rem;
}

.import-button:hover:not(:disabled) {
    background-color: #22c55e;
}

.import-button:disabled {
    background-color: #334155;
    color: #94a3b8;
    cursor: not-allowed;
}
</style>
