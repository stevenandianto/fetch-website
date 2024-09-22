package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func downloadAsset(baseUrl, assetUrl, saveDir string) (string, error) {
	if strings.HasPrefix(assetUrl, "//") {
		assetUrl = "https:" + assetUrl
	} else if strings.HasPrefix(assetUrl, "/") {
		assetUrl = baseUrl + assetUrl
	}

	// Download asset
	resp, err := http.Get(assetUrl)
	if err != nil {
		return "", fmt.Errorf("failed to download asset %s: %v", assetUrl, err)
	}
	defer resp.Body.Close()

	// Extract the filename from the URL
	parts := strings.Split(assetUrl, "/")
	filename := parts[len(parts)-1]
	savePath := filepath.Join(saveDir, filename)

	// Create the file to save the asset
	outFile, err := os.Create(savePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file for asset %s: %v", filename, err)
	}
	defer outFile.Close()

	// Write asset content to file
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save asset %s: %v", filename, err)
	}

	fmt.Printf("Downloaded asset: %s\n", savePath)
	return savePath, nil
}

func fetchAndDownloadAssets(url string) error {
	// Fetch the web page
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch %s: %v", url, err)
	}
	defer resp.Body.Close()

	// Parse the HTML using goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %v", err)
	}

	// Create a directory to store assets
	saveDir := "assets"
	if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create assets directory: %v", err)
	}

	// Download images
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		if src != "" {
			localPath, err := downloadAsset(url, src, saveDir)
			if err == nil {
				// Update the HTML to point to the locally saved image
				s.SetAttr("src", localPath)
			}
		}
	})

	// Download CSS
	doc.Find("link[rel='stylesheet']").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if href != "" {
			localPath, err := downloadAsset(url, href, saveDir)
			if err == nil {
				// Update the HTML to point to the locally saved CSS
				s.SetAttr("href", localPath)
			}
		}
	})

	// Download JS
	doc.Find("script[src]").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		if src != "" {
			localPath, err := downloadAsset(url, src, saveDir)
			if err == nil {
				// Update the HTML to point to the locally saved JS
				s.SetAttr("src", localPath)
			}
		}
	})

	// Save the modified HTML
	html, err := doc.Html()
	if err != nil {
		return fmt.Errorf("failed to generate HTML: %v", err)
	}

	// Save modified HTML to a file
	filename := strings.Replace(url, "https://", "", 1) + ".html"
	if err := os.WriteFile(filename, []byte(html), 0644); err != nil {
		return fmt.Errorf("failed to save modified HTML: %v", err)
	}

	fmt.Printf("Successfully saved mirrored HTML %s\n", filename)
	return nil
}

func fetchAndPrintMetadata(url string) error {
	// Fetch the web page
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch %s: %v", url, err)
	}
	defer resp.Body.Close()

	// Parse the HTML using goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %v", err)
	}

	// Count links and images
	numLinks := doc.Find("a").Length()
	numImages := doc.Find("img").Length()

	// Print the metadata
	fmt.Printf("site: %s\nnum_links: %d\nnum_images: %d\nlast_fetch: %s\n",
		url, numLinks, numImages, time.Now().UTC().Format(time.RFC1123))

	return nil
}

func main() {
	urls := os.Args[1:]

	for _, url := range urls {
		err := fetchAndDownloadAssets(url)
		if err != nil {
			fmt.Println(err)
		}

		err = fetchAndPrintMetadata(url)
		if err != nil {
			fmt.Println(err)
		}
	}
}
