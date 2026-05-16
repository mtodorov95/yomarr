<script setup>
import { ref, onMounted } from 'vue'
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
</script>

<template>
    <div class="min-h-screen bg-slate-900 text-white flex flex-col items-center justify-center p-4">
        <h1 class="text-5xl font-black mb-6 tracking-tighter text-blue-400">YOMARR</h1>
        <div class="bg-slate-800 p-6 rounded-xl border border-slate-700 shadow-2xl">
            <p class="text-lg font-medium">
                Backend:
                <span :class="status === 'ok' ? 'text-green-400' : 'text-red-400'" class="uppercase font-bold">
                    {{ status }}
                </span>
            </p>
        </div>
    </div>
</template>
