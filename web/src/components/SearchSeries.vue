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
  <div class="w-full max-w-xl bg-slate-800 p-6 rounded-xl border border-slate-700 shadow-2xl">
    <h2 class="text-xl font-bold mb-4">Search AniList</h2>
    <div class="flex gap-2 mb-4">
      <input 
        v-model="query" 
        @keyup.enter="search"
        type="text" 
        placeholder="Enter manga title..." 
        class="flex-1 bg-slate-900 border border-slate-600 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-blue-500"
      />
      <button 
        @click="search" 
        :disabled="searching" 
        class="bg-blue-600 hover:bg-blue-500 disabled:bg-slate-700 text-white font-bold px-6 py-2 rounded-lg transition"
      >
        {{ searching ? '...' : 'Search' }}
      </button>
    </div>

    <!-- Results using SeriesCard -->
    <div v-if="results && results.length > 0" class="space-y-3">
      <SeriesCard
        v-for="item in results" 
        :key="item.anilist_id ?? item.mangadex_id!" 
        :item="item" 
        @import="emit('import', $event)" 
      />
    </div>
    <p v-else-if="results && query" class="text-slate-400 text-sm italic">No results found.</p>
  </div>
</template>
