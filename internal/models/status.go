package models

type SeriesStatus string

const (
	SeriesOngoing   SeriesStatus = "Ongoing"
	SeriesHiatus    SeriesStatus = "Hiatus"
	SeriesCompleted SeriesStatus = "Completed"
	SeriesCancelled SeriesStatus = "Cancelled"
	SeriesUnknown   SeriesStatus = "Unknown"
)

type ChapterStatus string

const (
	ChapterMissing     ChapterStatus = "Missing"
	ChapterDownloading ChapterStatus = "Downloading"
	ChapterDownloaded  ChapterStatus = "Downloaded"
	ChapterIgnored     ChapterStatus = "Ignored"
)
