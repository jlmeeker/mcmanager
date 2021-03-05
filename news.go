package main

import (
	"fmt"

	"github.com/badoux/goscraper"
	"github.com/mmcdole/gofeed"
)

// NewsItem is the simplest form of data for a card on the "home" view
type NewsItem struct {
	Posted  string                     `json:"posted"`
	Preview *goscraper.DocumentPreview `json:"preview"`
}

// VanillaNews cached copy of the latest pulled vanilla news
var VanillaNews = []NewsItem{}

func sitePreview(site string) *goscraper.Document {
	p, e := goscraper.Scrape(site, 5)
	if e != nil {
		fmt.Printf("preview error: %s\n", e.Error())
	}
	return p
}

func vanillaNews(max int) ([]NewsItem, error) {
	var items []NewsItem
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL("https://minecraft.net/en-us/feeds/community-content/rss")
	if err != nil {
		return items, err
	}

	var count = 0
	for _, item := range feed.Items {
		//Mon Jan 2 15:04:05 -0700 MST 2006
		ni := NewsItem{
			Posted:  item.PublishedParsed.Local().Format("Jan 2, 2006"),
			Preview: &sitePreview(item.Link).Preview,
		}
		items = append(items, ni)

		count++
		if count >= max {
			break
		}
	}
	return items, nil
}
