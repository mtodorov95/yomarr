package indexer

import (
	"testing"

	"github.com/mtodorov95/yomarr/internal/models"
)

func TestParseTorrentTitle(t *testing.T) {
	tests := []struct {
		name       string
		title      string
		wantResult ParsedRelease
		wantValid  bool
	}{
		// --- BLACKLIST CASES ---
		{
			name:       "Blacklist exact tag",
			title:      "[Group] Awesome Manga [LN] Vol.01",
			wantResult: ParsedRelease{},
			wantValid:  false,
		},
		{
			name:       "Blacklist lower case raw text",
			title:      "Cool Light Novel series v02",
			wantResult: ParsedRelease{},
			wantValid:  false,
		},
		{
			name:       "Blacklist explicit isolated indicator",
			title:      "Manga Series ~ novel ~ Ch.55",
			wantResult: ParsedRelease{},
			wantValid:  false,
		},

		// --- SINGLE CHAPTER CASES ---
		{
			name:       "Single chapter with standard layout",
			title:      "[Group] Berserk - Chapter 375 [1080p]",
			wantResult: ParsedRelease{Type: models.TypeSingle, StartNum: 375, EndNum: 375},
			wantValid:  true,
		},
		{
			name:       "Single chapter short format",
			title:      "One Piece Ch.1110 (Digital)",
			wantResult: ParsedRelease{Type: models.TypeSingle, StartNum: 1110, EndNum: 1110},
			wantValid:  true,
		},
		{
			name:       "Single chapter lowercase c format",
			title:      "Kingdom c340",
			wantResult: ParsedRelease{Type: models.TypeSingle, StartNum: 340, EndNum: 340},
			wantValid:  true,
		},
		{
			name:       "Single chapter raw hyphen separator",
			title:      "My Hero Academia - 420",
			wantResult: ParsedRelease{Type: models.TypeSingle, StartNum: 420, EndNum: 420},
			wantValid:  true,
		},
		{
			name:       "Single chapter with dot decimal",
			title:      "Chainsaw Man Ch. 150.5",
			wantResult: ParsedRelease{Type: models.TypeSingle, StartNum: 150.5, EndNum: 150.5},
			wantValid:  true,
		},
		{
			name:       "Single chapter with legacy x decimal notation",
			title:      "Manga Title c040x1",
			wantResult: ParsedRelease{Type: models.TypeSingle, StartNum: 40.1, EndNum: 40.1},
			wantValid:  true,
		},

		// --- CHAPTER RANGE CASES ---
		{
			name:       "Chapter range with explicit prefixes",
			title:      "[Group] Claymore Ch 01-12 [Digital]",
			wantResult: ParsedRelease{Type: models.TypeRange, StartNum: 1, EndNum: 12},
			wantValid:  true,
		},
		{
			name:       "Chapter range short format",
			title:      "Bleach c01-74",
			wantResult: ParsedRelease{Type: models.TypeRange, StartNum: 1, EndNum: 74},
			wantValid:  true,
		},
		{
			name:       "Chapter range with decimal numbers",
			title:      "Bonus Content Ch 14.5 - 15.5",
			wantResult: ParsedRelease{Type: models.TypeRange, StartNum: 14.5, EndNum: 15.5},
			wantValid:  true,
		},

		// --- VOLUME CASES ---
		{
			name:       "Standard volume single",
			title:      "[Group] Monster Vol.02 [Digital]",
			wantResult: ParsedRelease{Type: models.TypeVolume, StartNum: 2, EndNum: 2},
			wantValid:  true,
		},
		{
			name:       "Volume explicit long-form",
			title:      "Hellsing Volume 4 (Batch)",
			wantResult: ParsedRelease{Type: models.TypeVolume, StartNum: 4, EndNum: 4},
			wantValid:  true,
		},
		{
			name:       "Volume short lowercase v format",
			title:      "Naruto v03",
			wantResult: ParsedRelease{Type: models.TypeVolume, StartNum: 3, EndNum: 3},
			wantValid:  true,
		},
		{
			name:       "Volume range batch",
			title:      "20th Century Boys v01-05",
			wantResult: ParsedRelease{Type: models.TypeVolume, StartNum: 1, EndNum: 5},
			wantValid:  true,
		},
		{
			name:       "Multi-Volume batch with repeated prefix keyword",
			title:      "Spy x Family Vol.16 - Vol.17",
			wantResult: ParsedRelease{Type: models.TypeVolume, StartNum: 16, EndNum: 17},
			wantValid:  true,
		},

		// --- JAPANESE VOLUME SYNTAX ---
		{
			name:       "Japanese volume structure exact",
			title:      "Manga Title 第02巻",
			wantResult: ParsedRelease{Type: models.TypeVolume, StartNum: 2, EndNum: 2},
			wantValid:  true,
		},
		{
			name:       "Japanese volume with internal whitespace layout",
			title:      "Manga Title 第 12 巻",
			wantResult: ParsedRelease{Type: models.TypeVolume, StartNum: 12, EndNum: 12},
			wantValid:  true,
		},
		{
			name:       "Japanese volume batch range",
			title:      "Manga Title 第01-04巻",
			wantResult: ParsedRelease{Type: models.TypeVolume, StartNum: 1, EndNum: 4},
			wantValid:  true,
		},

		// --- YEAR CLEANING EDGE CASES ---
		{
			name:       "Cleans year ranges safely without breaking volume matching",
			title:      "Pluto (2003-2008) Vol.1 [Digital]",
			wantResult: ParsedRelease{Type: models.TypeVolume, StartNum: 1, EndNum: 1},
			wantValid:  true,
		},

		// --- UNMATCHABLE FALLBACK CASES ---
		{
			name:       "No recognizable tags or numbers",
			title:      "[Group] Random Promo Art Assets Bundle",
			wantResult: ParsedRelease{},
			wantValid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotValid := ParseTorrentTitle(tt.title)

			if gotValid != tt.wantValid {
				t.Errorf("ParseTorrentTitle() gotValid = %v, want %v for title: %q", gotValid, tt.wantValid, tt.title)
			}

			if gotResult != tt.wantResult {
				t.Errorf("ParseTorrentTitle() gotResult = %+v, want %+v for title: %q", gotResult, tt.wantResult, tt.title)
			}
		})
	}
}
