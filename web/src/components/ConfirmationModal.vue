<script setup lang="ts">
defineProps<{
    isOpen: boolean
    title: string
    message: string
}>()

defineEmits<{
    (e: 'close'): void
    (e: 'confirm'): void
}>()
</script>

<template>
    <Transition name="modal-fade">
        <div v-if="isOpen" class="modal-overlay" @click.self="$emit('close')">
            <div class="modal-card">
                <h3 class="modal-title">{{ title }}</h3>
                <p class="modal-message">{{ message }}</p>
                <div class="modal-actions">
                    <button class="btn-cancel" @click="$emit('close')">
                        Cancel
                    </button>
                    <button class="btn-confirm" @click="$emit('confirm')">
                        Confirm
                    </button>
                </div>
            </div>
        </div>
    </Transition>
</template>

<style scoped>
.modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background-color: rgba(15, 23, 42, 0.75);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 999;
}

.modal-card {
    background-color: #1e293b;
    border: 1px solid #334155;
    border-radius: 0.75rem;
    padding: 1.5rem;
    width: 100%;
    max-width: 28rem;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
    font-family: system-ui, -apple-system, sans-serif;
}

.modal-title {
    margin-top: 0;
    margin-bottom: 0.75rem;
    font-size: 1.25rem;
    font-weight: 700;
    color: #ffffff;
}

.modal-message {
    margin-top: 0;
    margin-bottom: 1.5rem;
    font-size: 0.9375rem;
    color: #94a3b8;
    line-height: 1.5;
}

.modal-actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.75rem;
}

.btn-cancel {
    background-color: #334155;
    border: 1px solid #475569;
    color: #f8fafc;
    font-weight: 600;
    font-size: 0.875rem;
    padding: 0.5rem 1rem;
    border-radius: 0.5rem;
    cursor: pointer;
    transition: background-color 0.2s;
}

.btn-cancel:hover {
    background-color: #475569;
}

.btn-confirm {
    background-color: #dc2626;
    border: none;
    color: #ffffff;
    font-weight: 600;
    font-size: 0.875rem;
    padding: 0.5rem 1rem;
    border-radius: 0.5rem;
    cursor: pointer;
    transition: background-color 0.2s;
}

.btn-confirm:hover {
    background-color: #ef4444;
}

.modal-fade-enter-from,
.modal-fade-leave-to {
    opacity: 0;
}

.modal-fade-enter-from .modal-card {
    transform: scale(0.95);
}

.modal-fade-leave-to .modal-card {
    transform: scale(0.95);
}

.modal-fade-enter-active,
.modal-fade-leave-active {
    transition: opacity 0.2s ease;
}

.modal-fade-enter-active .modal-card,
.modal-fade-leave-active .modal-card {
    transition: transform 0.2s ease;
}
</style>
