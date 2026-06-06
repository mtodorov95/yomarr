export type SeriesStatus = 'Ongoing' | 'Completed' | 'Unmonitored';

export type ChapterStatus = 'Missing' | 'Downloading' | 'Downloaded' | 'Ignored';

export interface Series {
  id?: number;
  anilist_id?: string | null;
  mangadex_id?: string | null;
  title: string;
  alt_titles: string[];
  status: SeriesStatus;
  path: string;
  localPath?: string;
  total_chapters?: number;
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
