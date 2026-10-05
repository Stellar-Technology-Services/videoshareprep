package markdown

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/Stellar-Technology-Services/vidprep/internal/vtt"
)

func Generate(videoPath string, cues []vtt.Cue, outputPath string) error {
	title := titleFromPath(videoPath)
	duration := ""
	if len(cues) > 0 {
		duration = formatDuration(cues[len(cues)-1].End)
	}

	var sb strings.Builder
	sb.WriteString("# " + title + "\n\n")
	sb.WriteString(fmt.Sprintf("Transcript of %s", filepath.Base(videoPath)))
	if duration != "" {
		sb.WriteString(" · " + duration)
	}
	sb.WriteString("\n\n")

	// Group cues into paragraphs: new paragraph when gap between cues > 2s
	var para []string
	for i, c := range cues {
		para = append(para, c.Text)
		last := i == len(cues)-1
		bigGap := !last && cues[i+1].Start-c.End > 2*time.Second
		if last || bigGap {
			sb.WriteString(strings.Join(para, " ") + "\n\n")
			para = para[:0]
		}
	}

	return os.WriteFile(outputPath, []byte(sb.String()), 0644)
}

func titleFromPath(path string) string {
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")
	return toTitleCase(name)
}

func toTitleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) == 0 {
			continue
		}
		r := []rune(w)
		r[0] = unicode.ToUpper(r[0])
		words[i] = string(r)
	}
	return strings.Join(words, " ")
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}
