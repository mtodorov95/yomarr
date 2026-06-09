export function getAssetUrl(filename: string | undefined, seriesPath?: string): string {
    if (!seriesPath || !filename) return ''
    return `/api/assets?path=${encodeURIComponent(seriesPath + '/' + filename)}`
}
