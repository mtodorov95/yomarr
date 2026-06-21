package indexer

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/mtodorov95/yomarr/internal/models"
)

var (
	// Matches strict light novel indicators to block accidental downloads
	// Matches "[LN]", "(LN)", "~ LN ~", "Light Novel", "Novel v01"
	lnBlacklistRegex = regexp.MustCompile(`(?i)\b(ln|light\s*novel|novel)\b`)

	// Matches and targets year ranges to erase them: "(2023-2025)", "[2023-2025]"
	yearRangeRegex = regexp.MustCompile(`(?i)[([ ]\d{4}\s*-\s*\d{4}[)\] ]`)

	// Matches specific individual chapters: "Ch. 10", "- 10", "c040x1", "c24"
	chSingleRegex    = regexp.MustCompile(`(?i)(?:-\s*|ch(?:apter)?\.?\s*|\bc\.?\s*)(\d+(?:[\.x]\d+)?)`)

	// Matches ranges: "Ch 01-12", "c01-12", "05-15"
	chRangeRegex = regexp.MustCompile(`(?i)(?:ch(?:apter)?\.?\s+|\bc\.?\s*)?(\d+(?:[\.x]\d+)?)\s*-\s*(\d+(?:[\.x]\d+)?)`)

	// Matches volumes: "Vol.02", "v03", "Volume 4", "v01-05"
	VolRegex = regexp.MustCompile(`(?i)(?:vol(?:ume)?\.?\s*|v)(\d+)(?:\s*-\s*(?:vol(?:ume)?\.?\s*|v)?(\d+))?`)

	// Matches Japanese volume syntax variants like "第02巻", "第 2 巻", "第02-04巻"
	VolJaRegex = regexp.MustCompile(`(?i)第\s*(\d+)\s*(?:-\s*(\d+)\s*)?巻`)
)

type ParsedRelease struct {
	Type     models.ReleaseType
	StartNum float64
	EndNum   float64
}

func ParseTorrentTitle(title string) (ParsedRelease, bool) {
	if lnBlacklistRegex.MatchString(title) {
		return ParsedRelease{}, false
	}

	cleanedTitle := yearRangeRegex.ReplaceAllString(title, " ")

	if jaVolMatches := VolJaRegex.FindStringSubmatch(cleanedTitle); len(jaVolMatches) > 1 {
		start, _ := strconv.ParseFloat(jaVolMatches[1], 64)
		if jaVolMatches[2] != "" {
			end, _ := strconv.ParseFloat(jaVolMatches[2], 64)
			return ParsedRelease{Type: models.TypeVolume, StartNum: start, EndNum: end}, true
		}
		return ParsedRelease{Type: models.TypeVolume, StartNum: start, EndNum: start}, true
	}

	if volMatches := VolRegex.FindStringSubmatch(cleanedTitle); len(volMatches) > 1 {
		start, _ := strconv.ParseFloat(volMatches[1], 64)
		if volMatches[2] != "" {
			end, _ := strconv.ParseFloat(volMatches[2], 64)
			return ParsedRelease{Type: models.TypeVolume, StartNum: start, EndNum: end}, true
		}
		return ParsedRelease{Type: models.TypeVolume, StartNum: start, EndNum: start}, true
	}

	if rangeMatches := chRangeRegex.FindStringSubmatch(cleanedTitle); len(rangeMatches) > 2 {
		startStr := strings.ReplaceAll(strings.ToLower(rangeMatches[1]), "x", ".")
		endStr := strings.ReplaceAll(strings.ToLower(rangeMatches[2]), "x", ".")

		start, _ := strconv.ParseFloat(startStr, 64)
		end, _ := strconv.ParseFloat(endStr, 64)
		if end > start {
			return ParsedRelease{Type: models.TypeRange, StartNum: start, EndNum: end}, true
		}
	}

	if singleMatches := chSingleRegex.FindStringSubmatch(cleanedTitle); len(singleMatches) > 1 {
		numStr := strings.ReplaceAll(strings.ToLower(singleMatches[1]), "x", ".")

		num, _ := strconv.ParseFloat(numStr, 64)
		return ParsedRelease{Type: models.TypeSingle, StartNum: num, EndNum: num}, true
	}

	return ParsedRelease{}, false
}
