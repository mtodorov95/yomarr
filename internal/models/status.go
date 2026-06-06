package models

type SeriesStatus string

const (
	SeriesOngoing     SeriesStatus = "Ongoing"
	SeriesCompleted   SeriesStatus = "Completed"
	SeriesUnmonitored SeriesStatus = "Unmonitored"
)

type ChapterStatus string

const (
	ChapterMissing     ChapterStatus = "Missing"
	ChapterDownloading ChapterStatus = "Downloading"
	ChapterDownloaded  ChapterStatus = "Downloaded"
	ChapterIgnored     ChapterStatus = "Ignored"
)
