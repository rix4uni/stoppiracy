package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/rix4uni/stoppiracy/banner"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

// ANSI color codes
const (
	REDCOLOR    = "\033[91m"
	GREENCOLOR  = "\033[92m"
	YELLOWCOLOR = "\033[93m"
	CYANCOLOR   = "\033[96m"
	BLUECOLOR   = "\033[94m"
	RESETCOLOR  = "\033[0m"
)

// Result structure for final JSON output
type DomainResult struct {
	Name         string   `json:"name"`
	Logo         string   `json:"logo"`
	InternalLink []string `json:"internal_link"`
	Email        []string `json:"email"`
	Matched      []string `json:"matched"`
	LastUpdated  string   `json:"last_updated"`
}

// Combined result for a domain
type DomainData struct {
	URL           string
	Keywords      []string
	Favicons      []string
	Emails        []string
	InternalLinks []string
}

// Email extraction patterns
var excludePatterns = []string{
	".jpg", ".png", ".gif", ".webp", ".ico", ".mp4", ".pdf", ".eot",
	".doc", ".docx", ".xls", ".xlsx", ".woff", ".woff2", ".css", ".json",
	".xml", ".rss", ".svg", ".yaml", ".yml", ".csv", ".dockerfile", ".cfg",
	".lock", ".js", ".md", ".toml",
}

var emailRegex = regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)

// checkKEYBOARD checks if content contains any of the keywords
func checkKEYBOARD(content string, keywords []string) []string {
	var matchedKeywords []string
	for _, keyword := range keywords {
		if strings.Contains(content, keyword) {
			matchedKeywords = append(matchedKeywords, keyword)
		}
	}
	return matchedKeywords
}

// ensureProtocol ensures URL has http:// or https:// prefix
func ensureProtocol(input string, client *http.Client) string {
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		// Try HTTPS first
		testURL := "https://" + input
		resp, err := client.Head(testURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			return testURL
		}
		// Fallback to HTTP
		return "http://" + input
	}
	return input
}

// getFaviconUrlsFromHTML extracts all favicons from HTML content
func getFaviconUrlsFromHTML(htmlContent string, baseURL string) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	var favicons []string

	// Find all <link rel="icon"> and <link rel="shortcut icon"> elements
	doc.Find("link[rel='icon'], link[rel=\"icon\"], link[rel='shortcut icon'], link[rel=\"shortcut icon\"]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			var absoluteURL string
			if strings.HasPrefix(href, "http") {
				absoluteURL = href
			} else {
				absoluteURL = base.ResolveReference(&url.URL{Path: href}).String()
			}

			// Remove everything after .png or .ico (strip query parameters)
			if strings.Contains(absoluteURL, ".png") {
				absoluteURL = strings.Split(absoluteURL, ".png")[0] + ".png"
			} else if strings.Contains(absoluteURL, ".ico") {
				absoluteURL = strings.Split(absoluteURL, ".ico")[0] + ".ico"
			}

			favicons = append(favicons, absoluteURL)
		}
	})

	return favicons, nil
}

// getFaviconUrls tries to fetch /favicon.ico as fallback (only used when no favicons found in HTML)
func getFaviconUrls(baseURL string, client *http.Client) ([]string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	faviconURL := base.ResolveReference(&url.URL{Path: "/favicon.ico"}).String()
	resp, err := client.Get(faviconURL)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			return []string{faviconURL}, nil
		}
	}

	return []string{}, nil
}

// selectShortestFavicon selects the shortest favicon URL, or first if same length
func selectShortestFavicon(favicons []string) string {
	if len(favicons) == 0 {
		return ""
	}
	if len(favicons) == 1 {
		return favicons[0]
	}

	// Find the shortest length
	shortestLength := len(favicons[0])
	for _, favicon := range favicons {
		if len(favicon) < shortestLength {
			shortestLength = len(favicon)
		}
	}

	// Return the first favicon with the shortest length
	for _, favicon := range favicons {
		if len(favicon) == shortestLength {
			return favicon
		}
	}

	return favicons[0] // Fallback (should never reach here)
}

// shouldExclude checks if link should be excluded
func shouldExclude(link string) bool {
	lowerLink := strings.ToLower(link)
	for _, pattern := range excludePatterns {
		if strings.Contains(lowerLink, pattern) {
			return true
		}
	}
	return false
}

// isValidEmail validates email address
func isValidEmail(email string) bool {
	if strings.Contains(strings.ToLower(email), "example") {
		return false
	}
	if strings.Contains(strings.ToLower(email), "email") {
		return false
	}
	if len(email) < 5 {
		return false
	}
	if strings.Contains(email, ".png") || strings.Contains(email, ".jpg") ||
		strings.Contains(email, ".webp") || strings.Contains(email, ".gif") {
		return false
	}
	return true
}

// extractEmailsAndLinksFromHTML extracts emails and links from HTML content
func extractEmailsAndLinksFromHTML(html string, targetURL, baseURL string) (map[string]bool, map[string]bool) {
	emails := make(map[string]bool)
	links := make(map[string]bool)

	// Extract href links
	hrefPatterns := []string{
		`href="([^"]*)"`,
		`href='([^']*)'`,
	}

	for _, pattern := range hrefPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(html, -1)

		for _, match := range matches {
			if len(match) < 2 {
				continue
			}
			link := match[1]

			if shouldExclude(link) {
				continue
			}

			// Extract emails from mailto links
			if strings.HasPrefix(strings.ToLower(link), "mailto:") {
				emailPart := link[7:]
				emailsFound := emailRegex.FindAllString(emailPart, -1)
				for _, email := range emailsFound {
					if email != "" && isValidEmail(email) {
						emails[email] = true
					}
				}
				continue
			}

			// Extract emails from link text
			emailsInText := emailRegex.FindAllString(link, -1)
			for _, email := range emailsInText {
				if email != "" && isValidEmail(email) {
					emails[email] = true
				}
			}

			// Resolve relative URLs
			absoluteLink, err := url.Parse(link)
			if err != nil {
				continue
			}

			base, err := url.Parse(targetURL)
			if err != nil {
				continue
			}

			resolvedLink := base.ResolveReference(absoluteLink).String()

			// Check if it's an internal link
			baseParsed, err := url.Parse(baseURL)
			if err != nil {
				continue
			}

			resolvedParsed, err := url.Parse(resolvedLink)
			if err != nil {
				continue
			}

			if resolvedParsed.Host == baseParsed.Host {
				links[resolvedLink] = true
			}
		}
	}

	// Extract emails from entire page content
	allEmailsInPage := emailRegex.FindAllString(html, -1)
	for _, email := range allEmailsInPage {
		if email != "" && isValidEmail(email) {
			emails[email] = true
		}
	}

	return emails, links
}

// extractEmailsAndLinks extracts emails, links, and keywords from a URL (used for internal links)
func extractEmailsAndLinks(targetURL, baseURL string, keywords []string, client *http.Client) ([]string, map[string]bool, map[string]bool, error) {
	resp, err := client.Get(targetURL)
	if err != nil {
		return nil, nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, nil, err
	}

	html := string(body)
	emails, links := extractEmailsAndLinksFromHTML(html, targetURL, baseURL)

	// Extract keywords from HTML content
	matchedKeywords := checkKEYBOARD(html, keywords)

	return matchedKeywords, emails, links, nil
}

// processDomain processes a single domain and collects all data
func processDomain(domainURL string, keywords []string, timeout time.Duration, maxWorkers int) (*DomainData, error) {
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy:             http.ProxyFromEnvironment,
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives: false,
		},
	}

	// Ensure protocol
	domainURL = ensureProtocol(domainURL, client)

	data := &DomainData{
		URL:           domainURL,
		Keywords:      []string{},
		Favicons:      []string{},
		Emails:        []string{},
		InternalLinks: []string{},
	}

	// Step 1: Fetch main page ONCE
	resp, err := client.Get(domainURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch domain: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	htmlContent := string(body)

	// Step 2: Extract keywords from HTML (no additional request)
	matchedKeywords := checkKEYBOARD(htmlContent, keywords)
	keywordSet := make(map[string]bool)
	mainPageMatched := len(matchedKeywords) > 0
	for _, keyword := range matchedKeywords {
		keywordSet[keyword] = true
	}

	// Step 3: Extract favicons from HTML (no additional request)
	favicons, err := getFaviconUrlsFromHTML(htmlContent, domainURL)
	if err == nil && len(favicons) > 0 {
		data.Favicons = favicons
	} else {
		// Fallback: try /favicon.ico if no favicons found in HTML
		fallbackFavicons, err := getFaviconUrls(domainURL, client)
		if err == nil && len(fallbackFavicons) > 0 {
			data.Favicons = fallbackFavicons
		}
	}

	// Step 4: Extract emails and links from HTML (no additional request)
	mainEmails, mainLinks := extractEmailsAndLinksFromHTML(htmlContent, domainURL, domainURL)
	for email := range mainEmails {
		data.Emails = append(data.Emails, email)
	}

	// Step 5: Process internal links concurrently for more emails and keywords (MUST NEEDED)
	links := make([]string, 0, len(mainLinks))
	for link := range mainLinks {
		links = append(links, link)
	}

	if len(links) > 0 {
		var wg sync.WaitGroup
		emailChan := make(chan string, len(links))
		keywordChan := make(chan string, len(links))
		internalLinkChan := make(chan string, len(links))
		semaphore := make(chan struct{}, maxWorkers)

		for _, link := range links {
			wg.Add(1)
			go func(l string) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				linkMatched := false

				// Check keywords in URL path
				linkLower := strings.ToLower(l)
				for _, keyword := range keywords {
					if strings.Contains(linkLower, strings.ToLower(keyword)) {
						keywordChan <- keyword
						linkMatched = true
					}
				}

				// Extract emails, keywords, and links from HTML content
				linkKeywords, emails, _, err := extractEmailsAndLinks(l, domainURL, keywords, client)
				if err == nil {
					// Send emails
					for email := range emails {
						emailChan <- email
					}
					// Send keywords found in HTML
					if len(linkKeywords) > 0 {
						linkMatched = true
					}
					for _, keyword := range linkKeywords {
						keywordChan <- keyword
					}
				}

				// If link matched keywords, send it to the channel
				if linkMatched {
					internalLinkChan <- l
				}
			}(link)
		}

		go func() {
			wg.Wait()
			close(emailChan)
			close(keywordChan)
			close(internalLinkChan)
		}()

		emailSet := make(map[string]bool)
		for email := range emailChan {
			emailSet[email] = true
		}

		// Collect keywords from internal links
		for keyword := range keywordChan {
			keywordSet[keyword] = true
		}

		// Collect matched internal links
		internalLinkSet := make(map[string]bool)
		for link := range internalLinkChan {
			internalLinkSet[link] = true
		}

		// Add unique emails
		for email := range emailSet {
			// Check if not already in data.Emails
			found := false
			for _, existingEmail := range data.Emails {
				if existingEmail == email {
					found = true
					break
				}
			}
			if !found {
				data.Emails = append(data.Emails, email)
			}
		}

		// Convert internal link set to slice
		for link := range internalLinkSet {
			data.InternalLinks = append(data.InternalLinks, link)
		}
	}

	// Convert keyword set to slice
	for keyword := range keywordSet {
		data.Keywords = append(data.Keywords, keyword)
	}

	// Add base domain URL to internal links if keywords matched on main page
	if mainPageMatched {
		// Check if domainURL is not already in the list to avoid duplicates
		found := false
		for _, existingLink := range data.InternalLinks {
			if existingLink == domainURL {
				found = true
				break
			}
		}
		if !found {
			data.InternalLinks = append(data.InternalLinks, domainURL)
		}
	}

	return data, nil
}

// getCurrentDate returns current date in YYYY-MM-DD format
func getCurrentDate() string {
	return time.Now().Format("2006-01-02")
}

// updateJSONFile atomically updates the JSON file with a new result
func updateJSONFile(outputFile string, newResult DomainResult, mutex *sync.Mutex) error {
	mutex.Lock()
	defer mutex.Unlock()

	var results []DomainResult

	// Read existing file if it exists
	if _, err := os.Stat(outputFile); err == nil {
		// File exists, read and parse it
		data, err := os.ReadFile(outputFile)
		if err != nil {
			return fmt.Errorf("error reading existing JSON file: %v", err)
		}

		// Parse existing JSON (handle empty file)
		if len(data) > 0 {
			if err := json.Unmarshal(data, &results); err != nil {
				// If parse fails, start fresh
				results = []DomainResult{}
			}
		}
	}

	// Check if result with same URL already exists and update it, otherwise append
	found := false
	for i, result := range results {
		if result.Name == newResult.Name {
			results[i] = newResult
			found = true
			break
		}
	}

	if !found {
		results = append(results, newResult)
	}

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(results, "", "    ")
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	// Write to temporary file first for atomic operation
	tmpFile := outputFile + ".tmp"
	err = os.WriteFile(tmpFile, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error writing temporary file: %v", err)
	}

	// Atomically rename temp file to output file
	err = os.Rename(tmpFile, outputFile)
	if err != nil {
		// Clean up temp file on error
		os.Remove(tmpFile)
		return fmt.Errorf("error renaming temporary file: %v", err)
	}

	return nil
}

// loadKeywords loads keywords from either a file or a comma-separated string
func loadKeywords(wordlist string) ([]string, error) {
	var keywords []string

	// Check if wordlist is a file path (file exists)
	if _, err := os.Stat(wordlist); err == nil {
		// File exists, read from file
		file, err := os.Open(wordlist)
		if err != nil {
			return nil, fmt.Errorf("error opening keyword file: %v", err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			keyword := strings.TrimSpace(scanner.Text())
			if keyword != "" {
				keywords = append(keywords, keyword)
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("error reading keyword file: %v", err)
		}
	} else {
		// File doesn't exist, treat as comma-separated string
		parts := strings.Split(wordlist, ",")
		for _, part := range parts {
			keyword := strings.TrimSpace(part)
			if keyword != "" {
				keywords = append(keywords, keyword)
			}
		}
	}

	if len(keywords) == 0 {
		return nil, fmt.Errorf("no keywords found in wordlist")
	}

	return keywords, nil
}

func main() {
	// Initialize flags
	wordlist := pflag.StringP("wordlist", "w", "keywords.txt", "Path to the file containing keywords to check, or comma-separated keywords (e.g., \"Stranger Things,The Family Man\")")
	output := pflag.StringP("output", "o", "programs.json", "Path to the output JSON file")
	timeout := pflag.IntP("timeout", "t", 30, "Timeout for each HTTP request in seconds")
	concurrency := pflag.IntP("concurrency", "c", 50, "Number of concurrent workers")
	silent := pflag.Bool("silent", false, "Silent mode")
	versionFlag := pflag.Bool("version", false, "Print the version of the tool and exit")
	verbose := pflag.Bool("verbose", false, "Enable verbose logging")

	pflag.Parse()

	// Set logging level
	logrus.SetOutput(os.Stdout)
	if *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	if *versionFlag {
		banner.PrintBanner()
		banner.PrintVersion()
		return
	}

	if !*silent {
		banner.PrintBanner()
	}

	// Load keywords from file or comma-separated string
	keywords, err := loadKeywords(*wordlist)
	if err != nil {
		logrus.Fatalf("Error loading keywords: %v", err)
	}

	// Read URLs from standard input
	var urls []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url != "" {
			urls = append(urls, url)
		}
	}
	if err := scanner.Err(); err != nil {
		logrus.Fatalf("Error reading input: %v", err)
	}

	if len(urls) == 0 {
		logrus.Warn("No URLs provided via stdin")
		return
	}

	// Initialize output file with empty JSON array if it doesn't exist
	if _, err := os.Stat(*output); os.IsNotExist(err) {
		emptyJSON := []byte("[]\n")
		if err := os.WriteFile(*output, emptyJSON, 0644); err != nil {
			logrus.Warnf("Warning: Could not initialize output file: %v", err)
		}
	}

	// Process domains
	timeoutDuration := time.Duration(*timeout) * time.Second
	var wg sync.WaitGroup
	sem := make(chan struct{}, *concurrency)
	fileMutex := &sync.Mutex{}

	for _, url := range urls {
		wg.Add(1)
		sem <- struct{}{}

		go func(domainURL string) {
			defer func() {
				<-sem
				wg.Done()
			}()

			data, err := processDomain(domainURL, keywords, timeoutDuration, *concurrency)
			if err != nil {
				if *verbose {
					logrus.Warnf("Error processing %s: %v", domainURL, err)
				}
				return // Skip this domain
			}

			// Select shortest favicon
			logo := selectShortestFavicon(data.Favicons)
			if logo == "" {
				logo = "-"
			}

			// Prepare emails
			emails := data.Emails
			if len(emails) == 0 {
				emails = []string{"-"}
			}

			// Prepare internal links
			internalLinks := data.InternalLinks
			if len(internalLinks) == 0 {
				internalLinks = []string{}
			}

			// Create result
			result := DomainResult{
				Name:         data.URL,
				Logo:         logo,
				InternalLink: internalLinks,
				Email:        emails,
				Matched:      data.Keywords,
				LastUpdated:  getCurrentDate(),
			}

			// Update JSON file immediately
			if err := updateJSONFile(*output, result, fileMutex); err != nil {
				logrus.Warnf("Error updating JSON file for %s: %v", domainURL, err)
			}

			logrus.Infof("Processed: %s", domainURL)
		}(url)
	}

	wg.Wait()

	// All results have been written to JSON file immediately as they were processed
	logrus.Infof("All domains processed. Results saved to: %s", *output)
}
