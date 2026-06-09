<script setup lang="ts">
import { useToast } from '../composables/useToast'

const { toasts, remove } = useToast()
</script>

<template>
  <div class="toast-container">
    <TransitionGroup name="toast-slide">
      <div 
        v-for="toast in toasts" 
        :key="toast.id" 
        :class="['toast-item', `toast-${toast.type}`]"
        @click.stop="remove(toast.id)"
      >
        <span class="toast-icon">
          <span v-if="toast.type === 'success'">✓</span>
          <span v-else-if="toast.type === 'error'">✕</span>
          <span v-else>ℹ</span>
        </span>
        <p class="toast-message">{{ toast.message }}</p>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.toast-container {
  position: fixed;
  bottom: 3.5rem;
  left: 0.5rem;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  width: 100%;
  max-width: 18rem;
  pointer-events: none;
}

.toast-item {
  pointer-events: auto;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem;
  border-radius: 0.5rem;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.3), 0 4px 6px -4px rgba(0, 0, 0, 0.3);
  cursor: pointer;
  border: 1px solid transparent;
  color: #ffffff;
  font-family: system-ui, -apple-system, sans-serif;
  font-size: 0.875rem;
  font-weight: 500;
  user-select: none;
}

.toast-success {
  background-color: #064e3b;
  border-color: #059669;
}

.toast-error {
  background-color: #7f1d1d;
  border-color: #dc2626;
}

.toast-info {
  background-color: #1e3a8a;
  border-color: #2563eb;
}

.toast-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 1.25rem;
  height: 1.25rem;
  border-radius: 9999px;
  font-weight: 700;
  font-size: 0.75rem;
  background-color: rgba(255, 255, 255, 0.2);
  pointer-events: none;
  flex-shrink: 0;
}

.toast-message {
  margin: 0;
  flex: 1;
  pointer-events: none;
  word-break: break-word;
}

.toast-slide-enter-from {
  opacity: 0;
  transform: translateX(-2rem);
}

.toast-slide-leave-to {
  opacity: 0;
  transform: scale(0.9);
}

.toast-slide-enter-active,
.toast-slide-leave-active {
  transition: all 0.25s ease;
}
</style>
