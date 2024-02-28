package main

import (
	"net/url"
	"strings"
	"testing"
	"os"
	"golang.org/x/net/html"
	"flag"
)

func TestGetLinks(t *testing.T) {
	// Create a simple HTML document
	doc, _ := html.Parse(strings.NewReader(`
		<html>
			<body>
				<a href="https://example.com/page1">Page 1</a>
				<a href="https://example.com/page2">Page 2</a>
			</body>
		</html>
	`))

	// Create a base URL
	baseURL, _ := url.Parse("https://example.com")

	// Call the function with the HTML document and base URL
	links := getLinks(doc, baseURL)

	// Check the length of the links
	if len(links) != 2 {
		t.Errorf("Expected 2 links, got %d", len(links))
	}

	// Check the content of the links
	if links[0] != "https://example.com/page1" || links[1] != "https://example.com/page2" {
		t.Errorf("Links do not match expected values")
	}
}

func TestGetPdfs(t *testing.T) {
    // Create a list of links
    links := []string{
        "https://example.com/doc1.pdf",
        "https://example.com/doc2.pdf",
        "https://example.com/page1",
    }

    // Call the function with the list of links
    pdfs := getPdfs(&links)

    // Check the length of the pdfs
    if len(pdfs) != 2 {
        t.Errorf("Expected 2 pdfs, got %d", len(pdfs))
    }

    // Check the content of the pdfs
    if pdfs[0] != "https://example.com/doc1.pdf" || pdfs[1] != "https://example.com/doc2.pdf" {
        t.Errorf("PDF links do not match expected values")
    }
}

func TestGetUrlArgs(t *testing.T) {
    // Set command line arguments for the test
    os.Args = []string{"cmd", "-u", "https://example.com"}

    // Create a new flag set
    flagset := flag.NewFlagSet("test", flag.ContinueOnError)

    // Call the function
    url, err := getUrlArgs(flagset)

    // Check the URL
    if url != "https://example.com" {
        t.Errorf("Expected 'https://example.com', got '%s'", url)
    }

    // Check the error
    if err != nil {
        t.Errorf("Expected nil, got error: %v", err)
    }

    // Test with no URL provided
    os.Args = []string{"cmd"}

    // Create a new flag set
    flagset = flag.NewFlagSet("test", flag.ContinueOnError)

    // Call the function
    url, err = getUrlArgs(flagset)

    // Check the URL
    if url != "" {
        t.Errorf("Expected '', got '%s'", url)
    }

    // Check the error
    if err == nil {
        t.Errorf("Expected error, got nil")
    }
}