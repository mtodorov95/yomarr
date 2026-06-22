<script setup lang="ts">
import { useActivityStream } from '@/composables/useActivityStream';
import { useToast } from '@/composables/useToast'

const toast = useToast()

const { activeJobs } = useActivityStream((job) => {
    toast.success(`Download Imported Successfully: ${job.series_title}`)
})
</script>

<template>
    <div class="activity-container">
        <div v-if="activeJobs.length === 0" class="empty-state">
            <p class="empty-title">Activity Queue</p>
            <p class="empty-subtitle">No chapters are currently downloading or processing.</p>
        </div>

        <div v-else class="activity-table-wrapper">
            <table class="activity-table">
                <thead>
                    <tr>
                        <th style="width: 55%;">Series</th>
                        <th style="width: 15%;">Status</th>
                        <th style="width: 30%;">Progress</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="job in activeJobs" :key="job.torrent_hash" class="activity-row">
                        <td class="torrent-name-cell">
                            <RouterLink :to="`/series/${job.series_id}`" class="series-title-link">
                                {{ job.series_title }}
                            </RouterLink>

                            <span class="release-context-txt" :title="job.name">
                                <span class="context-label">Target:</span> {{ job.release_detail }}
                            </span>

                            <span class="hash-sub-label">{{ job.torrent_hash }}</span>
                        </td>
                        <td>
                            <span class="status-pill" :class="job.status.toLowerCase()">
                                {{ job.status }}
                            </span>
                        </td>
                        <td>
                            <div class="progress-cell-wrapper">
                                <div class="progress-track-bar">
                                    <div class="progress-fill-bar" :class="{ 'is-imported': job.status === 'Imported' }"
                                        :style="{ width: (job.progress * 100) + '%' }"></div>
                                </div>
                                <span class="progress-percentage-label">
                                    {{ Math.round(job.progress * 100) }}%
                                </span>
                            </div>
                            <div v-if="job.error" class="error-log-subtext">
                                ⚠ Error: {{ job.error }}
                            </div>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>

<style scoped>
@keyframes fadeIn {
    from {
        opacity: 0;
        transform: translateY(2px);
    }

    to {
        opacity: 1;
        transform: translateY(0);
    }
}

.activity-container {
    width: 100%;
    animation: fadeIn 0.15s cubic-bezier(0.16, 1, 0.3, 1);
}

.activity-table-wrapper {
    background-color: #111827;
    border: 1px solid #1f2937;
    border-radius: 0.5rem;
    overflow: hidden;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.3);
}

.activity-table {
    width: 100%;
    border-collapse: collapse;
    text-align: left;
    font-size: 0.85rem;
}

.activity-table th {
    background-color: #1f2937;
    color: #d1d5db;
    padding: 0.75rem 1rem;
    font-weight: 600;
    border-bottom: 1px solid #1f2937;
}

.activity-table td {
    padding: 1rem;
    border-bottom: 1px solid #1f2937;
    color: #e5e7eb;
    vertical-align: top;
}

.activity-row:last-child td {
    border-bottom: none;
}

.torrent-name-cell {
    display: flex;
    flex-direction: column;
    min-width: 0;
}

/* 🚀 Styled Clean Title Link without raw standard anchor underlines */
.series-title-link {
    color: #38bdf8;
    font-size: 1rem;
    font-weight: 600;
    text-decoration: none !important;
    margin-bottom: 0.35rem;
    display: inline-block;
    width: fit-content;
}

.series-title-link:hover {
    text-decoration: underline !important;
    color: #60a5fa;
}

/* 🚀 Demoted Release details helper layout */
.release-context-txt {
    font-weight: 400;
    color: #e5e7eb;
    font-size: 0.825rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 38rem;
    margin-bottom: 0.15rem;
}

.context-label {
    color: #6b7280;
    font-size: 0.75rem;
    text-transform: uppercase;
    font-weight: 700;
    margin-right: 0.25rem;
}

.hash-sub-label {
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    font-size: 0.675rem;
    color: #4b5563;
}

.status-pill {
    font-size: 0.75rem;
    padding: 0.2rem 0.5rem;
    border-radius: 0.25rem;
    font-weight: 600;
    display: inline-block;
}

.status-pill.downloading {
    color: #38bdf8;
    background-color: rgba(56, 189, 248, 0.15);
}

.status-pill.imported {
    color: #34d399;
    background-color: rgba(16, 185, 129, 0.15);
}

.status-pill.failedimport {
    color: #ef4444;
    background-color: rgba(239, 68, 68, 0.15);
}

.progress-cell-wrapper {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    width: 100%;
}

.progress-track-bar {
    flex: 1;
    height: 0.5rem;
    background-color: #374151;
    border-radius: 9999px;
    overflow: hidden;
}

.progress-fill-bar {
    height: 100%;
    background-color: #2563eb;
    border-radius: 9999px;
    transition: width 0.4s ease;
}

.progress-fill-bar.is-imported {
    background-color: #10b981;
}

.progress-percentage-label {
    font-weight: 700;
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    font-size: 0.85rem;
    min-width: 2.5rem;
    text-align: right;
}

.error-log-subtext {
    color: #f87171;
    font-size: 0.75rem;
    margin-top: 0.35rem;
    line-height: 1.4;
}

.empty-state {
    color: #9ca3af;
    text-align: center;
    padding: 5rem 1rem;
    display: flex;
    flex-direction: column;
    align-items: center;
}

.empty-title {
    font-size: 1.25rem;
    font-weight: 700;
    margin: 0 0 0.5rem 0;
    color: #ffffff;
}

.empty-subtitle {
    font-size: 0.875rem;
    color: #9ca3af;
    margin: 0;
}
</style>
