package main

import (
	"bufio"
	"flag"
	"net/url"
	"os"
	"strings"
	"testing"

	"golang.org/x/net/html"
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

func TestGetAbsoluteURL(t *testing.T) {
	baseURL, _ := url.Parse("https://example.com")
	attr := html.Attribute{
		Key: "href",
		Val: "/page1",
	}

	// Test with a relative URL
	resultURL, err := getAbsoluteURL(baseURL, attr)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if resultURL.String() != "https://example.com/page1" {
		t.Errorf("Expected 'https://example.com/page1', got '%s'", resultURL.String())
	}

	// Test with an absolute URL
	attr.Val = "https://other.com/page2"
	resultURL, err = getAbsoluteURL(baseURL, attr)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if resultURL.String() != "https://other.com/page2" {
		t.Errorf("Expected 'https://other.com/page2', got '%s'", resultURL.String())
	}

	// Test with an invalid URL
	attr.Val = "://invalid"
	_, err = getAbsoluteURL(baseURL, attr)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestWriteURLsToFile(t *testing.T) {
	// Test data
	urls := []string{"http://example.com", "http://test.com", "http://example.com"}

	// Test file name
	filename := "testfile.txt"

	// Call the function with the test data
	err := writeURLsToFile(urls, filename)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer file.Close()

	// Create a map to store the URLs
	urlMap := make(map[string]bool)

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Add each line to the map
		urlMap[scanner.Text()] = true
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check if the URLs were written correctly
	for _, u := range urls {
		if !urlMap[u] {
			t.Fatalf("Expected %v in file, not found", u)
		}
	}

	// Clean up
	os.Remove(filename)
}
