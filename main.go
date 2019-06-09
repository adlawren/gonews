package main

// TODO:
// - Fetch the N most recent posts from each feed and merge them into a single feed
// - Sort the feed by publish date, the most recent article first
// - Render the sorted feed

import (
	"fmt"
	"github.com/mmcdole/gofeed"
	"html"
	// "log"
	"net/http"
	"sort"
	"time"
)

var feeds = []string{
	"https://www.schneier.com/blog/atom.xml",
	"https://blog.talosintelligence.com/feeds/comments/default",
	//"https://googleprojectzero.blogspot.com/feeds/posts/default",
}

type ItemSlice []*gofeed.Item

// TODO: check nil case for each of these methods
// func (is *ItemSlice) Len() int {
// 	return len(*is)
// }

// func (is *ItemSlice) Less(i, j int) bool {
// 	iTime, _ := time.Parse(time.RFC3339, (*is)[i].Published)
// 	jTime, _ := time.Parse(time.RFC3339, (*is)[j].Published)

// 	return iTime.Before(jTime)
// }

// func (is *ItemSlice) Swap(i, j int) {
// 	(*is)[i], (*is)[j] = (*is)[j], (*is)[i]
// }
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
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello! (%q)", html.EscapeString(r.URL.Path))
	})

	//log.Fatal(http.ListenAndServe(":8080", nil))

	// fp := gofeed.NewParser()
	// feed, _ := fp.ParseURL("https://www.schneier.com/blog/atom.xml")
	// fmt.Println(time.Parse(time.RFC3339, feed.Items[0].Published))
	// time1, _ := time.Parse(time.RFC3339, feed.Items[0].Published)

	// feed2, _ := fp.ParseURL("https://blog.talosintelligence.com/feeds/comments/default")
	// fmt.Println(time.Parse(time.RFC3339, feed2.Items[0].Published))
	// time2, _ := time.Parse(time.RFC3339, feed2.Items[0].Published)

	// if time1.After(time2) {
	// 	fmt.Println("Yass")
	// }

	var allItems []*gofeed.Item
	var feed *gofeed.Feed
	fp := gofeed.NewParser()
	for _, feedUrl := range feeds {
		feed, _ = fp.ParseURL(feedUrl)
		allItems = append(allItems, feed.Items...)
	}

	// TODO: try to get pointer receiver working with this
	//var is ItemSlice = append(feed.Items, feed2.Items...)
	var is ItemSlice = allItems

	//ref := &is
	//fmt.Printf("%v\n", len(is))
	//fmt.Printf("%v\n", len(*ref))
	//fmt.Printf("%v\n", (*ref)[0])
	sort.Sort(is)

	for i := 0; i < len(is); i++ {
		fmt.Println(time.Parse(time.RFC3339, is[i].Published))
	}

	// fmt.Println(time.Parse(time.RFC3339, is[1].Published))
	// fmt.Println(time.Parse(time.RFC3339, is[2].Published))
}
