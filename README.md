## stoppiracy

A powerful Go tool that combines the functionality of `nsfwdetector`, `favinfo`, and `emailextractor` into a single unified tool. Process domains efficiently with just **one HTTP request** per domain while collecting keywords, favicons, and emails simultaneously.

## Why stoppiracy?

- **Single Request Optimization**: Instead of making 3 separate requests per domain, stoppiracy makes just 1 request and extracts all information from the same HTML response
- **Unified Tool**: No need to chain multiple tools together - get all data in one go
- **Efficient**: Reduces network overhead and processing time significantly
- **Comprehensive**: Collects keywords, favicons, and emails from domains and their internal links

## Features
- ✅ **Keyword Detection**: Check if domains contain specific keywords (e.g., movie/show titles) in HTML content and URL paths
- ✅ **Internal Link Tracking**: Tracks and saves internal links where keywords were matched
- ✅ **Favicon Extraction**: Automatically extracts favicon URLs and selects the shortest one
- ✅ **Email Extraction**: Crawls domains and internal links to find all unique email addresses
- ✅ **Flexible Wordlist**: Support for file-based or comma-separated keyword lists
- ✅ **Concurrent Processing**: Process multiple domains simultaneously
- ✅ **Real-time Output**: Saves results to JSON file immediately after each domain is processed
- ✅ **JSON Output**: Clean, structured JSON output for easy integration

## Installation

### Using Go Install
```
go install github.com/rix4uni/stoppiracy@latest
```

### Download Prebuilt Binaries
```
wget https://github.com/rix4uni/stoppiracy/releases/download/v0.0.2/stoppiracy-linux-amd64-0.0.2.tgz
tar -xvzf stoppiracy-linux-amd64-0.0.2.tgz
rm -rf stoppiracy-linux-amd64-0.0.2.tgz
mv stoppiracy ~/go/bin/stoppiracy
```
Or download [binary release](https://github.com/rix4uni/stoppiracy/releases) for your platform.

### Build from Source
```
git clone --depth 1 https://github.com/rix4uni/stoppiracy.git
cd stoppiracy; go install
```

## Usage
```yaml
Usage of stoppiracy:
  -c, --concurrency int   Number of concurrent workers (default 50)
  -o, --output string     Path to the output JSON file (default "programs.json")
      --silent            Silent mode
  -t, --timeout int       Timeout for each HTTP request in seconds (default 30)
      --verbose           Enable verbose logging
      --version           Print the version of the tool and exit
  -w, --wordlist string   Path to the file containing keywords to check, or comma-separated keywords (e.g., "Stranger Things,The Family Man") (default "keywords.txt")
```

## Usage Examples

### Basic Usage
```yaml
cat targets.txt | stoppiracy --wordlist keywords.txt
```

### With Comma-Separated Keywords
```yaml
echo "https://example.com" | stoppiracy --wordlist "Stranger Things,The Family Man"
```

### Single Keyword
```yaml
echo "https://example.com" | stoppiracy --wordlist "Stranger Things"
```

### Silent Mode
```yaml
cat targets.txt | stoppiracy --silent --wordlist keywords.txt
```

## Command-Line Flags
| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--wordlist` | `-w` | `keywords.txt` | Path to the file containing keywords to check, or comma-separated keywords (e.g., `"Stranger Things,The Family Man"`) |
| `--output` | `-o` | `programs.json` | Path to the output JSON file |
| `--timeout` | `-t` | `30` | Timeout for each HTTP request in seconds |
| `--concurrency` | `-c` | `50` | Number of concurrent workers |
| `--silent` | | `false` | Silent mode (no banner output) |
| `--verbose` | | `false` | Enable verbose logging |
| `--version` | | `false` | Print the version of the tool and exit |

## Wordlist Formats

The `--wordlist` flag supports three formats:

### 1. File Path (One keyword per line)
```yaml
stoppiracy --wordlist keywords.txt
```

**keywords.txt:**
```yaml
Stranger Things
The Family Man
Iron Man
```

### 2. Single Keyword
```yaml
stoppiracy --wordlist "Stranger Things"
```

### 3. Comma-Separated Keywords
```yaml
stoppiracy --wordlist "Stranger Things,The Family Man"
```

Or with spaces:

```yaml
stoppiracy --wordlist "Stranger Things, The Family Man, Iron Man"
```

**Note**: If the wordlist value is a valid file path, it will be read as a file. Otherwise, it will be treated as a comma-separated string.

## Output Format
The tool outputs a JSON array to the specified output file (default: `programs.json`):

```json
[
    {
        "name": "https://www.glia.com",
        "logo": "https://cdn.prod.website-files.com/680f1550811d9719bdbcf21b/6835fd1ea895df4327eae39d_favicon%20(4).png",
        "internal_link": [
            "https://www.glia.com/security-bounty"
        ],
        "email": [
            "bugs@glia.com",
            "privacy@glia.com",
            "bounty-hunters@glia.atlassian.net"
        ],
        "matched": [
            "We offer reward",
            "eligible for a reward"
        ],
        "last_updated": "2025-11-29"
    }
]
```

### Field Descriptions
- **name**: The domain URL that was processed
- **logo**: The shortest favicon URL found (or `"-"` if none found)
- **internal_link**: Array of internal links where keywords were matched (includes base domain URL if keywords matched on main page, or empty array `[]` if no matches)
- **email**: Array of unique email addresses found (or `["-"]` if none found)
- **matched**: Array of keywords that were found in the domain content (HTML or URL paths)
- **last_updated**: Current date in `YYYY-MM-DD` format

## Examples

### Example 1: Process Multiple Domains

**targets.txt:**
```yaml
https://desiremovies.group
https://vegamovies.select
```

**Command:**
```yaml
cat targets.txt | stoppiracy --wordlist keywords.txt --output results.json
```

### Example 2: Using Inline Keywords
```yaml
echo "https://example.com" | stoppiracy --wordlist "Stranger Things,The Family Man,Iron Man" --silent
```

### Example 3: Custom Timeout and Concurrency
```yaml
cat domains.txt | stoppiracy \
  --wordlist keywords.txt \
  --timeout 60 \
  --concurrency 100 \
  --output results.json
```

### Example 4: Verbose Mode for Debugging
```yaml
echo "https://example.com" | stoppiracy --wordlist keywords.txt --verbose
```

### Example 5: Bug Bounty Program Discovery
```yaml
echo "www.glia.com" | stoppiracy --silent --wordlist "Eligible Targets,we offer a monetary,We offer reward,monetary reward,eligible for a reward,we award a bounty,We offer monetary rewards"
```

## Use Cases

While originally designed for detecting piracy movie websites, stoppiracy is a versatile web intelligence tool that can be used for various purposes:

### 1. Bug Bounty Program Discovery
Discover websites with active bug bounty programs:
```yaml
echo "targets.txt" | stoppiracy --wordlist "Eligible Targets,we offer a monetary,We offer reward,monetary reward,eligible for a reward,we award a bounty,We offer monetary rewards"
```

### 2. Security & Vulnerability Research
Find websites running vulnerable software or security advisories:
```yaml
echo "targets.txt" | stoppiracy --wordlist "WordPress 4.0,jQuery 1.11,Apache 2.2"
echo "targets.txt" | stoppiracy --wordlist "security advisory,vulnerability disclosure,CVE-2024"
```

### 3. Brand Monitoring & Trademark Protection
Find unauthorized use of brand names or counterfeit product sites:
```yaml
echo "domains.txt" | stoppiracy --wordlist "Nike Sneakers,Apple iPhone,Samsung Galaxy"
echo "suspicious-sites.txt" | stoppiracy --wordlist "replica,counterfeit,knockoff"
```

### 4. Job Board Aggregation
Find job postings with specific keywords:
```yaml
echo "job-sites.txt" | stoppiracy --wordlist "remote work,work from home,hiring now,apply now"
echo "tech-job-boards.txt" | stoppiracy --wordlist "Python Developer,DevOps Engineer,Full Stack"
```

### 5. Competitor Intelligence
Find competitors mentioning your products/services or monitor announcements:
```yaml
echo "competitor-list.txt" | stoppiracy --wordlist "your-product-name,your-brand,alternative to"
echo "competitors.txt" | stoppiracy --wordlist "new product launch,announcement,coming soon"
```

### 6. Content Discovery & Research
Find educational resources or research papers on specific topics:
```yaml
echo "edu-sites.txt" | stoppiracy --wordlist "tutorial,learn,how to,course"
echo "academic-sites.txt" | stoppiracy --wordlist "research paper,study,publication"
```

### 7. Affiliate Program Discovery
Find websites with affiliate programs:
```yaml
echo "domains.txt" | stoppiracy --wordlist "affiliate program,earn commission,partner with us"
```

### 8. API & Developer Resource Discovery
Find websites offering APIs or open source projects:
```yaml
echo "dev-sites.txt" | stoppiracy --wordlist "API documentation,developer portal,API key"
echo "github-proxies.txt" | stoppiracy --wordlist "open source,contribute,star us on GitHub"
```

### 9. Event Discovery
Find conferences, events, or webinars:
```yaml
echo "event-sites.txt" | stoppiracy --wordlist "conference 2024,register now,buy tickets"
echo "tech-sites.txt" | stoppiracy --wordlist "webinar,join us live,free event"
```

### 10. Compliance & Certification Checking
Find websites with specific certifications or compliance mentions:
```yaml
echo "company-sites.txt" | stoppiracy --wordlist "ISO 27001,SOC 2,GDPR compliant"
echo "sites.txt" | stoppiracy --wordlist "privacy policy,data protection,GDPR"
```

### 11. Technology Stack Detection
Find sites using specific frameworks or hosting platforms:
```yaml
echo "websites.txt" | stoppiracy --wordlist "React,Vue.js,Angular,Django,Flask"
echo "domains.txt" | stoppiracy --wordlist "powered by WordPress,built with Shopify"
```

### 12. News & Media Monitoring
Find news articles or track mentions of specific topics:
```yaml
echo "news-sites.txt" | stoppiracy --wordlist "breaking news,exclusive,latest update"
echo "media-outlets.txt" | stoppiracy --wordlist "your-topic-name,trending now"
```

### 13. Legal & Compliance Monitoring
Find terms of service, copyright notices, or legal pages:
```yaml
echo "domains.txt" | stoppiracy --wordlist "terms of service,user agreement,legal"
echo "sites.txt" | stoppiracy --wordlist "copyright notice,all rights reserved"
```

### 14. Lead Generation
Extract emails from industry-specific sites:
```yaml
echo "industry-sites.txt" | stoppiracy --wordlist "contact us,sales,get in touch" --output leads.json
```

### 15. Domain Portfolio Analysis
Check which domains have specific content:
```yaml
echo "domain-portfolio.txt" | stoppiracy --wordlist "under construction,coming soon,for sale"
```

### 16. Content Moderation
Find inappropriate content:
```yaml
echo "user-submitted-sites.txt" | stoppiracy --wordlist "inappropriate-keywords-list"
```

## Real-time Output

stoppiracy saves results to the JSON output file **immediately** after each domain is processed, not just at the end. This means:

- ✅ **Monitor Progress**: Watch results appear in real-time as domains are processed
- ✅ **Early Access**: Start analyzing results before all domains finish processing
- ✅ **Fault Tolerance**: If the process is interrupted, completed results are already saved
- ✅ **Live Updates**: The output file is updated atomically after each domain completes

The JSON file is updated atomically (using temporary files) to ensure data integrity even with concurrent processing.

## How It Works
1. **Single Request**: Makes one HTTP request to fetch the main page
2. **Keyword Detection**: Scans the HTML content and URL paths for matched keywords
3. **Favicon Extraction**: Extracts all favicon URLs from HTML and selects the shortest one
4. **Email Extraction**: Extracts emails from the main page
5. **Internal Link Crawling**: Crawls internal links concurrently to find additional emails and keywords
6. **Internal Link Tracking**: Tracks which internal links contain matched keywords (in URL path or HTML content)
7. **Real-time Output**: Saves results to JSON file immediately after each domain is processed
8. **JSON Output**: Combines all results into a structured JSON format

## Favicon Selection Logic
When multiple favicon URLs are found:
- The tool selects the **shortest URL** (by character length)
- If multiple URLs have the same shortest length, the **first one** is selected

## Email Extraction
The tool extracts emails from:
- The main domain page
- Internal links found on the page (crawled concurrently)
- Mailto links
- Plain text content

All emails are deduplicated and validated before being included in the output.

## Keyword Matching
The tool matches keywords in two ways:
- **HTML Content**: Searches for keywords within the HTML content of pages
- **URL Paths**: Checks if keywords appear in the URL path (e.g., `/home/`, `/series`)

Keywords are matched on:
- The main domain page
- Internal links (both in their HTML content and URL paths)

Internal links that match keywords (either in URL or HTML) are tracked and saved in the `internal_link` field. The base domain URL is included in `internal_link` only if keywords matched on the main page.

## Error Handling
- Domains that fail to fetch are **skipped** (not included in output)
- Domains with no emails will have `["-"]` in the email field
- Domains with no matched keywords will have an empty `[]` array in the `matched` field
- Domains with no matched internal links will have an empty `[]` array in the `internal_link` field
- Domains with no favicons will have `"-"` in the logo field

## Performance
- **Optimized**: Uses only 1 HTTP request per domain (instead of 3)
- **Concurrent**: Processes multiple domains simultaneously
- **Efficient**: Reuses HTML content for all extraction operations

## Related Tools
- [stoppiracy](https://github.com/rix4uni/stoppiracy) - NSFW content detection
- [favinfo](https://github.com/rix4uni/favinfo) - Favicon information extraction
- [emailextractor](https://github.com/rix4uni/emailextractor) - Email extraction tool
