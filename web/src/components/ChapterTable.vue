<script setup lang="ts">
import { ref, computed, watch } from 'vue'

interface LanguageVariant {
    id: number
    status: string
    language: string
    file_path?: string
}

interface GroupedChapter {
    number: number
    volume: number | null
    variants: LanguageVariant[]
}

const props = defineProps<{
    chapters: GroupedChapter[]
}>()

const currentPage = ref(1)
const itemsPerPage = ref(50)

const totalPages = computed(() => {
    return Math.ceil(props.chapters.length / itemsPerPage.value) || 1
})

const paginatedChapters = computed(() => {
    const start = (currentPage.value - 1) * itemsPerPage.value
    const end = start + itemsPerPage.value
    return props.chapters.slice(start, end)
})

watch(() => props.chapters, () => {
    currentPage.value = 1
})

function getSortedVariants(variants: LanguageVariant[]) {
    if (!variants) return []
    return [...variants].sort((a, b) => a.language.localeCompare(b.language))
}

function getStatusClass(status: string | undefined): string {
    const s = status?.toLowerCase()
    if (s === 'downloaded') return 'is-downloaded'
    if (s === 'downloading') return 'is-downloading'
    return 'is-missing'
}
</script>

<template>
    <div v-if="chapters && chapters.length > 0" class="arr-table-card">
        <div class="table-container-scroller">
            <table class="arr-manifest-data-table">
                <thead>
                    <tr class="header-labels-row">
                        <th class="column-th text-left">Chapter Number</th>
                        <th class="column-th text-left">Volume</th>
                        <th class="column-th text-left">Available Releases / Language Status</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="ch in paginatedChapters" :key="ch.number" class="body-data-row">
                        <td class="column-td font-bold text-white">
                            Chapter {{ ch.number }}
                        </td>
                        <td class="column-td text-muted">
                            {{ ch.volume !== null && ch.volume !== undefined ? `Vol. ${ch.volume}` : '—' }}
                        </td>
                        <td class="column-td">
                            <div class="badge-flex-layout">
                                <div 
                                    v-for="v in getSortedVariants(ch.variants)" 
                                    :key="v.id"
                                    class="badge-wrapper-node"
                                    :title="v.file_path || 'No local source matched'"
                                >
                                    <span :class="['arr-status-tag', getStatusClass(v.status)]">
                                        <span class="lang-token-prefix">{{ v.language }}:</span> {{ v.status }}
                                    </span>
                                </div>
                            </div>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <div v-if="totalPages > 1" class="arr-pagination-bar">
            <div class="pagination-info">
                Showing {{ (currentPage - 1) * itemsPerPage + 1 }}–{{ Math.min(currentPage * itemsPerPage, chapters.length) }} of {{ chapters.length }} chapters
            </div>
            <div class="pagination-actions">
                <button 
                    @click="currentPage = 1" 
                    :disabled="currentPage === 1"
                    class="page-btn text-btn"
                >
                    « First
                </button>
                <button 
                    @click="currentPage--" 
                    :disabled="currentPage === 1"
                    class="page-btn"
                >
                    ◀ Previous
                </button>
                
                <span class="page-indicator">
                    Page <strong>{{ currentPage }}</strong> of {{ totalPages }}
                </span>

                <button 
                    @click="currentPage++" 
                    :disabled="currentPage === totalPages"
                    class="page-btn"
                >
                    Next ▶
                </button>
                <button 
                    @click="currentPage = totalPages" 
                    :disabled="currentPage === totalPages"
                    class="page-btn text-btn"
                >
                    Last »
                </button>
            </div>
        </div>
    </div>

    <div v-else class="empty-dashed-box">
        <p class="empty-title">Manifest Index Unpopulated</p>
        <p class="empty-subtitle">Local folder storage matching or system synchronizer hasn't parsed this workspace path target yet.</p>
    </div>
</template>

<style scoped>
.arr-table-card {
    background-color: #111827;
    border: 1px solid #1f2937;
    border-radius: 0.375rem;
    overflow: hidden;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}

.table-container-scroller {
    width: 100%;
    overflow-x: auto;
}

.arr-manifest-data-table {
    width: 100%;
    text-align: left;
    border-collapse: collapse;
}

.header-labels-row {
    border-bottom: 1px solid #1f2937;
    background-color: #111827;
    color: #9ca3af;
    font-size: 0.75rem;
    text-transform: uppercase;
    font-weight: 700;
    letter-spacing: 0.025em;
}

.column-th {
    padding: 0.75rem 1rem;
}

.body-data-row {
    border-bottom: 1px solid #1f2937;
    color: #d1d5db;
    font-size: 0.85rem;
    transition: background-color 0.1s;
}

.body-data-row:hover {
    background-color: #1f2937;
}

.column-td {
    padding: 0.75rem 1rem;
}

.text-left { text-align: left; }
.font-bold { font-weight: 700; }
.text-white { color: #ffffff; }
.text-muted { color: #9ca3af; }

.badge-flex-layout {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4rem;
}

.badge-wrapper-node {
    display: inline-flex;
}

.arr-status-tag {
    padding: 0.2rem 0.5rem;
    font-size: 0.7rem;
    font-weight: 700;
    border-radius: 0.25rem;
    text-transform: uppercase;
    display: inline-flex;
    align-items: center;
    gap: 0.2rem;
}

.lang-token-prefix {
    opacity: 0.75;
}

.arr-status-tag.is-downloaded {
    background-color: rgba(16, 185, 129, 0.15);
    color: #10b981;
    border: 1px solid rgba(16, 185, 129, 0.25);
}

.arr-status-tag.is-missing {
    background-color: rgba(239, 68, 68, 0.1);
    color: #f87171;
    border: 1px solid rgba(239, 68, 68, 0.15);
}

.arr-status-tag.is-downloading {
    background-color: rgba(245, 158, 11, 0.15); 
    color: #f59e0b;              
    border: 1px solid rgba(245, 158, 11, 0.25);
}

.empty-dashed-box {
    color: #9ca3af;
    text-align: center;
    padding: 3rem 1rem;
    background-color: #111827;
    border: 1px dashed #1f2937;
    border-radius: 0.375rem;
    width: 100%;
}

.empty-title {
    font-size: 1rem;
    font-weight: 700;
    margin: 0 0 0.25rem 0;
    color: #ffffff;
}

.empty-subtitle {
    font-size: 0.75rem;
    color: #9ca3af;
    margin: 0;
}

.arr-pagination-bar {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    align-items: center;
    justify-content: space-between;
    padding: 0.75rem 1rem;
    background-color: #111827;
    border-top: 1px solid #1f2937;
    font-size: 0.8rem;
    color: #9ca3af;
}

@media (min-width: 640px) {
    .arr-pagination-bar {
        flex-direction: row;
    }
}

.pagination-actions {
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.page-btn {
    background-color: #1f2937;
    border: 1px solid #374151;
    color: #f3f4f6;
    padding: 0.35rem 0.75rem;
    border-radius: 0.25rem;
    font-weight: 600;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    transition: all 0.1s ease;
}

.page-btn:hover:not(:disabled) {
    background-color: #374151;
    color: #ffffff;
    border-color: #4b5563;
}

.page-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
}

.page-btn.text-btn {
    background: transparent;
    border-color: transparent;
    color: #9ca3af;
}

.page-btn.text-btn:hover:not(:disabled) {
    color: #ffffff;
}

.page-indicator {
    padding: 0 0.5rem;
    color: #9ca3af;
}

.page-indicator strong {
    color: #38bdf8;
}
</style>
