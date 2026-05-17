<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { Series } from './types'
import SearchSeries from './components/SearchSeries.vue'
const status = ref('loading...')

onMounted(async () => {
    try {
        const res = await fetch('/api/health')
        if (!res.ok) throw new Error('fail')
        const data = await res.json()
        status.value = data.status
    } catch (e) {
        status.value = 'error'
    }
})
async function importSeries(item: Series) {
    try {
        const res = await fetch('/api/series', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                anilist_id: item.anilist_id,
                path: item.localPath
            })
        })
        if (!res.ok) throw new Error('import fail')
        alert(`Imported: ${item.title}`)
    } catch (e) {
        console.error(e)
        alert('Import failed')
    }
}
</script>

<template>
    <div class="min-h-screen bg-slate-900 text-white flex flex-col items-center p-8">
        <h1 class="text-5xl font-black mb-6 tracking-tighter text-blue-400">YOMARR</h1>

        <!-- Health Status -->
        <div class="bg-slate-800 p-4 rounded-xl border border-slate-700 shadow-2xl mb-6 w-full max-w-xl">
            <p class="text-sm font-medium">
                Backend:
                <span :class="status === 'ok' ? 'text-green-400' : 'text-red-400'" class="uppercase font-bold">
                    {{ status }}
                </span>
            </p>
        </div>
        
        <SearchSeries @import="importSeries" />
    </div>
</template>
