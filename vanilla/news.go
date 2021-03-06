package vanilla

import (
	"fmt"

	"github.com/badoux/goscraper"
	"github.com/jlmeeker/mcmanager/newsitem"
	"github.com/mmcdole/gofeed"
)

// News cached copy of the latest pulled vanilla news
var News = []newsitem.NewsItem{}

func sitePreview(site string) *goscraper.Document {
	p, e := goscraper.Scrape(site, 5)
	if e != nil {
		fmt.Printf("preview error: %s\n", e.Error())
	}
	return p
}

// RefreshNews pulls and caches the latest news
func RefreshNews(max int) error {
	var items []newsitem.NewsItem
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL("https://minecraft.net/en-us/feeds/community-content/rss")
	if err != nil {
		return err
	}

	var count = 0
	for _, item := range feed.Items {
		//Mon Jan 2 15:04:05 -0700 MST 2006
		ni := newsitem.NewsItem{
			Posted:  item.PublishedParsed.Local().Format("Jan 2, 2006"),
			Preview: &sitePreview(item.Link).Preview,
		}
		items = append(items, ni)

		count++
		if count >= max {
			break
		}
	}

	News = items
	return nil
}
