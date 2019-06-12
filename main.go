package main

// TODO:
// - Extract parameters to config file (TOML)
//   - Including feed list, feed item limit
// - Make the html nicer

import (
	// "fmt"
	"github.com/mmcdole/gofeed"
	// "html"
	"html/template"
	"log"
	"net/http"
	"sort"
	"time"
)

var feeds = []string{
	"https://www.schneier.com/blog/atom.xml",
	"https://blog.talosintelligence.com/feeds/comments/default",
	"https://news.ycombinator.com/rss",
	"https://krebsonsecurity.com/feed/",
	"https://news.softpedia.com/newsRSS/Security-5.xml",
	"http://feeds.arstechnica.com/arstechnica/security",
	"https://www.wired.com/feed/category/security/latest/rss",
	"https://www.vice.com/en_us/rss/section/tech",
	"https://www.reddit.com/r/golang.rss",
	"https://www.reddit.com/r/devops.rss",
	"https://www.reddit.com/r/netsec.rss",
	"https://www.reddit.com/r/artc.rss",
	"https://www.reddit.com/r/crypto.rss",
	//"https://googleprojectzero.blogspot.com/feeds/posts/default",
}

const (
	feedItemLimit = 5
)

type ItemSlice []*gofeed.Item

type IndexTemplateParams struct {
	Title    string
	ItemList ItemSlice
}

func (is ItemSlice) Len() int {
	return len(is)
}

func (is ItemSlice) Less(i, j int) bool {
	iTime, _ := time.Parse(time.RFC3339, is[i].Published)
	jTime, _ := time.Parse(time.RFC3339, is[j].Published)

	return iTime.After(jTime)
}

func (is ItemSlice) Swap(i, j int) {
	is[i], is[j] = is[j], is[i]
}

func main() {
	var allItems []*gofeed.Item
	var feed *gofeed.Feed
	fp := gofeed.NewParser()
	for _, feedUrl := range feeds {
		feed, _ = fp.ParseURL(feedUrl)
		allItems = append(
			allItems, feed.Items[:feedItemLimit]...)
	}

	var is ItemSlice = allItems
	sort.Sort(is)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// TODO: do any of the fields need to be html.EscapeString-ed?
		t, _ := template.ParseFiles("index.html")
		t.Execute(w, &IndexTemplateParams{
			Title:    "Thar be Blogs",
			ItemList: is,
		})
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
