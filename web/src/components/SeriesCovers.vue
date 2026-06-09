<script setup lang="ts">
import { getAssetUrl } from '@/utils/utils';
import { computed } from 'vue'

const props = defineProps<{
    covers: string[] | undefined
    seriesPath: string | undefined
}>()

const hasCovers = computed(() => props.covers && props.covers.length > 0)
</script>

<template>
    <div class="tab-pane-view">
        <div v-if="hasCovers" class="arr-covers-fluid-grid">
            <div 
                v-for="(coverFile, index) in covers" 
                :key="index" 
                class="arr-cover-archive-card"
            >
            <a 
                    :href="getAssetUrl(coverFile, seriesPath)" 
                    target="_blank" 
                    rel="noopener noreferrer" 
                    class="archive-image-link"
                >
                <div class="archive-image-wrapper">
                    <img 
                        :src="getAssetUrl(coverFile, seriesPath)" 
                        alt="Volume Artwork Variant" 
                        class="archive-raw-img" 
                        loading="lazy"
                    />
                </div>
            </a>
                <div class="archive-label-tag">
                    Variant {{ index + 1 }}
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
    grid-template-columns: repeat(auto-fill, minmax(130px, 1fr));
    gap: 1.25rem;
}

.arr-cover-archive-card {
    background-color: #111827;
    border: 1px solid #1f2937;
    border-radius: 0.25rem;
    padding: 0.4rem;
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
    transition: transform 0.2s ease, border-color 0.2s ease;
}

.arr-cover-archive-card:hover {
    transform: translateY(-2px);
}

.archive-image-link {
    display: block;
    width: 100%;
    cursor: pointer;
}

.archive-image-wrapper {
    width: 100%;
    aspect-ratio: 2 / 3;
    overflow: hidden;
    border-radius: 0.125rem;
    background-color: #030712;
    transition: opacity 0.2s ease;
}

.archive-image-link:hover .archive-image-wrapper {
    opacity: 0.85;
}

.archive-raw-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.archive-label-tag {
    font-size: 0.7rem;
    color: #9ca3af;
    text-align: center;
    font-weight: 600;
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
