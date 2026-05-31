<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { Series, Chapter } from '../types'

const props = defineProps<{
    series: Series
}>()

defineEmits<{
    (e: 'back'): void
}>()

const chapters = ref<Chapter[]>([])
const loading = ref(true)

async function fetchChapters() {
    loading.value = true
    try {
        const res = await fetch(`/api/chapters?series_id=${props.series.id}`)
        if (!res.ok) throw new Error('Failed to fetch chapters')
        chapters.value = await res.json()
    } catch (e) {
        console.error(e)
    } finally {
        loading.value = false
    }
}

onMounted(fetchChapters)
</script>

<template>
    <div class="w-full max-w-4xl animate-fade-in">
        <div class="flex items-center gap-4 mb-6">
            <button @click="$emit('back')"
                class="bg-slate-800 hover:bg-slate-700 border border-slate-700 px-4 py-2 rounded-xl text-sm font-bold transition flex items-center gap-2">
                <span>←</span> Back to Library
            </button>
            <div>
                <h2 class="text-3xl font-black text-blue-400 tracking-tight">{{ series.title }}</h2>
                <p class="text-xs text-slate-400 mt-0.5">
                    Storage Path: <code class="text-slate-300 bg-black/40 px-1.5 py-0.5 rounded font-mono">{{
                        series.path }}</code>
                </p>
            </div>
        </div>

        <div v-if="loading" class="text-slate-500 italic py-8 text-center">
            Querying Chapter Database Manifest...
        </div>

        <div v-else-if="chapters && chapters.length > 0"
            class="bg-slate-800 border border-slate-700 rounded-2xl overflow-hidden shadow-2xl">
            <div class="overflow-x-auto">
                <table class="w-full text-left border-collapse">
                    <thead>
                        <tr
                            class="border-b border-slate-700 bg-slate-900/40 text-slate-400 text-xs uppercase font-black tracking-wider">
                            <th class="p-4">Chapter Number</th>
                            <th class="p-4">Volume</th>
                            <th class="p-4">Tracking Status</th>
                            <th class="p-4">Language</th>
                            <th class="p-4">File Association</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-700/50 text-sm font-medium text-slate-200">
                        <tr v-for="ch in chapters" :key="ch.id" class="hover:bg-slate-700/20 transition">
                            <td class="p-4 font-bold text-slate-100">
                                Chapter {{ ch.number }}
                            </td>
                            <td class="p-4 text-slate-400 font-normal">
                                {{ ch.volume !== null && ch.volume !== undefined ? `Vol. ${ch.volume}` : '—' }}
                            </td>
                            <td class="p-4">
                                <span :class="ch.status === 'Downloaded'
                                    ? 'bg-green-500/10 text-green-400 border-green-500/20'
                                    : 'bg-red-500/10 text-red-400 border-red-500/20'"
                                    class="px-2.5 py-0.5 text-xs font-bold rounded-full border tracking-wide uppercase">
                                    {{ ch.status }}
                                </span>
                            </td>
                            <td class="p-4 text-xs font-mono text-slate-400 max-w-[300px] truncate">
                                {{ ch.language || '—' }}
                            </td>
                            <td class="p-4 text-xs font-mono text-slate-400 max-w-[300px] truncate">
                                {{ ch.file_path || '—' }}
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>

        <div v-else
            class="text-slate-400 text-center py-12 bg-slate-800/40 border border-slate-800 rounded-2xl border-dashed">
            <p class="text-lg font-bold mb-1">Manifest Unpopulated</p>
            <p class="text-xs text-slate-500">MangaDex feed contains no records or synchronization hasn't run yet.</p>
        </div>
    </div>
</template>
