package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	reader := strings.NewReader("")
	var client http.Client
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, reader)
	if err != nil {
		fmt.Println("Issue forming http req: $s", err)
	}

	req.Header.Set("User-Agent", "gator")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Issue sending http request %s\n", err)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Issue reading data %s", err)
	}

	var feed RSSFeed
	err = xml.Unmarshal(data, &feed)
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for _, value := range feed.Channel.Item {
		value.Title = html.UnescapeString(value.Title)
		value.Description = html.UnescapeString(value.Description)
	}

	return &feed, err
}
