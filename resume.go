package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

const htmlPreamble = `<html lang="en">
<head>
<meta charset="UTF-8">
<title>%s</title>
<style>
%s
</style>
</head>
<body>
<div id="resume">
`

const htmlPostamble = `</div>
</body>
</html>
`

var (
	noHTML     bool
	noPDF      bool
	chromePath string
	cssFile    string
	outputName string
	quiet      bool
	debug      bool
)

// Chrome/Chromium path guesses by platform
var chromeGuesses = map[string][]string{
	"darwin": {
		"/Applications/Chromium.app/Contents/MacOS/Chromium",
		"/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary",
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
	},
	"windows": {
		`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files\Google\Chrome\Application\chrome.exe`,
	},
	"linux": {
		"/usr/bin/google-chrome",
		"/usr/bin/chrome",
		"/usr/bin/chromium",
		"/usr/bin/chromium-browser",
		"/usr/local/bin/chrome",
		"/usr/local/bin/chromium",
	},
}

func main() {
	flag.BoolVar(&noHTML, "no-html", false, "Do not write HTML output")
	flag.BoolVar(&noPDF, "no-pdf", false, "Do not write PDF output")
	flag.StringVar(&chromePath, "chrome-path", "", "Path to Chrome or Chromium executable")
	flag.StringVar(&cssFile, "css", "", "Path to CSS file (default: <input>.css or resume.css)")
	flag.StringVar(&outputName, "o", "", "Output filename (without extension)")
	flag.StringVar(&outputName, "output", "", "Output filename (without extension)")
	flag.BoolVar(&quiet, "q", false, "Quiet mode")
	flag.BoolVar(&debug, "debug", false, "Debug mode")
	flag.Parse()

	args := flag.Args()
	inputFile := "resume.md"
	if len(args) > 0 {
		inputFile = args[0]
	}

	// Setup logging
	if quiet {
		log.SetOutput(io.Discard)
	}

	// Read markdown file
	mdContent, err := os.ReadFile(inputFile)
	if err != nil {
		log.Fatalf("Failed to read file %s: %v", inputFile, err)
	}

	// Get prefix (filename without extension)
	prefix := strings.TrimSuffix(inputFile, filepath.Ext(inputFile))

	// Use custom output name if provided
	if outputName != "" {
		prefix = outputName
	}

	// Generate HTML
	htmlContent, err := makeHTML(string(mdContent))
	if err != nil {
		log.Fatalf("Failed to generate HTML: %v", err)
	}

	// Write HTML if requested
	if !noHTML {
		htmlFile := prefix + ".html"
		if err := os.WriteFile(htmlFile, []byte(htmlContent), 0644); err != nil {
			log.Fatalf("Failed to write HTML file: %v", err)
		}
		log.Printf("Wrote %s", htmlFile)
	}

	// Generate PDF if requested
	if !noPDF {
		if err := writePDF(htmlContent, prefix); err != nil {
			log.Fatalf("Failed to generate PDF: %v", err)
		}
		log.Printf("Wrote %s.pdf", prefix)
	}
}

// stripFrontmatter removes YAML frontmatter from markdown content
func stripFrontmatter(md string) string {
	// Check if markdown starts with ---
	if !strings.HasPrefix(md, "---") {
		return md
	}

	// Find the closing ---
	lines := strings.Split(md, "\n")
	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endIdx = i
			break
		}
	}

	// If we found the closing ---, return content after it
	if endIdx > 0 && endIdx < len(lines)-1 {
		return strings.Join(lines[endIdx+1:], "\n")
	}

	// Otherwise return original
	return md
}

// extractTitle extracts the first H1 heading from markdown
func extractTitle(md string) string {
	re := regexp.MustCompile(`(?m)^#[^#]\s*(.+)$`)
	matches := re.FindStringSubmatch(md)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return "Resume"
}

// makeHTML converts markdown to HTML with CSS styling
func makeHTML(md string) (string, error) {
	// Strip Hugo/Jekyll style frontmatter
	md = stripFrontmatter(md)
	var cssContent []byte
	var err error

	// Use custom CSS if provided, otherwise default to resume.css
	cssPath := "resume.css"
	if cssFile != "" {
		cssPath = cssFile
	}

	cssContent, err = os.ReadFile(cssPath)
	if err != nil {
		log.Printf("Warning: Could not read CSS file %s: %v", cssPath, err)
		log.Println("Output will be unstyled.")
		cssContent = []byte("")
	}

	// Extract title
	title := extractTitle(md)

	// Convert markdown to HTML using goldmark (same as Hugo)
	gm := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Typographer,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(), // Allow raw HTML in markdown
		),
	)

	var buf strings.Builder
	if err := gm.Convert([]byte(md), &buf); err != nil {
		return "", fmt.Errorf("markdown conversion failed: %w", err)
	}

	// Build complete HTML
	html := fmt.Sprintf(htmlPreamble, title, string(cssContent))
	html += buf.String()
	html += htmlPostamble

	return html, nil
}

// findChrome attempts to locate Chrome or Chromium executable
func findChrome() (string, error) {
	if chromePath != "" {
		return chromePath, nil
	}

	platform := runtime.GOOS
	guesses, ok := chromeGuesses[platform]
	if !ok {
		guesses = chromeGuesses["linux"] // default to linux paths
	}

	for _, path := range guesses {
		if _, err := os.Stat(path); err == nil {
			log.Printf("Found Chrome/Chromium at %s", path)
			return path, nil
		}
	}

	return "", fmt.Errorf("could not find Chrome. Please set --chrome-path")
}

// writePDF generates a PDF from HTML using Chrome headless
func writePDF(htmlContent, prefix string) error {
	chrome, err := findChrome()
	if err != nil {
		return err
	}

	// Encode HTML to base64
	encoded := base64.StdEncoding.EncodeToString([]byte(htmlContent))
	dataURI := "data:text/html;base64," + encoded

	// Create temporary directory for Chrome
	tempDir, err := os.MkdirTemp("", "resume-generator-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Chrome options
	pdfPath := prefix + ".pdf"
	args := []string{
		"--no-sandbox",
		"--headless",
		"--print-to-pdf-no-header",
		"--no-pdf-header-footer",
		"--enable-logging=stderr",
		"--log-level=2",
		"--in-process-gpu",
		"--disable-gpu",
		"--disable-software-rasterizer",
		"--disable-dev-shm-usage",
		"--disable-background-networking",
		"--disable-background-timer-throttling",
		"--disable-backgrounding-occluded-windows",
		"--disable-renderer-backgrounding",
		"--run-all-compositor-stages-before-draw",
		"--crash-dumps-dir=" + tempDir,
		"--user-data-dir=" + tempDir,
		"--print-to-pdf=" + pdfPath,
		dataURI,
	}

	if debug {
		log.Printf("Running: %s %s", chrome, strings.Join(args, " "))
	}

	// Set a timeout for Chrome execution
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, chrome, args...)
	_ = cmd.Run() // Ignore error, Chrome often doesn't exit cleanly

	// Check if PDF was created successfully
	if _, err := os.Stat(pdfPath); err == nil {
		// PDF exists, we're good!
		return nil
	}

	// PDF doesn't exist, something went wrong
	return fmt.Errorf("PDF was not created at %s", pdfPath)
}
