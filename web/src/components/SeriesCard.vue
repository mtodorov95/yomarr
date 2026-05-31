<script setup lang="ts">
import { ref } from 'vue'
import type { Series } from '../types'

const props = defineProps<{
    item: Series
}>()

const emit = defineEmits<{
    (e: 'import', series: Series): void
}>()

const localPath = ref(props.item.localPath || `/mnt/manga/${props.item.title}`)
</script>

<template>
    <div class="p-4 bg-slate-900 rounded-lg border border-slate-700 flex flex-col gap-3">
        <div class="flex justify-between items-start">
            <div>
                <p class="font-bold text-white text-lg">{{ item.title }}</p>
                <p class="text-xs text-slate-400">Status: {{ item.status }}</p>
            </div>
        </div>

        <div class="flex gap-2">
            <input v-model="localPath" type="text" placeholder="Storage path..."
                class="flex-1 text-sm bg-slate-800 border border-slate-600 rounded-md px-3 py-1.5 text-white focus:outline-none focus:border-blue-500" />
            <button @click="emit('import', { ...item, localPath })"
                class="bg-green-600 hover:bg-green-500 text-white text-sm font-bold px-4 py-1.5 rounded-md transition">
                Import
            </button>
        </div>
    </div>
</template>
