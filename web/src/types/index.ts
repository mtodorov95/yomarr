export interface Series {
  id?: number;
  anilist_id?: string | null;
  mangadex_id?: string | null;
  title: string;
  status: string;
  path: string;
  localPath?: string;
  total_chapters?: number
}

export interface Chapter {
  id: number
  series_id: number
  number: number
  volume?: number | null
  file_path?: string | null
  status: string
  release_date?: string | null
  language: string
}
