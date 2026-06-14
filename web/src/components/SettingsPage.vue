<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useToast } from '../composables/useToast'
import { DownloadClient, Indexer } from '@/types'

const toast = useToast()
const currentTab = ref<'indexers' | 'download-clients'>('indexers')
const isSaving = ref(false)

const nyaaConfig = ref<Indexer>({
    name: 'Nyaa',
    url: 'https://nyaa.si',
    priority: 1,
    enable_rss: true,
    enable_search: true,
    additional_parameters: '&cats=3_1&filter=1',
    minimum_seeders: 1,
    seed_time: 0
})

const qbitConfig = ref<DownloadClient>({
    name: 'qBittorrent',
    host: 'localhost',
    port: 8080,
    use_ssl: false,
    username: 'admin',
    password: '',
    category: 'yomarr'
})

async function loadSettings() {
    try {
        const idxRes = await fetch('/api/indexers')
        if (idxRes.ok) {
            const indexers: Indexer[] = await idxRes.json()
            const foundNyaa = indexers.find(i => i.name.toLowerCase() === 'nyaa')
            if (foundNyaa) nyaaConfig.value = foundNyaa
        }

        const dcRes = await fetch('/api/download-clients')
        if (dcRes.ok) {
            const clients: DownloadClient[] = await dcRes.json()
            const foundQbit = clients.find(c => c.name.toLowerCase() === 'qbittorrent')
            if (foundQbit) qbitConfig.value = foundQbit
        }
    } catch (e) {
        console.error(e)
        toast.error('Failed to load system settings profile')
    }
}

async function saveIndexer() {
    isSaving.value = true
    const method = nyaaConfig.value.id ? 'PUT' : 'POST'
    try {
        const res = await fetch('/api/indexers', {
            method,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(nyaaConfig.value)
        })
        if (!res.ok) throw new Error('Save failed')
        const updatedData = await res.json()
        nyaaConfig.value = updatedData
        toast.success('Nyaa configuration saved successfully')
    } catch (e) {
        console.error(e)
        toast.error('Failed to preserve indexer updates')
    } finally {
        isSaving.value = false
    }
}

async function saveDownloadClient() {
    isSaving.value = true
    const method = qbitConfig.value.id ? 'PUT' : 'POST'
    try {
        const res = await fetch('/api/download-clients', {
            method,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(qbitConfig.value)
        })
        if (!res.ok) throw new Error('Save failed')
        const updatedData = await res.json()
        qbitConfig.value = updatedData
        toast.success('qBittorrent configuration saved successfully')
    } catch (e) {
        console.error(e)
        toast.error('Failed to preserve client link parameters')
    } finally {
        isSaving.value = false
    }
}

onMounted(() => {
    loadSettings()
})
</script>
<template>
    <div class="settings-dashboard">
        <header class="settings-header">
            <h1 class="page-title">Settings</h1>
            <p class="page-subtitle">Configure tracker indexers and automatic download dispatches.</p>
        </header>

        <div class="settings-content-container">
            
            <nav class="settings-top-nav">
                <button 
                    class="top-nav-btn" 
                    :class="{ 'is-active': currentTab === 'indexers' }"
                    @click="currentTab = 'indexers'"
                >
                    <span class="btn-icon">🌐</span> Indexers
                </button>
                <button 
                    class="top-nav-btn" 
                    :class="{ 'is-active': currentTab === 'download-clients' }"
                    @click="currentTab = 'download-clients'"
                >
                    <span class="btn-icon">📥</span> Download Clients
                </button>
            </nav>

            <main class="settings-card">
                
                <div v-if="currentTab === 'indexers'" class="settings-card-content">
                    <div class="card-header-row">
                        <img src="https://nyaa.si/static/favicon.png" class="client-avatar-ico" alt="" @error="(e: any) => e.target.style.display='none'"/>
                        <h2>Nyaa Tracker Profile</h2>
                    </div>
                    <hr class="card-divider"/>
                    
                    <form @submit.prevent="saveIndexer" class="form-semantic-flex">
                        
                        <div class="form-row flex-row-align">
                            <div class="form-field f-width-sm">
                                <label>Minimum Seeders</label>
                                <input type="number" v-model.number="nyaaConfig.minimum_seeders" min="0" required />
                            </div>

                            <div class="form-field f-width-sm">
                                <label>Seed Retention Limit (Minutes)</label>
                                <input type="number" v-model.number="nyaaConfig.seed_time" min="0" required />
                                <small class="input-hint hint-center">0 means keep seeding indefinitely</small>
                            </div>

                            <div class="form-field f-width-sm">
                                <label>Query Priority</label>
                                <input type="number" v-model.number="nyaaConfig.priority" required />
                            </div>
                        </div>

                        <div class="form-row flex-row-align f-width-full">
                            
                            <div class="form-col f-width-60">
                                <div class="form-field">
                                    <label>Base URL</label>
                                    <input type="url" v-model="nyaaConfig.url" required placeholder="https://nyaa.si" />
                                </div>
                                <div class="form-field">
                                    <label>Additional URL Parameters</label>
                                    <input type="text" v-model="nyaaConfig.additional_parameters" placeholder="&cats=1_0&filter=1" />
                                </div>
                            </div>

                            <div class="form-col form-checkbox-group f-width-40">
                                <label class="toggle-control-label">
                                    <input type="checkbox" v-model="nyaaConfig.enable_search" />
                                    <span>Enable Automated Dynamic Search Queries</span>
                                </label>
                                <label class="toggle-control-label">
                                    <input type="checkbox" v-model="nyaaConfig.enable_rss" />
                                    <span>Enable Periodic Automated RSS Monitoring</span>
                                </label>
                            </div>
                        </div>
                    </form>
                </div>

                <div v-if="currentTab === 'download-clients'" class="settings-card-content">
                    <div class="card-header-row">
                        <span class="client-avatar-ico text-ico">⚡</span>
                        <h2>qBittorrent Daemon Endpoint</h2>
                    </div>
                    <hr class="card-divider"/>

                    <form @submit.prevent="saveDownloadClient" class="form-semantic-flex">
                        
                        <div class="form-row flex-row-align f-width-full f-wrap-s">
                            <div class="form-field f-width-vsm">
                                <label>Port</label>
                                <input type="number" v-model.number="qbitConfig.port" required placeholder="8080" />
                            </div>
                            
                            <div class="form-field checkbox-inline">
                                <label class="toggle-control-label">
                                    <input type="checkbox" v-model="qbitConfig.use_ssl" />
                                    <span>Use Secure SSL Connection (HTTPS)</span>
                                </label>
                            </div>

                            <div class="form-field f-width-md">
                                <label>WebUI Username</label>
                                <input type="text" v-model="qbitConfig.username" />
                            </div>

                            <div class="form-field f-width-md">
                                <label>WebUI Password</label>
                                <input type="password" v-model="qbitConfig.password" placeholder="••••••••" />
                            </div>
                        </div>

                        <div class="form-row flex-row-align f-width-full f-wrap-xs">
                            <div class="form-field f-width-half">
                                <label>Host/IP Address</label>
                                <input type="text" v-model="qbitConfig.host" required placeholder="localhost" />
                            </div>

                            <div class="form-field f-width-half">
                                <label>Download Manager Routing Category</label>
                                <input type="text" v-model="qbitConfig.category" required placeholder="yomarr" />
                                <small class="input-hint">Saves assets under a dedicated category boundary within qBittorrent</small>
                            </div>
                        </div>
                    </form>
                </div>

            </main>
        </div>

        <div class="floating-submit-control">
            <button 
                class="action-save-btn" 
                :class="{ 'is-loading': isSaving }"
                :disabled="isSaving"
                @click="currentTab === 'indexers' ? saveIndexer() : saveDownloadClient()"
            >
                {{ isSaving ? '⏳ Preserve Setup' : '💾 Save Config' }}
            </button>
        </div>
    </div>
</template>

<style scoped>
.settings-dashboard, 
.settings-dashboard * {
    box-sizing: border-box;
}

.settings-dashboard {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    color: #e2e8f0;
}

.settings-header {
    margin-bottom: 0.25rem;
}

.page-title {
    font-size: 1.85rem;
    font-weight: 800;
    margin: 0;
}

.page-subtitle {
    color: #94a3b8;
    margin: 0.2rem 0 0 0;
    font-size: 0.95rem;
    line-height: 1.4;
}

.settings-content-container {
    width: 100%;
    max-width: 1080px; 
}

.settings-top-nav {
    display: flex;
    gap: 1.25rem;
    border-bottom: 1px solid #334155;
    padding-bottom: 1rem;
    margin-bottom: 1.5rem;
    width: 100%;
}

.top-nav-btn {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.65rem 1.1rem;
    background: transparent;
    border: 1px solid transparent;
    color: #94a3b8;
    font-size: 0.95rem;
    font-weight: 600;
    border-radius: 0.375rem;
    cursor: pointer;
    transition: all 0.15s ease;
    min-height: 40px;
}

.top-nav-btn:hover {
    color: #ffffff;
    background-color: #1e293b;
}

.top-nav-btn.is-active {
    color: #ffffff;
    background-color: #1e293b;
    border-color: #60a5fa;
    box-shadow: 0 0 10px rgba(96, 165, 250, 0.4);
}

.btn-icon {
    font-size: 1.1rem;
}

.settings-card {
    width: 100%;
    background-color: #1e293b;
    border: 1px solid #334155;
    border-radius: 0.6rem;
    box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.2);
    overflow: hidden; 
}

.settings-card-content {
    padding: 1.5rem 2rem;
}

.card-header-row {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 1rem;
}

.card-header-row h2 {
    font-size: 1.2rem;
    margin: 0;
    font-weight: 700;
    color: #ffffff;
}

.client-avatar-ico {
    width: 20px;
    height: 20px;
    object-fit: contain;
    border-radius: 4px;
}

.client-avatar-ico.text-ico {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    background-color: #2563eb;
    color: white;
    font-size: 0.7rem;
    font-weight: bold;
}

.card-divider {
    border: none;
    border-top: 1px solid #334155;
    margin: 0 0 1.5rem 0;
}

.form-semantic-flex {
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
    width: 100%;
}

.form-row {
    display: flex;
    flex-wrap: wrap;
    gap: 1.25rem;
    width: 100%;
}

.flex-row-align {
    align-items: flex-start; 
}

.form-col {
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
}

.form-field {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
}

.form-field.checkbox-inline {
    flex-direction: row;
    align-items: flex-start;
    padding-bottom: 0.5rem;
}

.form-checkbox-group {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    justify-content: center;
    padding-top: 1.5rem;
}

.form-field label {
    font-size: 0.85rem;
    font-weight: 600;
    color: #cbd5e1;
    margin: 0;
}

.input-hint {
    font-size: 0.75rem;
    color: #64748b;
    margin-top: 0.25rem;
    line-height: 1.4;
    display: block;
}

.hint-center {
    text-align: center;
}

.form-field input[type="text"],
.form-field input[type="password"],
.form-field input[type="number"],
.form-field input[type="url"] {
    background-color: #0f172a;
    border: 1px solid #475569;
    border-radius: 0.375rem;
    padding: 0.6rem 0.75rem;
    color: #ffffff;
    font-size: 0.95rem;
    outline: none;
    transition: all 0.15s ease;
    width: 100%;
    min-height: 40px;
}

.form-field input:focus {
    border-color: #60a5fa;
    background-color: #0d121f;
    box-shadow: 0 0 5px rgba(96, 165, 250, 0.2);
}

.toggle-control-label {
    display: flex;
    align-items: flex-start;
    gap: 0.65rem;
    cursor: pointer;
    user-select: none;
}

.toggle-control-label input[type="checkbox"] {
    width: 1.1rem;
    height: 1.1rem;
    accent-color: #2563eb;
    cursor: pointer;
    flex-shrink: 0;
    margin-top: 0.15rem;
}

.toggle-control-label span {
    font-size: 0.9rem !important;
    font-weight: 500 !important;
    color: #f1f5f9;
    line-height: 1.45;
}

.f-width-full { width: 100%; }

.f-width-half { flex: 1 1 calc(50% - 0.75rem); min-width: 250px; }
.f-width-md   { flex: 1 1 calc(25% - 1rem); min-width: 180px; }
.f-width-60   { flex: 1 1 calc(60% - 1.25rem); min-width: 300px; }
.f-width-40   { flex: 1 1 calc(40% - 1.25rem); min-width: 250px; }

.f-width-sm { width: 110px; flex-shrink: 0; }
.f-width-vsm { width: 85px; flex-shrink: 0; }

.floating-submit-control {
    display: flex;
    justify-content: flex-end;
    width: 100%;
    margin-top: -1.25rem;
    transform: translateX(1rem); 
    padding-right: 1rem;
    pointer-events: none;
    z-index: 50;
}

.action-save-btn {
    background-color: #2563eb;
    color: #ffffff;
    border: none;
    border-radius: 99px;
    padding: 0.7rem 1.75rem;
    font-weight: 700;
    font-size: 0.95rem;
    cursor: pointer;
    transition: all 0.2s ease;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.3), 0 2px 4px -1px rgba(0, 0, 0, 0.2);
    pointer-events: auto;
    min-width: 160px;
}

.action-save-btn:hover:not(:disabled) {
    background-color: #1d4ed8;
    transform: translateY(-2px);
}

.action-save-btn.is-loading {
    background-color: #334155;
    color: #94a3b8;
}

.action-save-btn:disabled {
    opacity: 0.7;
    cursor: not-allowed;
}

@media (max-width: 1024px) {
    .settings-sub-sidebar { display: none; }
    .floating-submit-control { transform: none; margin-top: 1.5rem; justify-content: flex-end; padding: 0;}
    .f-width-md { flex: 1 1 100%; min-width: 100%; } 
}

@media (max-width: 800px) {
    .f-wrap-s { flex-direction: column; align-items: flex-start;}
    .f-wrap-xs { flex-direction: column;}
    .f-width-vsm { width: 100%; flex-basis: auto;}
    .form-checkbox-group { padding-top: 0;}
}

@media (max-width: 640px) {
    .settings-dashboard { gap: 1rem; }
    .settings-card-content { padding: 1.25rem 1rem; }
    .form-semantic-flex { gap: 1rem;}
    .form-row { flex-direction: column; gap: 1rem; align-items: stretch; }
    .settings-grid-form { grid-template-columns: 1fr; gap: 1rem;}
    .f-width-half { flex-basis: auto; width: 100%; }
    .f-width-60, .f-width-40 { width: 100%; flex-basis: auto; }
    
    .settings-top-nav { gap: 0.5rem; }
    .top-nav-btn { flex: 1; justify-content: center; padding: 0.5rem; }

    .action-save-btn { width: 100%; padding: 1rem; text-align: center;}
}
</style>
