package main

// TODO:
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
