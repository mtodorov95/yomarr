<script setup lang="ts">
import type { Series } from '../types'
import LibrarySeriesCard from './LibrarySeriesCard.vue'

defineProps<{
    seriesList: Series[]
}>()

defineEmits<{
    (e: 'delete', id: number): void
    (e: 'select', series: Series): void
}>()
</script>

<template>
    <div class="arr-series-grid">
        <LibrarySeriesCard 
            v-for="s in seriesList" 
            :key="s.id"
            :series="s"
            @delete="$emit('delete', $event)"
            @select="$emit('select', $event)"
        />
    </div>
</template>

<style scoped>
.arr-series-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: 1.5rem 1.25rem;
    width: 100%;
    box-sizing: border-box;
}
</style>
