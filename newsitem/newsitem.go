package newsitem

import "github.com/badoux/goscraper"

// NewsItem is the simplest form of data for a card on the "home" view
type NewsItem struct {
	Posted  string                     `json:"posted"`
	Preview *goscraper.DocumentPreview `json:"preview"`
}
