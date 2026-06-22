export type SeriesStatus = 'Ongoing' | 'Completed' | 'Unmonitored';

export type ChapterStatus = 'Missing' | 'Downloading' | 'Downloaded' | 'Ignored';

export interface Series {
    id?: number;
    anilist_id?: string | null;
    mangadex_id?: string | null;
    title: string;
    alt_titles: Record<string, string[]>;
    status: SeriesStatus;
    path: string;
    localPath?: string;
    total_chapters?: number;
    downloaded_count?: number;
    thumbnail?: string;
    historical_covers: VolumeCover[];
    author?: string | null;
    artist?: string | null;
    genres: string[];
    description?: string | null;
}

export interface Chapter {
    id: number;
    series_id: number;
    number: number;
    volume?: number | null;
    file_path?: string | null;
    status: ChapterStatus;
    release_date?: string | null;
    language: string;
}

export interface VolumeCover {
    volume: number;
    url: string;
}

export interface TorrentVariant {
    title: string
    link: string
    seeders: number
    leechers: number
    size: string
    info_hash: string
}

export interface ChapterGroup {
    chapter_number: number
    volume?: number
    english: TorrentVariant[]
    raws: TorrentVariant[]
}

export interface SystemStats {
    total_series: number;
    downloaded_chapters: number;
    missing_chapters: number;
    size_on_disk_bytes: number;
}

export interface Indexer {
    id?: number
    name: string
    url: string
    api_key?: string
    priority: number
    enable_rss: boolean
    enable_search: boolean
    additional_parameters: string
    minimum_seeders: number
    seed_time: number
}

export interface DownloadClient {
    id?: number
    name: string
    host: string
    port: number
    use_ssl: boolean
    username?: string
    password?: string
    category: string
}

export interface QueueEvent {
    torrent_hash: string
    status: string
    progress: number
    name: string
    error?: string
    series_id: number
    release_detail: string
    series_title: string
}

