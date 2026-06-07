package scraper

import (
	"fmt"

	"github.com/Wizzerin/immogucker-go/internal/models"
)

type Provider string

const (
	ProviderWGGesucht     Provider = "wg-gesucht"
	ProviderKleinanzeigen Provider = "kleinanzeigen"
)

type Scraper interface {
	Parse(task models.WorkerTask, taskID string) ([]models.Apartment, error)
}

func NewScraper(provider Provider) (Scraper, error) {
	switch provider {
	case ProviderWGGesucht:
		return &WGGesuchtScraper{}, nil
	case ProviderKleinanzeigen:
		return &KleinanzeigenScraper{}, nil
	default:
		return nil, fmt.Errorf("unsupported scraper provider: %s", provider)
	}
}
