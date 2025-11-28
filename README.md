## stoppiracy

A powerful Go tool that combines the functionality of `nsfwdetector`, `favinfo`, and `emailextractor` into a single unified tool. Process domains efficiently with just **one HTTP request** per domain while collecting keywords, favicons, and emails simultaneously.

## Why stoppiracy?

- **Single Request Optimization**: Instead of making 3 separate requests per domain, stoppiracy makes just 1 request and extracts all information from the same HTML response
- **Unified Tool**: No need to chain multiple tools together - get all data in one go
- **Efficient**: Reduces network overhead and processing time significantly
- **Comprehensive**: Collects keywords, favicons, and emails from domains and their internal links

## Features
- ✅ **Keyword Detection**: Check if domains contain specific keywords (e.g., movie/show titles)
- ✅ **Favicon Extraction**: Automatically extracts favicon URLs and selects the shortest one
- ✅ **Email Extraction**: Crawls domains and internal links to find all unique email addresses
- ✅ **Flexible Wordlist**: Support for file-based or comma-separated keyword lists
- ✅ **Concurrent Processing**: Process multiple domains simultaneously
- ✅ **JSON Output**: Clean, structured JSON output for easy integration

## Installation

### Using Go Install
```
go install github.com/rix4uni/stoppiracy@latest
```

### Download Prebuilt Binaries
```
wget https://github.com/rix4uni/stoppiracy/releases/download/v0.0.1/stoppiracy-linux-amd64-0.0.1.tgz
tar -xvzf stoppiracy-linux-amd64-0.0.1.tgz
rm -rf stoppiracy-linux-amd64-0.0.1.tgz
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
        "name": "https://example.com",
        "logo": "https://example.com/favicon.png",
        "email": [
            "contact@example.com",
            "admin@example.com"
        ],
        "matched": [
            "Stranger Things",
            "The Family Man"
        ],
        "last_updated": "2025-01-28"
    }
]
```

### Field Descriptions
- **name**: The domain URL that was processed
- **logo**: The shortest favicon URL found (or `"-"` if none found)
- **email**: Array of unique email addresses found (or `["-"]` if none found)
- **matched**: Array of keywords that were found in the domain content
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

## How It Works
1. **Single Request**: Makes one HTTP request to fetch the main page
2. **Keyword Detection**: Scans the HTML content for matched keywords
3. **Favicon Extraction**: Extracts all favicon URLs from HTML and selects the shortest one
4. **Email Extraction**: Extracts emails from the main page
5. **Internal Link Crawling**: Crawls internal links concurrently to find additional emails
6. **JSON Output**: Combines all results into a structured JSON format

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

## Error Handling
- Domains that fail to fetch are **skipped** (not included in output)
- Domains with no emails will have `["-"]` in the email field
- Domains with no matched keywords will have an empty `[]` array
- Domains with no favicons will have `"-"` in the logo field

## Performance
- **Optimized**: Uses only 1 HTTP request per domain (instead of 3)
- **Concurrent**: Processes multiple domains simultaneously
- **Efficient**: Reuses HTML content for all extraction operations

## Related Tools
- [stoppiracy](https://github.com/rix4uni/stoppiracy) - NSFW content detection
- [favinfo](https://github.com/rix4uni/favinfo) - Favicon information extraction
- [emailextractor](https://github.com/rix4uni/emailextractor) - Email extraction tool
