<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { Series } from '../types'
import SearchSeriesCard from './SearchSeriesCard.vue';
import { useToast } from '@/composables/useToast';

const query = ref('')
const results = ref<Series[]>([])
const libraryIds = ref<Set<string>>(new Set())
const searching = ref(false)
const toast = useToast()

async function fetchLibraryIds() {
    try {
        const res = await fetch('/api/series')
        const data = await res.json()
        if (Array.isArray(data)) {
            const ids = new Set<string>()
            data.forEach((s: Series) => {
                if (s.mangadex_id) ids.add(`md:${s.mangadex_id}`)
                if (s.anilist_id) ids.add(`al:${s.anilist_id}`)
            })
            libraryIds.value = ids
        }
    } catch (e) {
        console.error('Failed fetching library constraints:', e)
        toast.error("Failed fetching library")
    }
}

async function search() {
    if (!query.value.trim()) return
    searching.value = true
    try {
        await fetchLibraryIds()

        const res = await fetch(`/api/series?search=${encodeURIComponent(query.value)}`)
        const data = await res.json()
        results.value = data ? data.map((s: Series) => ({
            ...s,
            localPath: `/Manga/${s.title}`
        })) : []
    } catch (e) {
        console.error(e)
        toast.error("Network error while searching")
        results.value = []
    } finally {
        searching.value = false
    }
}

function isAlreadyInLibrary(item: Series): boolean {
    if (item.mangadex_id && libraryIds.value.has(`md:${item.mangadex_id}`)) return true
    if (item.anilist_id && libraryIds.value.has(`al:${item.anilist_id}`)) return true
    return false
}

async function handleImport({ series, onStart, onEnd }: { series: Series, onStart: () => void, onEnd: () => void }) {
    onStart()
    try {
        const res = await fetch('/api/series', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                ...series,
                path: series.localPath
            })
        })
        if (res.ok) {
            if (series.mangadex_id) libraryIds.value.add(`md:${series.mangadex_id}`)
            if (series.anilist_id) libraryIds.value.add(`al:${series.anilist_id}`)
            toast.success("Successfully imported series")
        } else {
            console.error('Failed importing tracked target entity:', await res.text())
            toast.error("Failed to import series")
        }
    } catch (e) {
        console.error(e)
        toast.error("Network error while trying to import")
    } finally {
        onEnd()
    }
}

onMounted(() => {
    fetchLibraryIds()
})
</script>

<template>
  <div class="search-container">
    <h2 class="search-title">Search for Manga</h2>
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
      <SearchSeriesCard
        v-for="item in results" 
        :key="item.mangadex_id ?? item.anilist_id!" 
        :item="item"
        :in-library="isAlreadyInLibrary(item)"
        @import="handleImport" 
      />
    </div>
    <p v-else-if="results && query" class="no-results-text">No results found.</p>
  </div>
</template>

<style scoped>
.search-container {
    width: 100%;
    background-color: #1e293b;
    padding: 1.5rem;
    border-radius: 0.75rem;
    border: 1px solid #334155;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
    box-sizing: border-box;
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
    width: 100%;
}
.no-results-text {
    color: #94a3b8;
    font-size: 0.875rem;
    font-style: italic;
    margin: 0;
}
</style>
