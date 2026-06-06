package indexer

import (
	"regexp"
	"strconv"
)

var (
	// Matches strict light novel indicators to block accidental downloads
	// Matches "[LN]", "(LN)", "~ LN ~", "Light Novel", "Novel v01"
	lnBlacklistRegex = regexp.MustCompile(`(?i)\b(ln|light\s*novel|novel)\b`)

	// Matches and targets year ranges to erase them: "(2023-2025)", "[2023-2025]"
	yearRangeRegex = regexp.MustCompile(`(?i)[([ ]\d{4}\s*-\s*\d{4}[)\] ]`)

	// Matches specific individual chapters: "Ch. 10", "- 10"
	chSingleRegex = regexp.MustCompile(`(?i)(?:-\s+|ch(?:apter)?\s+)(\d+(?:\.\d+)?)`)
	
	// Matches ranges: "Ch 01-12", "05-15"
	chRangeRegex  = regexp.MustCompile(`(?i)(?:ch(?:apter)?\s+)?(\d+)\s*-\s*(\d+)`)
	
	// Matches volumes: "Vol.02", "v03", "Volume 4", "v01-05"
	volRegex      = regexp.MustCompile(`(?i)(?:vol(?:ume)?\.?\s*|v)(\d+)(?:\s*-\s*(\d+))?`)

	// Matches Japanese volume syntax variants like "第02巻", "第 2 巻", "第02-04巻"
	volJaRegex = regexp.MustCompile(`(?i)第\s*(\d+)\s*(?:-\s*(\d+)\s*)?巻`)
)

type ReleaseType string
const (
	TypeSingle  ReleaseType = "single"
	TypeRange   ReleaseType = "range"
	TypeVolume  ReleaseType = "volume"
)

type ParsedRelease struct {
	Type     ReleaseType
	StartNum float64
	EndNum   float64 
}

func ParseTorrentTitle(title string) (ParsedRelease, bool) {
	if lnBlacklistRegex.MatchString(title) {
		return ParsedRelease{}, false
	}

	cleanedTitle := yearRangeRegex.ReplaceAllString(title, " ")

	if jaVolMatches := volJaRegex.FindStringSubmatch(cleanedTitle); len(jaVolMatches) > 1 {
		start, _ := strconv.ParseFloat(jaVolMatches[1], 64)
		if jaVolMatches[2] != "" {
			end, _ := strconv.ParseFloat(jaVolMatches[2], 64)
			return ParsedRelease{Type: TypeVolume, StartNum: start, EndNum: end}, true
		}
		return ParsedRelease{Type: TypeVolume, StartNum: start, EndNum: start}, true
	}

	if volMatches := volRegex.FindStringSubmatch(cleanedTitle); len(volMatches) > 1 {
		start, _ := strconv.ParseFloat(volMatches[1], 64)
		if volMatches[2] != "" {
			end, _ := strconv.ParseFloat(volMatches[2], 64)
			return ParsedRelease{Type: TypeVolume, StartNum: start, EndNum: end}, true
		}
		return ParsedRelease{Type: TypeVolume, StartNum: start, EndNum: start}, true
	}

	if rangeMatches := chRangeRegex.FindStringSubmatch(cleanedTitle); len(rangeMatches) > 2 {
		start, _ := strconv.ParseFloat(rangeMatches[1], 64)
		end, _ := strconv.ParseFloat(rangeMatches[2], 64)
		if end > start {
			return ParsedRelease{Type: TypeRange, StartNum: start, EndNum: end}, true
		}
	}

	if singleMatches := chSingleRegex.FindStringSubmatch(cleanedTitle); len(singleMatches) > 1 {
		num, _ := strconv.ParseFloat(singleMatches[1], 64)
		return ParsedRelease{Type: TypeSingle, StartNum: num, EndNum: num}, true
	}

	return ParsedRelease{}, false
}
