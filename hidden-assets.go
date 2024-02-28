package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/net/html"
)

// getLinks is a function that extracts all the href values from anchor tags in an HTML document.
// It takes a pointer to an html.Node as an argument and returns a slice of strings.
func getLinks(doc *html.Node, baseURL *url.URL) (links []string) {
	// Check if the node is an ElementNode and if it's an anchor tag.
	if doc.Type == html.ElementNode && doc.Data == "a" {
		// Iterate over the attributes of the node.
		for _, attr := range doc.Attr {
			// If the attribute is an href, append its value to the links slice.
			if attr.Key == "href" {
				// Parse the href value as a URL.
				linkURL, err := url.Parse(attr.Val)
				if err != nil {
					log.Println("Failed to parse URL:", err)
					continue
				}

				// If the link is a relative path, resolve it against the base URL.
				if !linkURL.IsAbs() {
					linkURL = baseURL.ResolveReference(linkURL)
				}

				// Append the resolved URL to the links slice.
				links = append(links, linkURL.String())
			}
		}
	}

	// Iterate over the children of the current node.
	for c := doc.FirstChild; c != nil; c = c.NextSibling {
		// Recursively call getLinks on each child and append the results to the links slice.
		links = append(links, getLinks(c, baseURL)...)
	}

	// Return the links slice.
	return links
}

func getPdfs(links *[]string) (pdfs []string) {
	for _, link := range *links {
		if len(link) > 4 && link[len(link)-4:] == ".pdf" {
			pdfs = append(pdfs, link)
		}
	}
	return pdfs
}

func getUrlArgs(flagset *flag.FlagSet) (string, error) {
	u := flagset.String("u", "", "URL to scan")
	flagset.Parse(os.Args[1:])

	if *u == "" {
		err := errors.New("URL is required")
		return "", err
	}

	return *u, nil
}

func main() {
  flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	u, err := getUrlArgs(flagSet)
	if err != nil {
		log.Fatal("URL is required")
		os.Exit(1)
	}

	url, err := url.Parse(u)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Get the URL
	resp, err := http.Get(u)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatal("Error: ", resp.Status)
		os.Exit(1)
	}

	// Parse the HTML
	doc, err := html.Parse(resp.Body)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	links := getLinks(doc, url)

	pdfs := getPdfs(&links)
	for _, pdf := range pdfs {
		fmt.Println(pdf)
	}

	fmt.Println("Done")
}
