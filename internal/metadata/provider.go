package metadata

import "github.com/mtodorov95/yomarr/internal/models"

type Provider interface {
	Search(query string) ([]models.Series, error)
	GetDetails(id string) (*models.Series, error)
}
