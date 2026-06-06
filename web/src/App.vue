<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { Series } from './types'
import ToastContainer from './components/ToastContainer.vue'
import { useToast } from './composables/useToast'

const status = ref('loading...')
const toast = useToast()

onMounted(async () => {
    try {
        const res = await fetch('/api/health')
        if (!res.ok) throw new Error('fail')
        const data = await res.json()
        status.value = data.status
    } catch (e) {
        status.value = 'error'
        toast.error("Backend offline")
    }
})

async function handleImport(item: Series) {
    try {
        const res = await fetch('/api/series', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                mangadex_id: item.mangadex_id || null,  
                anilist_id: item.anilist_id || null,     
                title: item.title,                  
                alt_titles: item.alt_titles || [],
                status: item.status || 'Ongoing',     
                path: item.localPath,
                total_chapters: item.total_chapters || 0
            })
        })
        if (!res.ok) throw new Error('import fail')
        toast.success(`Imported: ${item.title}`);
    } catch (e) {
        console.error(e)
        toast.error(`Import failed for ${item.title}`)
    }
}
</script>
<template>
    <div class="app-container">
        <h1 class="app-logo">YOMARR</h1>

        <div class="health-card">
            <p class="health-text">
                Backend:
                <span :class="['status-badge', status === 'ok' ? 'status-ok' : 'status-error']">
                    {{ status }}
                </span>
            </p>
        </div>

        <div class="nav-tabs">
            <RouterLink 
                to="/" 
                class="tab-button" 
                active-class="active-tab"
            >
                Library
            </RouterLink>
            <RouterLink 
                to="/add" 
                class="tab-button" 
                active-class="active-tab"
            >
                Add New
            </RouterLink>
        </div>

        <RouterView @import="handleImport" />

        <ToastContainer />
    </div>
</template>

<style scoped>
.app-container {
    min-height: 100vh;
    background-color: #0f172a;
    color: #ffffff;
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 2rem;
    font-family: system-ui, -apple-system, sans-serif;
}

.app-logo {
    font-size: 3rem;
    font-weight: 900;
    margin-bottom: 1.5rem;
    letter-spacing: -0.05em;
    color: #60a5fa;
}

.health-card {
    background-color: #1e293b;
    padding: 1rem;
    border-radius: 0.75rem;
    border: 1px solid #334155;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
    margin-bottom: 1.5rem;
    width: 100%;
    max-width: 36rem;
}

.health-text {
    font-size: 0.875rem;
    font-weight: 500;
    margin: 0;
}

.status-badge {
    text-transform: uppercase;
    font-weight: 700;
}

.status-ok {
    color: #34d399;
}

.status-error {
    color: #f87171;
}

.nav-tabs {
    display: flex;
    gap: 1rem;
    margin-bottom: 2rem;
}

.tab-button {
    background: none;
    border: none;
    color: #94a3b8;
    padding-bottom: 0.25rem;
    font-weight: 700;
    cursor: pointer;
    font-size: 1rem;
    transition: color 0.2s, border-color 0.2s;
    border-bottom: 2px solid transparent;
}

.tab-button:hover {
    color: #ffffff;
}

.active-tab {
    color: #60a5fa;
    border-bottom: 2px solid #60a5fa;
}
</style>
