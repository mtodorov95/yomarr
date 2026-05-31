package metadata

import "github.com/mtodorov95/yomarr/internal/models"


type AggregatorMetadataProvider struct {
	MangaDex *MangaDexProvider
	Anilist  *AnilistProvider
}

func NewAggregatorMetadataProvider(md *MangaDexProvider, al *AnilistProvider) *AggregatorMetadataProvider {
	return &AggregatorMetadataProvider{
		MangaDex: md,
		Anilist:  al,
	}
}

func (a *AggregatorMetadataProvider) Search(query string) ([]models.Series, error) {
	return a.MangaDex.Search(query)
}

func (a *AggregatorMetadataProvider) GetDetails(id string) (*models.Series, error) {
	series, err := a.MangaDex.GetDetails(id)
	if err != nil {
		return nil, err
	}

	if series.AnilistID != nil && *series.AnilistID != "" {
		alDetails, err := a.Anilist.GetDetails(*series.AnilistID)
		if err == nil && alDetails != nil {
			series.TotalChapters = alDetails.TotalChapters 
		}
	}

	return series, nil
}
