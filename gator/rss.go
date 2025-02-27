package main

import (
    "encoding/xml"
    
)

type RSSFeed struct {
    XMLName     xml.Name    `xml:"rss"`
    Channel     RSSChannel  `xml:"channel"`
}

type RSSChannel struct {
    Title       string      `xml:"title"`
    Description string      `xml:"description"`
    Link        string      `xml:"link"`
    Items       []RSSItem   `xml:"item"`
}

type RSSItem struct {
    Title       string      `xml:"title"`
    Description string      `xml:"description"`
    Link        string      `xml:"link"`
    PubDate     string      `xml:"pubDate"`
}