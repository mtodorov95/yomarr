import { ref } from "vue";

export type ToastType = "success" | "error" | "info";

export interface Toast {
    id: string
    message: string
    type: ToastType
    duration?: number
}

const DEFAULT_DURATION: number = 3000

const toasts = ref<Toast[]>([])

export function useToast() {
    function removeToast(id: string) {
        toasts.value = toasts.value.filter((t) => t.id !== id)
    }

    function addToast(message: string, type: ToastType = "info", duration: number = DEFAULT_DURATION) {
        const id = crypto.randomUUID()
        const toast: Toast = { id, message, type, duration }

        toasts.value.push(toast)

        if (duration > 0) {
            setTimeout(() => {
                removeToast(id)
            }, duration)
        }
    }

    return {
        toasts: toasts,
        success: (msg: string, dur?: number) => addToast(msg, 'success', dur),
        error: (msg: string, dur?: number) => addToast(msg, 'error', dur),
        info: (msg: string, dur?: number) => addToast(msg, 'info', dur),
        remove: removeToast
    }
}
