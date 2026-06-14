<script setup lang="ts">
import { VolumeCover } from '@/types';
import { getAssetUrl } from '@/utils/utils';
import { computed } from 'vue'

const props = defineProps<{
    covers: VolumeCover[] | undefined
    currentThumbnail: string | undefined;
    seriesPath: string | undefined
}>()

const emit = defineEmits<{
    (e: 'promote', coverFile: string): void
    (e: 'remove', coverFile: string): void
}>()

const hasCovers = computed(() => props.covers && props.covers.length > 0)

const sortedCovers = computed(() => {
    if (!props.covers) return []
    
    return [...props.covers].sort((a, b) => {
        const volA = a.volume === -1 ? -Infinity : a.volume
        const volB = b.volume === -1 ? -Infinity : b.volume
        
        return volB - volA
    })
})
</script>

<template>
    <div class="tab-pane-view">
        <div v-if="hasCovers" class="arr-covers-fluid-grid">
            <div 
                v-for="(cover, _index) in sortedCovers" 
                :key="cover.url" 
                class="arr-cover-archive-card"
                :class="{ 'is-primary-border': cover.url === currentThumbnail }"
            >
                <a 
                    :href="getAssetUrl(cover.url, seriesPath)" 
                    target="_blank" 
                    rel="noopener noreferrer" 
                    class="archive-image-wrapper"
                >
                    <img 
                        :src="getAssetUrl(cover.url, seriesPath)" 
                        alt="Volume Artwork Variant" 
                        class="archive-raw-img" 
                        loading="lazy"
                    />

                    <div class="cover-hover-overlay">
                        <div class="action-buttons-row">
                            <button 
                                v-if="cover.url !== currentThumbnail"
                                @click.stop.prevent="emit('promote', cover.url)"
                                class="overlay-icon-btn promote"
                                title="Set as Main Cover"
                            >
                                ⭐
                            </button>
                            <span v-else class="active-badge-pill" title="Active Main Cover">⭐</span>

                            <button 
                                @click.stop.prevent="emit('remove', cover.url)"
                                :disabled="cover.url === currentThumbnail"
                                class="overlay-icon-btn delete"
                                title="Delete Cover"
                            >
                                🗑️
                            </button>
                        </div>
                    </div>
                </a>
                
                <div class="archive-label-tag">
                    <span v-if="cover.volume === -1">Unassigned Cover</span>
                    <span v-else>Volume {{ cover.volume }}</span>
                    <span v-if="cover.url === currentThumbnail" class="primary-label"> (Active)</span>
                </div>
            </div>
        </div>

        <div v-else class="empty-dashed-box">
            <p class="empty-title">Single Local Specimen</p>
            <p class="empty-subtitle">This media sequence asset currently tracks only its baseline thumbnail identity cover.</p>
        </div>
    </div>
</template>

<style scoped>
.tab-pane-view {
    width: 100%;
}

.arr-covers-fluid-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
    gap: 1.25rem;
}

.arr-cover-archive-card {
    background-color: #111827;
    border: 1px solid #1f2937;
    border-radius: 0.375rem;
    padding: 0.4rem;
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
    transition: transform 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

.arr-cover-archive-card:hover {
    transform: translateY(-4px);
}

.arr-cover-archive-card.is-primary-border {
    border-color: #eab308;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.2);
}

.arr-cover-archive-card:hover.is-primary-border {
    box-shadow: 0 10px 15px -3px rgba(234, 179, 8, 0.2);
}

.archive-image-wrapper {
    display: block;
    width: 100%;
    aspect-ratio: 2 / 3;
    overflow: hidden;
    border-radius: 0.25rem;
    background-color: #030712;
    position: relative;
    cursor: pointer;
}

.archive-raw-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.cover-hover-overlay {
    position: absolute;
    inset: 0;
    background: linear-gradient(to bottom, rgba(15, 23, 42, 0.1), rgba(15, 23, 42, 0.85));
    display: flex;
    flex-direction: column;
    justify-content: flex-end;
    align-items: center;
    padding: 0.5rem;
    opacity: 0;
    pointer-events: none;
    transition: opacity 0.2s ease-in-out;
}

.archive-image-wrapper:hover .cover-hover-overlay {
    opacity: 1;
    pointer-events: auto;
}

.action-buttons-row {
    display: flex;
    width: 100%;
    gap: 0.4rem;
    justify-content: center;
    align-items: center;
}

.overlay-icon-btn {
    flex: 1;
    height: 2.2rem;
    font-size: 0.9rem;
    border: 1px solid transparent;
    border-radius: 0.25rem;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    transition: all 0.15s;
}

.overlay-icon-btn.promote {
    background-color: #1e293b;
    border-color: #451a03;
}

.overlay-icon-btn.promote:hover {
    background-color: #eab308;
    border-color: #facc15;
}

.overlay-icon-btn.delete {
    background-color: #1e293b;
    border-color: #450a0a;
}

.overlay-icon-btn.delete:hover:not(:disabled) {
    background-color: #991b1b;
    border-color: #ef4444;
}

.overlay-icon-btn:disabled {
    background-color: #1f2937;
    border-color: #374151;
    cursor: not-allowed;
    opacity: 0.4;
}

.active-badge-pill {
    flex: 1;
    height: 2.2rem;
    font-size: 0.7rem;
    background-color: rgba(234, 179, 8, 0.1);
    border: 1px solid rgba(234, 179, 8, 0.3);
    color: #eab308;
    border-radius: 0.25rem;
    font-weight: 700;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    backdrop-filter: blur(4px);
}

.archive-label-tag {
    font-size: 0.7rem;
    color: #9ca3af;
    text-align: center;
    font-weight: 600;
    display: flex;
    flex-direction: column;
    gap: 0.1rem;
    padding-top: 0.25rem;
}

.primary-label {
    color: #eab308;
    font-size: 0.65rem;
}

.empty-dashed-box {
    color: #9ca3af;
    text-align: center;
    padding: 3rem 1rem;
    background-color: #111827;
    border: 1px dashed #1f2937;
    border-radius: 0.375rem;
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
</style>
