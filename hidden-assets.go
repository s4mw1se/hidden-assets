package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"bufio"
	"golang.org/x/net/html"
)

func getAbsoluteURL(baseURL *url.URL, attr html.Attribute) (*url.URL, error) {

	linkURL, err := url.Parse(attr.Val)
	if err != nil {
		log.Println("Failed to parse URL:", err)
		return nil, err
	}

	if !linkURL.IsAbs() {
		resolvedURL := baseURL.ResolveReference(linkURL)
		return resolvedURL, nil
	}

	// Return the resolved URL as a string.
	return linkURL, nil
}

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

				linkURL, _ := getAbsoluteURL(baseURL, attr)

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

func writeURLsToFile(urls []string, filename string) error {
    // Create a map to store the URLs
    urlMap := make(map[string]bool)

    // Check if the file exists
    if _, err := os.Stat(filename); err == nil {
        // Open the file
        file, err := os.Open(filename)
        if err != nil {
            return err
        }
        defer file.Close()

        // Read the file line by line
        scanner := bufio.NewScanner(file)
        for scanner.Scan() {
            // Add each line to the map
            urlMap[scanner.Text()] = true
        }

        // Check for errors during scanning
        if err := scanner.Err(); err != nil {
            return err
        }
    }

    // Open the file in append mode
    file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer file.Close()

    // Write the URLs to the file
    for _, u := range urls {
        // Check if the URL is already in the file
        if !urlMap[u] {
            // Write the URL to the file
            if _, err := file.WriteString(u + "\n"); err != nil {
                return err
            }
        }
    }

    return nil
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
	
	writeURLsToFile(pdfs, "urls.txt")

	fmt.Println("Done")
}
