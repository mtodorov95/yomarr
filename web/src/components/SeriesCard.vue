<script setup lang="ts">
import { ref } from 'vue'
import type { Series } from '../types'

const props = defineProps<{
    item: Series
}>()

const emit = defineEmits<{
    (e: 'import', series: Series): void
}>()

const localPath = ref(props.item.localPath || `/Manga/${props.item.title}`)
</script>

<template>
    <div class="card-container">
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
</template>

<style scoped>
.card-container {
    padding: 1rem;
    background-color: #0f172a;
    border-radius: 0.5rem;
    border: 1px solid #334155;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
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
}

.series-status {
    font-size: 0.75rem;
    color: #94a3b8;
    margin: 0;
}

.action-row {
    display: flex;
    gap: 0.5rem;
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
}

.import-button:hover {
    background-color: #22c55e;
}
</style>
