package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

type Crawler struct {
	isVisited map[string]bool
	pdfLinks  []string
	visitedUrlCount int
}

var allowedHosts = map[string]bool{
	"gadoe.org":     true,
	"www.gadoe.org": true,
}

func NewCrawler() *Crawler {
	return &Crawler{
		isVisited: make(map[string]bool),
		pdfLinks:  make([]string, 0),
		visitedUrlCount: 0,
	}
}

func isAllowed(url *url.URL) bool {
	for hostPattern := range allowedHosts {
		match, _ := pathMatchesHostPattern(url.Hostname(), hostPattern)
		if match {
			return true
		}
	}
	return false
}

func pathMatchesHostPattern(path, pattern string) (bool, error) {
	pathSegments := strings.Split(path, ".")
	patternSegments := strings.Split(pattern, ".")
	if len(pathSegments) != len(patternSegments) {
		return false, nil
	}
	for i := range patternSegments {
		if patternSegments[i] == "*" {
			continue
		}
		if patternSegments[i] != pathSegments[i] {
			return false, nil
		}
	}
	return true, nil
}

func (c *Crawler) getAbsoluteURL(baseURL *url.URL, attr html.Attribute) (*url.URL, error) {
	linkURL, err := url.Parse(attr.Val)
	if err != nil {
		return nil, err
	}
	if !linkURL.IsAbs() {
		return baseURL.ResolveReference(linkURL), nil
	}

	return linkURL, nil
}

func (c *Crawler) getLinks(doc *html.Node, baseURL *url.URL) {
	var links []string
	if doc.Type == html.ElementNode && doc.Data == "a" {
		for _, attr := range doc.Attr {
			if attr.Key == "href" {
				linkURL, err := c.getAbsoluteURL(baseURL, attr)

				if err != nil {
					continue
				}

				if linkURL.Scheme != "http" && linkURL.Scheme != "https" {
					continue
				}

				if !isAllowed(linkURL) {
					continue
				}

				if c.isVisited[linkURL.String()] {
					continue
				}

				log.Print("found: ", linkURL)

				links = append(links, linkURL.String())
			}
		}
	}

	var wg sync.WaitGroup
	for _, link := range links {
		log.Println("Checking link: ", link)
		if strings.Contains(link, ".pdf") {
			c.pdfLinks = append(c.pdfLinks, link)
			c.isVisited[link] = true
			continue
		}

		c.isVisited[link] = false

		wg.Add(1)
		go func(link string) {
			defer wg.Done()

			u, err := url.Parse(link)
			if err != nil {
				log.Println("Invalid link:", link)
				return
			}
			c.crawl(u) 
		}(link)
	}
	wg.Wait()

	// write pdf links to file
	c.writeURLsToFile("pdf_urls.txt", "pdf")
	c.writeURLsToFile("visited_urls.json", "url")

	for child := doc.FirstChild; child != nil; child = child.NextSibling {
		c.getLinks(child, baseURL)
	}
}

func (c *Crawler) parseHtml(url *url.URL) {
	c.isVisited[url.String()] = true
	resp, err := http.Get(url.String())

	if err != nil {
		log.Println("Failed to get the URL:", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Error:", resp.Status)

	}

	log.Println("Status:", resp.Status)

	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Println("Failed to parse the HTML:", err)
		return
	}

	log.Println("Parsed HTML for: ", url.String())

	c.getLinks(doc, url)
}

func (c *Crawler) crawl(url *url.URL) {
	if url == nil {
		return
	}

	//check if url is in c.visitedUrl
	if c.isVisited[url.String()] {
		return
	}

	c.isVisited[url.String()] = true
	c.visitedUrlCount++
	
	if !isAllowed(url) {
		log.Println("Host not allowed:", url.Hostname())
		return
	}
	
	log.Println("Crawling:", url.String())

	if !isAllowed(url) {
		log.Println("Host not allowed:", url.Hostname())
		return
	}

	c.parseHtml(url)
}

func (c *Crawler) flushURLs() {
	c.writeURLsToFile("pdf_urls.txt", "pdf")
	c.writeURLsToFile("visited_urls.json", "url")
	c.visitedUrlCount = 0
}


func (c *Crawler) writeURLsToFile(filename string, data string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	if data == "pdf" {
		for _, pdf := range c.pdfLinks {
			if _, err := file.WriteString(pdf + "\n"); err != nil {
				return err
			}
		}
	}

	if data == "url" {
		jason_data, err := json.Marshal(c.isVisited)
		if err != nil {
			return err
		}

		file.Write(jason_data)

	}

	return nil
}

func start_crawler(starting_url string) {
	crawler := NewCrawler()
	u, err := url.Parse(starting_url)

	if err != nil {
		log.Fatal(err)
	}

	crawler.isVisited[u.String()] = false
	crawler.crawl(u)

}

func main() {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	u := flagSet.String("u", "", "URL to scan")
	flagSet.Parse(os.Args[1:])

	if *u == "" {
		log.Fatal("URL is required")
	}

	start_crawler(*u)
}