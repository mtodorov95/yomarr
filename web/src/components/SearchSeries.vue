<script setup lang="ts">
import { ref } from 'vue'
import type { Series } from '../types'
import SeriesCard from './SeriesCard.vue';

const emit = defineEmits<{
    (e: 'import', series: Series): void
}>()

const query = ref('')
const results = ref<Series[]>([])
const searching = ref(false)

async function search() {
    if (!query.value.trim()) return
    searching.value = true
    try {
        const res = await fetch(`/api/series?search=${encodeURIComponent(query.value)}`)
        const data = await res.json()
        results.value = data ? data.map((s: Series) => ({
            ...s,
            localPath: `/mnt/manga/${s.title}`
        })) : []
    } catch (e) {
        console.error(e)
        results.value = []
    }
    finally {
        searching.value = false
    }
}
</script>

<template>
  <div class="search-container">
    <h2 class="search-title">Search AniList</h2>
    <div class="search-box">
      <input 
        v-model="query" 
        @keyup.enter="search"
        type="text" 
        placeholder="Enter manga title..." 
        class="search-input"
      />
      <button 
        @click="search" 
        :disabled="searching" 
        class="search-button"
      >
        {{ searching ? '...' : 'Search' }}
      </button>
    </div>

    <div v-if="results && results.length > 0" class="results-list">
      <SeriesCard
        v-for="item in results" 
        :key="item.anilist_id ?? item.mangadex_id!" 
        :item="item" 
        @import="emit('import', $event)" 
      />
    </div>
    <p v-else-if="results && query" class="no-results-text">No results found.</p>
  </div>
</template>

<style scoped>
.search-container {
    width: 100%;
    max-width: 36rem;
    background-color: #1e293b;
    padding: 1.5rem;
    border-radius: 0.75rem;
    border: 1px solid #334155;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
}

.search-title {
    font-size: 1.25rem;
    font-weight: 700;
    margin-top: 0;
    margin-bottom: 1rem;
}

.search-box {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 1rem;
}

.search-input {
    flex: 1;
    background-color: #0f172a;
    border: 1px solid #475569;
    border-radius: 0.5rem;
    padding: 0.5rem 1rem;
    color: #ffffff;
    font-size: 1rem;
    transition: border-color 0.2s;
}

.search-input:focus {
    outline: none;
    border-color: #3b82f6;
}

.search-button {
    background-color: #2563eb;
    color: #ffffff;
    font-weight: 700;
    padding: 0.5rem 1.5rem;
    border-radius: 0.5rem;
    border: none;
    cursor: pointer;
    transition: background-color 0.2s;
}

.search-button:hover {
    background-color: #3b82f6;
}

.search-button:disabled {
    background-color: #334155;
    color: #94a3b8;
    cursor: not-allowed;
}

.results-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
}

.no-results-text {
    color: #94a3b8;
    font-size: 0.875rem;
    font-style: italic;
    margin: 0;
}
</style>
