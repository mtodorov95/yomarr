import { ref, onMounted, onUnmounted } from 'vue'
import { QueueEvent } from '../types'

export function useActivityStream(onImportSuccess?: (job: QueueEvent) => void) {
    const activeJobs = ref<QueueEvent[]>([])
    const isConnected = ref(false)
    let eventSource: EventSource | null = null

    function connect() {
        if (eventSource) return

        eventSource = new EventSource('/api/activity')

        eventSource.onopen = () => {
            isConnected.value = true
        }

        eventSource.onerror = () => {
            isConnected.value = false
        }

        eventSource.onmessage = (event) => {
            try {
                const rawPayload: QueueEvent = JSON.parse(event.data)

                if (rawPayload.status === 'Imported') {
                    if (onImportSuccess) onImportSuccess(rawPayload)
                    const targetIdx = activeJobs.value.findIndex(j => j.torrent_hash === rawPayload.torrent_hash)
                    if (targetIdx !== -1) {
                        activeJobs.value[targetIdx] = rawPayload
                    } else {
                        activeJobs.value.push(rawPayload)
                    }

                    setTimeout(() => {
                        activeJobs.value = activeJobs.value.filter(j => j.torrent_hash !== rawPayload.torrent_hash)
                    }, 4000)

                    return
                }

                const targetIdx = activeJobs.value.findIndex(j => j.torrent_hash === rawPayload.torrent_hash)
                if (targetIdx !== -1) {
                    activeJobs.value[targetIdx] = rawPayload
                } else {
                    activeJobs.value.push(rawPayload)
                }
            } catch (err) {
                console.error('[SSE Parse Error] Invalid event frame:', err)
            }
        }
    }

    function disconnect() {
        if (eventSource) {
            eventSource.close()
            eventSource = null
            isConnected.value = false
        }
    }

    onMounted(() => {
        connect()
    })

    onUnmounted(() => {
        disconnect()
    })

    return {
        activeJobs,
        isConnected,
        connect,
        disconnect
    }
}
