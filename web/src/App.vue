<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import type { Series } from './types'
import ToastContainer from './components/ToastContainer.vue'
import { useToast } from './composables/useToast'

const status = ref('loading...')
const toast = useToast()
const checkingHealth = ref(false)
let healthIntervalId: ReturnType<typeof setInterval> | null = null

async function checkHealth() {
    if (checkingHealth.value) return
    checkingHealth.value = true
    try {
        const res = await fetch('/api/health')
        if (!res.ok) throw new Error('fail')
        const data = await res.json()
        status.value = data.status
    } catch (e) {
        status.value = 'offline'
        toast.error("Backend offline")
    } finally {
        checkingHealth.value = false
    }
}

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

onMounted(() => {
    checkHealth()
    
    healthIntervalId = setInterval(checkHealth, 10 * 60 * 1000)
})

onUnmounted(() => {
    if (healthIntervalId) clearInterval(healthIntervalId)
})
</script>

<template>
    <div class="app-layout">
        <input type="checkbox" id="sidebar-toggle" class="sidebar-state-checkbox" />

        <header class="mobile-navbar">
            <label for="sidebar-toggle" class="burger-menu-btn" aria-label="Toggle Navigation Panel">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2.5" stroke="currentColor" class="burger-icon">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" />
                </svg>
            </label>
            <RouterLink to="/" class="brand-logo-link">
                <img src="/logo.png" alt="Yomarr Logo" class="brand-logo-img" />
            </RouterLink>
        </header>

        <aside class="app-sidebar">
            <div class="sidebar-top">
                <RouterLink to="/" class="brand-logo-link sidebar-logo-wrapper">
                    <img src="/logo.png" alt="Yomarr Logo" class="brand-logo-img" />
                </RouterLink>
                <nav class="sidebar-nav">
                    <RouterLink to="/" class="nav-item" active-class="active-nav">
                        <span class="nav-icon">📚</span>
                        <span class="nav-text">Library</span>
                    </RouterLink>
                    <RouterLink to="/add" class="nav-item" active-class="active-nav">
                        <span class="nav-icon">➕</span>
                        <span class="nav-text">Add New</span>
                    </RouterLink>
                    <RouterLink to="/settings" class="nav-item" active-class="active-nav">
                        <span class="nav-icon">⚙️</span>
                        <span class="nav-text">Settings</span>
                    </RouterLink>
                </nav>
            </div>

            <div class="sidebar-footer">
                <div class="system-status">
                    <span class="status-indicator-dot" :class="{ 'is-ok': status === 'ok', 'is-error': status !== 'ok' }"></span>
                    <span class="status-label">System Status: <strong>{{ status }}</strong></span>
                    
                    <button 
                        @click="checkHealth" 
                        class="health-refresh-btn" 
                        :class="{ 'is-spinning': checkingHealth }"
                        :disabled="checkingHealth"
                        title="Force Health Validation Check"
                        aria-label="Refresh status"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2.5" stroke="currentColor" class="refresh-svg-icon">
                            <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99" />
                        </svg>
                    </button>
                </div>
            </div>
        </aside>

        <label for="sidebar-toggle" class="sidebar-overlay-backdrop"></label>

        <main class="main-content">
            <RouterView @import="handleImport" />
        </main>

        <ToastContainer />
    </div>
</template>

<style scoped>
.app-layout {
    display: flex;
    min-height: 100vh;
    background-color: #0f172a;
    color: #ffffff;
    font-family: system-ui, -apple-system, sans-serif;
}

.sidebar-state-checkbox {
    position: absolute;
    opacity: 0;
    pointer-events: none;
    visibility: hidden;
}

.mobile-navbar {
    display: none;
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    height: 3.5rem;
    background-color: #1e293b;
    border-bottom: 1px solid #334155;
    align-items: center;
    padding: 0 1rem;
    gap: 1rem;
    z-index: 40;
}

.burger-menu-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    color: #94a3b8;
    cursor: pointer;
    padding: 0.25rem;
    border-radius: 0.375rem;
    transition: color 0.2s, background-color 0.2s;
}

.burger-menu-btn:hover {
    color: #ffffff;
    background-color: #334155;
}

.burger-icon {
    width: 1.5rem;
    height: 1.5rem;
}

.app-sidebar {
    width: 240px;
    background-color: #1e293b;
    border-right: 1px solid #334155;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    position: fixed;
    top: 0;
    bottom: 0;
    left: 0;
    z-index: 60;
    transition: transform 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

.brand-logo-link {
    display: inline-flex;
    align-items: center;
    text-decoration: none;
    transition: opacity 0.2s ease;
}

.brand-logo-link:hover {
    opacity: 0.85;
}

.brand-logo-img {
    height: 3rem;
    width: auto;
    object-fit: contain;
    border-radius: 3px;
}

.sidebar-logo-wrapper {
    margin: 0 0 2rem 0.5rem;
}

.sidebar-top {
    padding: 1.5rem 1rem;
}

.sidebar-logo {
    font-size: 1.75rem;
    font-weight: 900;
    letter-spacing: -0.05em;
    color: #60a5fa;
    margin: 0 0 2rem 0.5rem;
}

.sidebar-nav {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
}

.nav-item {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.75rem 1rem;
    color: #94a3b8;
    text-decoration: none;
    font-weight: 600;
    font-size: 0.95rem;
    border-radius: 0.375rem;
    transition: all 0.2s;
}

.nav-item:hover {
    color: #ffffff;
    background-color: #334155;
}

.active-nav {
    color: #60a5fa;
    background-color: #0f172a;
    box-shadow: inset 4px 0 0 0 #60a5fa;
}

.sidebar-footer {
    padding: 1rem;
    border-top: 1px solid #334155;
    background-color: #111827;
}

.system-status {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.8rem;
    color: #94a3b8;
}

.status-indicator-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
}

.status-indicator-dot.is-ok {
    background-color: #10b981;
    box-shadow: 0 0 8px #10b981;
}

.status-indicator-dot.is-error {
    background-color: #ef4444;
    box-shadow: 0 0 8px #ef4444;
}

.status-label {
    white-space: nowrap;
}

.status-label strong {
    text-transform: uppercase;
}

.main-content {
    flex: 1;
    margin-left: 240px;
    padding: 2rem 3rem;
    min-width: 0;
    box-sizing: border-box;
}

.sidebar-overlay-backdrop {
    display: none;
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(4px);
    z-index: 50;
    cursor: pointer;
}

.health-refresh-btn {
    background: transparent;
    border: none;
    color: #64748b;
    cursor: pointer;
    padding: 0.2rem;
    border-radius: 4px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s ease;
    margin-left: auto;
}

.health-refresh-btn:hover:not(:disabled) {
    color: #ffffff;
    background-color: #1e293b;
}

.health-refresh-btn:disabled {
    cursor: not-allowed;
    opacity: 0.5;
}

.refresh-svg-icon {
    width: 0.9rem;
    height: 0.9rem;
}

.health-refresh-btn.is-spinning .refresh-svg-icon {
    animation: spin-clockwise 1s linear infinite;
}

@keyframes spin-clockwise {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
}

@media (max-width: 768px) {
    .mobile-navbar {
        display: flex; 
    }

    .app-sidebar {
        transform: translateX(-100%); 
        top: 0;
    }

    .main-content {
        margin-left: 0;
        padding: 5rem 1rem 1.5rem 1rem;
        width: 100%;
    }

    .sidebar-state-checkbox:checked ~ .app-sidebar {
        transform: translateX(0); 
    }

    .sidebar-state-checkbox:checked ~ .sidebar-overlay-backdrop {
        display: block; 
    }
}
</style>
