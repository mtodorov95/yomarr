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
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div v-for="s in seriesList" :key="s.id" @click="$emit('select', s)"
            class="p-4 bg-slate-800 hover:bg-slate-800/80 cursor-pointer rounded-xl border border-slate-700 flex flex-col justify-between transition group hover:border-blue-500/40 shadow-md">
            <div>
                <div class="flex justify-between items-start gap-2">
                    <h3 class="font-bold text-blue-400 truncate max-w-[80%]">{{ s.title }}</h3>
                    <button @click="handleDelete(s.id)" class="text-xs text-red-400 hover:text-red-300 transition">
                        Remove
                    </button>
                </div>
                <p class="text-xs text-slate-400 mb-2">Status: {{ s.status }}</p>
            </div>
            <code class="text-xs bg-black p-1 rounded block truncate mt-2">{{ s.path }}</code>
        </div>
    </div>
</template>
