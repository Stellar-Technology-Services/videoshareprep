package htmlgen

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
)

const htmlTmpl = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{.Title}}</title>
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      background: #111; color: #ddd;
      max-width: 900px; margin: 0 auto; padding: 2rem 1.25rem;
    }
    h1 { font-size: 1.5rem; font-weight: 600; color: #fff; margin-bottom: 1rem; }
    video {
      width: 100%; border-radius: 10px; background: #000;
      box-shadow: 0 8px 32px rgba(0,0,0,.7);
    }
    .description {
      margin-top: 1.75rem; line-height: 1.75;
      color: #bbb; font-size: 0.97rem;
    }
    .description p { margin-bottom: 1em; }
    .description h2 { font-size: 1.1rem; color: #eee; margin: 1.5em 0 0.5em; }
    .description h3 { font-size: 1rem; color: #ccc; margin: 1.2em 0 0.4em; }
    .meta { font-size: 0.85rem; color: #666; margin-bottom: 1.5rem; }
  </style>
</head>
<body>
  <h1>{{.Title}}</h1>
  <video controls preload="metadata">
    <source src="{{.VideoFile}}" type="{{.MIME}}">
    <track kind="captions" src="{{.VTTFile}}" srclang="en" label="English" default>
    Your browser does not support the video element.
  </video>
  <div class="description">
    {{.DescriptionHTML}}
  </div>
</body>
</html>`

type pageData struct {
	Title           string
	VideoFile       string
	VTTFile         string
	MIME            string
	DescriptionHTML template.HTML
}

func Generate(videoPath, vttPath, mdPath, outputPath, mime string) error {
	md, err := os.ReadFile(mdPath)
	if err != nil {
		return err
	}

	title, body := splitTitleBody(string(md))

	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(body), &buf); err != nil {
		return err
	}

	tmpl, err := template.New("page").Parse(htmlTmpl)
	if err != nil {
		return err
	}

	data := pageData{
		Title:           title,
		VideoFile:       filepath.Base(videoPath),
		VTTFile:         filepath.Base(vttPath),
		MIME:            mime,
		DescriptionHTML: template.HTML(buf.String()),
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

// splitTitleBody extracts the first # heading and returns the rest as body.
func splitTitleBody(md string) (title, body string) {
	lines := strings.Split(md, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "# ") {
			title = strings.TrimSpace(strings.TrimPrefix(line, "# "))
			body = strings.Join(lines[i+1:], "\n")
			return
		}
	}
	return "Video", md
}
