<script setup lang="ts">
defineProps<{
  scanning: boolean
  loading: boolean
}>()

const emit = defineEmits<{
  (e: 'scan-library'): void
}>()
</script>

<template>
  <header class="action-bar">
    <div class="action-group">
      <button 
        @click="emit('scan-library')" 
        :disabled="scanning || loading" 
        class="action-item-btn primary"
      >
        <span class="btn-icon">🔄</span>
        <span class="btn-text">{{ scanning ? 'Scanning...' : 'Scan Library' }}</span>
      </button>
    </div>
    
  </header>
</template>

<style scoped>
.action-bar {
  background-color: #1e293b;
  border: 1px solid #334155;
  border-radius: 0.5rem;
  padding: 0.5rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}

.action-group {
  display: flex;
  gap: 0.25rem;
}

.action-item-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: transparent;
  color: #cbd5e1;
  border: 1px solid transparent;
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  font-weight: 600;
  border-radius: 0.375rem;
  cursor: pointer;
  transition: all 0.15s ease;
}

.action-item-btn:hover:not(:disabled) {
  background-color: #334155;
  color: #ffffff;
  border-color: #475569;
}

.action-item-btn.primary {
  color: #38bdf8;
}

.action-item-btn.primary:hover:not(:disabled) {
  background-color: #0c4a6e;
  border-color: #0369a1;
}

.action-item-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
</style>
