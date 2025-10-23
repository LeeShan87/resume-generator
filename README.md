# Resume Generator (Go version)

Markdown to HTML/PDF resume generator written in Go.

**Based on**: [mikepqr/resume-markdown](https://github.com/mikepqr/resume-markdown) - Python version by Mike Lee Williams

This Go port adds:

- Same goldmark parser that Hugo uses (better compatibility)
- Flexible output naming (`-o` flag)
- Shared CSS across multiple resume versions
- Improved Chrome process handling with timeout

## Features

- Converts Markdown to HTML using Goldmark (same parser as Hugo)
- Generates PDF via Chrome/Chromium headless
- CSS styling support
- Compatible with the Python version's output format
- **YAML frontmatter stripping** (Hugo/Jekyll compatible)

## Demo

Check the `example/` directory for:

- `resume.md` - Example markdown resume
- `resume.css` - Default styling
- `resume.html` - Generated HTML output
- `resume.pdf` - Generated PDF output

## Installation

```bash
go mod download
go build -o resume-generator resume.go
```

## Usage

Basic usage:

```bash
./resume-generator resume.md
```

This will generate:

- `resume.html`
- `resume.pdf`

### Options

```bash
./resume-generator [options] [input.md]

Options:
  -o, -output NAME  Output filename (without extension)
  -css PATH         Path to CSS file (default: resume.css)
  -no-html          Do not write HTML output
  -no-pdf           Do not write PDF output
  -chrome-path PATH Path to Chrome/Chromium executable
  -q                Quiet mode
  -debug            Debug mode
```

### Examples

Generate with custom output name:

```bash
./resume-generator -o my-resume resume.md
# Creates: my-resume.html and my-resume.pdf
```

Generate multiple versions from one CSS:

```bash
./resume-generator -o resume-v1 version1.md
./resume-generator -o resume-v2 version2.md
# Both use the same resume.css by default
```

Use custom CSS file:

```bash
./resume-generator -css custom-style.css -o output resume.md
```

Generate only PDF:

```bash
./resume-generator -no-html resume.md
```

Generate only HTML:

```bash
./resume-generator -no-pdf resume.md
```

Use custom Chrome path:

```bash
./resume-generator -chrome-path /path/to/chrome resume.md
```

## Requirements

- Go 1.21 or later
- Chrome or Chromium (for PDF generation)

The tool will automatically detect Chrome/Chromium on:

- macOS: `/Applications/Google Chrome.app/...`
- Linux: `/usr/bin/google-chrome`, `/usr/bin/chromium`, etc.
- Windows: `C:\Program Files\Google\Chrome\...`

## File Structure

- `resume.md` - Your resume in Markdown format
- `resume.css` - Styling for the resume (optional)
- Output: `resume.html` and `resume.pdf`

## Markdown Format

The tool expects:

- First `#` heading becomes the HTML title
- Standard Markdown syntax
- GitHub Flavored Markdown (GFM) support
- Typographic improvements (smart quotes, dashes, etc.)
- **YAML frontmatter** is automatically stripped (Hugo/Jekyll compatible)

See `example/resume.md` for an example format.

## Hugo Integration

This tool works great with Hugo static sites! See `example/Makefile.example` for a complete Hugo integration setup.

### Quick Setup

1. Copy this tool to your Hugo project:

   ```bash
   cd your-hugo-project
   git clone https://github.com/yourusername/resume-generator.git
   # or add as git submodule
   ```

2. Copy the example Makefile:

   ```bash
   cp resume-generator/example/Makefile.example ./Makefile
   ```

3. Edit the Makefile and set your name:

   ```makefile
   CV_NAME = Your_Name_CV
   ```

4. Create CV markdown files in `content/cv/`:

   ```bash
   mkdir -p content/cv
   # Create your CV markdown files here
   ```

5. Generate CVs:
   ```bash
   make cv
   ```

Your PDFs will be generated in `static/cv/<version>/Your_Name_CV.pdf`

### Hugo Frontmatter Support

The generator automatically strips YAML frontmatter, so your markdown can have Hugo metadata:

```markdown
---
title: "Resume - Full Version"
type: "cv"
pdf: "/cv/full/Your_Name_CV.pdf"
weight: 1
---

# Your Name

Your resume content here...
```

## Maintenance

**Note:** This is a personal tool that I maintain for my own use. Updates happen when I need new features or fixes. Feel free to fork and customize for your needs!

If you find bugs or have suggestions, open an issue. PRs are welcome, but response time may vary.

## License

See LICENSE file for details.
