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
</script>

<template>
    <div class="series-grid">
        <div 
            v-for="s in seriesList" 
            :key="s.id" 
            @click="$emit('select', s)"
            class="series-item-card"
        >
            <div>
                <div class="card-top-row">
                    <h3 class="series-card-title">{{ s.title }}</h3>
                    <button 
                        @click.stop="handleDelete(s.id)" 
                        class="remove-button"
                    >
                        Remove
                    </button>
                </div>
                <p class="series-card-status">Status: {{ s.status }}</p>
            </div>
            <code class="series-card-path">{{ s.path }}</code>
        </div>
    </div>
</template>

<style scoped>
.series-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: 1rem;
}

@media (min-width: 768px) {
    .series-grid {
        grid-template-columns: repeat(2, 1fr);
    }
}

.series-item-card {
    padding: 1rem;
    background-color: #1e293b;
    cursor: pointer;
    border-radius: 0.75rem;
    border: 1px solid #334155;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    transition: background-color 0.2s, border-color 0.2s;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
}

.series-item-card:hover {
    background-color: rgba(30, 41, 59, 0.8);
    border-color: rgba(96, 165, 250, 0.4);
}

.card-top-row {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 0.5rem;
}

.series-card-title {
    font-weight: 700;
    color: #60a5fa;
    margin: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 80%;
}

.remove-button {
    font-size: 0.75rem;
    color: #f87171;
    background: none;
    border: none;
    cursor: pointer;
    padding: 0;
    transition: color 0.2s;
}

.remove-button:hover {
    color: #fca5a5;
}

.series-card-status {
    font-size: 0.75rem;
    color: #94a3b8;
    margin-top: 0;
    margin-bottom: 0.5rem;
}

.series-card-path {
    font-size: 0.75rem;
    background-color: #000000;
    padding: 0.25rem;
    border-radius: 0.25rem;
    display: block;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    margin-top: 0.5rem;
}
</style>
