<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import SettingsModal from '../components/SettingsModal.vue'

const route = useRoute()

const activeTab = computed(() => (route.query.tab as string) || 'indexers')

const indexers = ref([
    { id: 1, name: 'Nyaa', protocol: 'Torrent', url: 'https://nyaa.si', enabled: true }
])

const downloadClients = ref([
    { id: 1, name: 'qBittorrent', protocol: 'Torrent', host: 'localhost', enabled: true }
])

const isModalOpen = ref(false)
const modalType = ref<'indexer' | 'download-client'>('indexer')

function openAddModal() {
    modalType.value = activeTab.value === 'indexers' ? 'indexer' : 'download-client'
    isModalOpen.value = true
}

function onSaveConfig(payload: { type: string; preset: string; data: any }) {
    if (payload.type === 'indexer') {
        indexers.value.push({
            id: Date.now(),
            name: payload.data.name,
            protocol: 'Torrent',
            url: payload.data.url || 'Custom API Target',
            enabled: true
        })
    } else {
        downloadClients.value.push({
            id: Date.now(),
            name: payload.data.name,
            protocol: 'Torrent',
            host: payload.data.host || '127.0.0.1',
            enabled: true
        })
    }
}
</script>

<template>
    <div class="settings-view">
        <div v-if="activeTab === 'indexers'" class="tab-content">
            <header class="tab-header">
                <h2>Indexers</h2>
                <p class="tab-subtitle">Configure torrent trackers and Usenet indexers for dynamic content monitoring.</p>
            </header>
            <hr class="section-divider" />

            <div class="cards-grid">
                <div v-for="indexer in indexers" :key="indexer.id" class="entry-card">
                    <div class="card-main">
                        <span class="card-avatar">🌐</span>
                        <div class="card-details">
                            <h3 class="card-title">{{ indexer.name }}</h3>
                            <span class="card-meta">{{ indexer.protocol }} / {{ indexer.url }}</span>
                        </div>
                    </div>
                    <span class="status-indicator enabled">Enabled</span>
                </div>

                <button class="add-entry-card" @click="openAddModal">
                    <span class="add-plus">+</span>
                </button>
            </div>
        </div>

        <div v-if="activeTab === 'download-clients'" class="tab-content">
            <header class="tab-header">
                <h2>Download Clients</h2>
                <p class="tab-subtitle">Configure automated completion managers and backend downloading daemons.</p>
            </header>
            <hr class="section-divider" />

            <div class="cards-grid">
                <div v-for="client in downloadClients" :key="client.id" class="entry-card">
                    <div class="card-main">
                        <span class="card-avatar text-avatar">⚡</span>
                        <div class="card-details">
                            <h3 class="card-title">{{ client.name }}</h3>
                            <span class="card-meta">{{ client.protocol }} / {{ client.host }}</span>
                        </div>
                    </div>
                    <span class="status-indicator enabled">Enabled</span>
                </div>

                <button class="add-entry-card" @click="openAddModal">
                    <span class="add-plus">+</span>
                </button>
            </div>
        </div>

        <div v-if="activeTab !== 'indexers' && activeTab !== 'download-clients'" class="tab-content">
            <header class="tab-header">
                <h2>{{ activeTab.replace('-', ' ') }}</h2>
                <p class="tab-subtitle">Configuration block placeholder.</p>
            </header>
            <hr class="section-divider" />
            <div class="empty-state-notice">This segment layout is ready for your unique local controls.</div>
        </div>

        <SettingsModal 
            :is-open="isModalOpen" 
            :type="modalType" 
            @close="isModalOpen = false" 
            @save="onSaveConfig" 
        />
    </div>
</template>

<style scoped>
.settings-view {
    width: 100%;
    max-width: 1120px;
    margin: 0 auto;
}

.tab-header h2 {
    font-size: 1.65rem;
    font-weight: 400;
    margin: 0 0 0.4rem 0;
    color: #ffffff;
}

.tab-subtitle {
    font-size: 0.92rem;
    color: #94a3b8;
    margin: 0;
}

.section-divider {
    border: none;
    border-top: 1px solid #334155;
    margin: 1.25rem 0 2rem 0;
}

.cards-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: 1.25rem;
}

.entry-card {
    background-color: #1e293b;
    border: 1px solid #334155;
    border-radius: 6px;
    padding: 1.25rem;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    align-items: flex-start;
    min-height: 125px;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}

.card-main {
    display: flex;
    gap: 1rem;
    align-items: flex-start;
    width: 100%;
}

.card-avatar {
    font-size: 1.5rem;
    flex-shrink: 0;
}

.card-avatar.text-avatar {
    background-color: #2563eb;
    color: white;
    width: 28px;
    height: 28px;
    border-radius: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.85rem;
    font-weight: bold;
}

.card-details {
    display: flex;
    flex-direction: column;
    min-width: 0;
}

.card-title {
    font-size: 1.05rem;
    font-weight: 600;
    margin: 0 0 0.25rem 0;
    color: #ffffff;
}

.card-meta {
    font-size: 0.8rem;
    color: #64748b;
    word-break: break-all;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
}

.status-indicator {
    font-size: 0.72rem;
    font-weight: 700;
    padding: 0.15rem 0.45rem;
    border-radius: 3px;
    margin-top: 0.75rem;
}

.status-indicator.enabled {
    background-color: rgba(16, 185, 129, 0.1);
    color: #34d399;
    border: 1px solid rgba(16, 185, 129, 0.2);
}

.add-entry-card {
    background-color: #0f172a;
    border: 2px dashed #334155;
    border-radius: 6px;
    min-height: 125px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.15s ease;
}

.add-entry-card:hover {
    border-color: #3b82f6;
    background-color: rgba(30, 41, 59, 0.5);
}

.add-plus {
    font-size: 2rem;
    color: #475569;
    font-weight: 300;
    transition: color 0.15s ease;
}

.add-entry-card:hover .add-plus {
    color: #3b82f6;
}

.empty-state-notice {
    color: #64748b;
    font-size: 0.95rem;
    padding: 2rem 0;
    text-align: center;
}
</style>
