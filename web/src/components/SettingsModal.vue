<script setup lang="ts">
import { DownloadClient, Indexer } from '@/types';
import { ref, watch } from 'vue'

const props = defineProps<{
    isOpen: boolean
    type: 'indexer' | 'download-client'
}>()

const emit = defineEmits<{
    (e: 'close'): void
    (e: 'save', payload: { type: string; preset: string; data: any }): void
}>()

const step = ref<'presets' | 'form'>('presets')
const selectedPreset = ref('')

watch(() => props.isOpen, (newVal) => {
    if (newVal) {
        step.value = 'presets'
        selectedPreset.value = ''
    }
})

const nyaaForm = ref<Indexer>({
    name: 'Nyaa',
    url: 'https://nyaa.si',
    minimum_seeders: 1,
    seed_time: 0,
    enable_rss: true,
    enable_search: true,
    additional_parameters: '',
    priority: 25
})

const qbitForm = ref<DownloadClient>({
    name: 'qBittorrent',
    host: 'localhost',
    port: 8080,
    use_ssl: false,
    username: '',
    password: '',
    category: 'yomarr'
})

function selectPreset(presetName: string) {
    selectedPreset.value = presetName
    step.value = 'form'
}

function backToPresets() {
    step.value = 'presets'
    selectedPreset.value = ''
}

function handleSave() {
    const formData = selectedPreset.value === 'Nyaa' ? nyaaForm.value : qbitForm.value
    emit('save', {
        type: props.type,
        preset: selectedPreset.value,
        data: JSON.parse(JSON.stringify(formData))
    })
    emit('close')
}
</script>

<template>
    <div v-if="isOpen" class="dialog-overlay" @click.self="emit('close')">
        <div class="dialog-window" :class="step === 'presets' ? 'size-lg' : 'size-md'">
            
            <header class="dialog-header">
                <h3>
                    <span v-if="step === 'presets'">Add {{ type === 'indexer' ? 'Indexer' : 'Download Client' }}</span>
                    <span v-else>Add {{ type === 'indexer' ? 'Indexer' : 'Download Client' }} — {{ selectedPreset }}</span>
                </h3>
                <button class="dialog-close-x" @click="emit('close')">×</button>
            </header>

            <div v-if="step === 'presets'" class="dialog-body">
                <div v-if="type === 'indexer'">
                    <h4 class="preset-group-title">Torrents</h4>
                    <div class="preset-selectors">
                        <div class="preset-option-btn" @click="selectPreset('Nyaa')">
                            <h4>Nyaa</h4>
                        </div>
                    </div>
                </div>

                <div v-if="type === 'download-client'">
                    <h4 class="preset-group-title">Torrents</h4>
                    <div class="preset-selectors">
                        <div class="preset-option-btn" @click="selectPreset('qBittorrent')">
                            <h4>qBittorrent</h4>
                        </div>
                    </div>
                </div>
            </div>

            <div v-else-if="step === 'form'" class="dialog-body scrollable">
                <form v-if="selectedPreset === 'Nyaa'" class="sonarr-aligned-form" @submit.prevent="handleSave">
                    <div class="form-row">
                        <div class="label-col"><label>Name</label></div>
                        <div class="input-col"><input type="text" v-model="nyaaForm.name" required /></div>
                    </div>
                    <div class="form-row">
                        <div class="label-col"><label>Base URL</label></div>
                        <div class="input-col"><input type="url" v-model="nyaaForm.url" required /></div>
                    </div>
                    <div class="form-row">
                        <div class="label-col"><label class="text-warn">Additional Parameters</label></div>
                        <div class="input-col">
                            <input type="text" v-model="nyaaForm.additional_parameters" placeholder="&cats=1_0" />
                            <span class="input-explanation">Appends raw argument logic parameters to your structural lookup pipelines.</span>
                        </div>
                    </div>
                    <div class="form-row">
                        <div class="label-col"><label>Minimum Seeders</label></div>
                        <div class="input-col"><input type="number" class="w-sm" v-model.number="nyaaForm.minimum_seeders" min="0" /></div>
                    </div>
                    <div class="form-row">
                        <div class="label-col"><label>Seed Time</label></div>
                        <div class="input-col">
                            <div class="input-with-unit">
                                <input type="number" class="w-sm" v-model.number="nyaaForm.seed_time" min="0" />
                                <span class="unit-label">minutes</span>
                            </div>
                            <span class="input-explanation">The minimum time tracking sessions must remain active before stop routines trigger. Use 0 for infinite seeding.</span>
                        </div>
                    </div>
                    <div class="form-row">
                        <div class="label-col"><label>Priority</label></div>
                        <div class="input-col">
                            <input type="number" class="w-sm" v-model.number="nyaaForm.priority" min="1" max="50" />
                            <span class="input-explanation">Priority ranking for this indexer (1 is highest priority, 50 is lowest). Used to break ties when identical releases are found on multiple trackers.</span>
                        </div>
                    </div>
                    <div class="form-row">
                        <div class="label-col"><label>Enable RSS Sync</label></div>
                        <div class="input-col">
                            <label class="custom-checkbox">
                                <input type="checkbox" v-model="nyaaForm.enable_rss" />
                                <span>Enables automated periodic chronological index pooling entries.</span>
                            </label>
                        </div>
                    </div>
                    <div class="form-row">
                        <div class="label-col"><label>Enable Search</label></div>
                        <div class="input-col">
                            <label class="custom-checkbox">
                                <input type="checkbox" v-model="nyaaForm.enable_search" />
                                <span>Allows direct keyword or ID inquiries during automatic backlog or manual user lookups.</span>
                            </label>
                        </div>
                    </div>
                </form>

                <form v-else-if="selectedPreset === 'qBittorrent'" class="sonarr-aligned-form" @submit.prevent="handleSave">
                    <div class="form-row">
                        <div class="label-col"><label>Name</label></div>
                        <div class="input-col"><input type="text" v-model="qbitForm.name" required /></div>
                    </div>
                    <div class="form-row">
                        <div class="label-col"><label>Host/IP Address</label></div>
                        <div class="input-col"><input type="text" v-model="qbitForm.host" required /></div>
                    </div>
                    <div class="form-row">
                        <div class="label-col"><label>Port</label></div>
                        <div class="input-col"><input type="number" class="w-sm" v-model.number="qbitForm.port" required /></div>
                    </div>
                    <div class="form-row">
                        <div class="label-col"><label>Username</label></div>
                        <div class="input-col"><input type="text" v-model="qbitForm.username" autocomplete="username" /></div>
                    </div>
                    <div class="form-row">
                        <div class="label-col"><label>Password</label></div>
                        <div class="input-col"><input type="password" v-model="qbitForm.password" autocomplete="current-password" placeholder="••••••••" /></div>
                    </div>
                    <div class="form-row">
                        <div class="label-col"><label class="text-warn">Category Namespace</label></div>
                        <div class="input-col">
                            <input type="text" v-model="qbitForm.category" />
                            <span class="input-explanation">Isolates assets cleanly beneath a tracking boundary container inside qBittorrent.</span>
                        </div>
                    </div>
                </form>

                <div v-else class="generic-form-placeholder">
                    <p>Configuration layout architecture for <strong>{{ selectedPreset }}</strong> is primed and awaiting assembly hook mappings.</p>
                </div>
            </div>

            <footer class="dialog-footer">
                <button v-if="step === 'form'" class="footer-btn btn-test" type="button">Test</button>
                <div class="footer-right-cluster">
                    <button v-if="step === 'form'" class="footer-btn btn-back" type="button" @click="backToPresets">Back</button>
                    <button class="footer-btn btn-cancel" type="button" @click="emit('close')">Cancel</button>
                    <button v-if="step === 'form'" class="footer-btn btn-save" type="button" @click="handleSave">Save</button>
                </div>
            </footer>

        </div>
    </div>
</template>

<style scoped>
.dialog-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background-color: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 500;
    padding: 1.5rem;
}

.dialog-window {
    background-color: #1e293b;
    border: 1px solid #334155;
    border-radius: 8px;
    display: flex;
    flex-direction: column;
    max-height: 85vh;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
    overflow: hidden;
    transition: width 0.2s ease;
}

.dialog-window.size-lg { width: 800px; }
.dialog-window.size-md { width: 660px; }

.dialog-header {
    background-color: #111827;
    padding: 1rem 1.5rem;
    border-bottom: 1px solid #334155;
    display: flex;
    align-items: center;
    justify-content: space-between;
}

.dialog-header h3 {
    margin: 0;
    font-size: 1.15rem;
    font-weight: 500;
    color: #f1f5f9;
}

.dialog-close-x {
    background: transparent;
    border: none;
    color: #64748b;
    font-size: 1.5rem;
    cursor: pointer;
    line-height: 1;
}

.dialog-close-x:hover { color: #f1f5f9; }

.dialog-body {
    padding: 1.5rem;
    overflow-y: auto;
    flex: 1;
}

.dialog-body.scrollable {
    padding: 2rem 2.5rem;
}

.preset-group-title {
    font-size: 1.05rem;
    font-weight: 400;
    color: #cbd5e1;
    margin: 0 0 0.75rem 0;
    border-bottom: 1px solid #334155;
    padding-bottom: 0.35rem;
}

.preset-group-title:not(:first-child) {
    margin-top: 1.75rem;
}

.preset-selectors {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(170px, 1fr));
    gap: 0.75rem;
}

.preset-option-btn {
    background-color: #0f172a;
    border: 1px solid #334155;
    border-radius: 4px;
    padding: 1.5rem 1rem;
    text-align: center;
    cursor: pointer;
    transition: all 0.15s ease;
}

.preset-option-btn:hover {
    background-color: #111827;
    border-color: #475569;
}

.preset-option-btn h4 {
    margin: 0;
    font-size: 1rem;
    font-weight: 500;
    color: #ffffff;
}

.sonarr-aligned-form {
    display: flex;
    flex-direction: column;
    gap: 1.2rem;
}

.form-row {
    display: grid;
    grid-template-columns: 180px 1fr;
    gap: 1.5rem;
    align-items: start;
}

.label-col {
    text-align: right;
    padding-top: 0.45rem;
}

.label-col label {
    font-size: 0.9rem;
    font-weight: 600;
    color: #cbd5e1;
}

.label-col label.text-warn {
    color: #f59e0b;
}

.input-col {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
}

.input-col input[type="text"],
.input-col input[type="password"],
.input-col input[type="number"],
.input-col input[type="url"] {
    background-color: #0f172a;
    border: 1px solid #475569;
    border-radius: 4px;
    padding: 0.45rem 0.6rem;
    color: #ffffff;
    font-size: 0.92rem;
    outline: none;
    width: 100%;
    max-width: 420px;
}

.input-col input:focus {
    border-color: #3b82f6;
}

.input-col input.w-sm {
    width: 90px;
}

.input-explanation {
    font-size: 0.8rem;
    color: #64748b;
    line-height: 1.4;
    max-width: 420px;
}

.custom-checkbox {
    display: flex;
    align-items: flex-start;
    gap: 0.5rem;
    cursor: pointer;
    padding-top: 0.45rem;
}

.custom-checkbox input[type="checkbox"] {
    width: 0.95rem;
    height: 0.95rem;
    margin: 0.15rem 0 0 0;
    accent-color: #2563eb;
    flex-shrink: 0;
}

.custom-checkbox span {
    font-size: 0.88rem;
    color: #94a3b8;
    line-height: 1.4;
}

.dialog-footer {
    background-color: #111827;
    padding: 0.85rem 1.5rem;
    border-top: 1px solid #334155;
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.footer-right-cluster {
    display: flex;
    gap: 0.5rem;
    margin-left: auto;
}

.footer-btn {
    border: none;
    border-radius: 4px;
    padding: 0.4rem 1.25rem;
    font-size: 0.88rem;
    font-weight: 600;
    cursor: pointer;
    transition: background-color 0.1s ease;
}

.btn-save { background-color: #2563eb; color: #ffffff; }
.btn-save:hover { background-color: #1d4ed8; }

.btn-test { background-color: #334155; color: #e2e8f0; }
.btn-test:hover { background-color: #475569; }

.btn-cancel { background-color: #1f2937; color: #94a3b8; border: 1px solid #334155; }
.btn-cancel:hover { background-color: #334155; color: #ffffff; }

.btn-back { background-color: #1f2937; color: #e2e8f0; border: 1px solid #334155; }
.btn-back:hover { background-color: #334155; }

.generic-form-placeholder {
    color: #64748b;
    font-size: 0.95rem;
    padding: 2rem 0;
    text-align: center;
}
</style>
