package vtt

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type Cue struct {
	Start time.Duration
	End   time.Duration
	Text  string
}

func Parse(path string) ([]Cue, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cues []Cue
	var current *Cue
	sc := bufio.NewScanner(f)

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())

		if line == "" {
			if current != nil && current.Text != "" {
				cues = append(cues, *current)
				current = nil
			}
			continue
		}
		if line == "WEBVTT" {
			continue
		}
		if strings.Contains(line, " --> ") {
			parts := strings.SplitN(line, " --> ", 2)
			// strip any position metadata after the end timestamp
			endPart := strings.Fields(parts[1])[0]
			current = &Cue{
				Start: parseTimestamp(strings.TrimSpace(parts[0])),
				End:   parseTimestamp(endPart),
			}
			continue
		}
		// cue identifier line (number or text before timing)
		if current == nil {
			continue
		}
		if current.Text != "" {
			current.Text += " "
		}
		current.Text += line
	}
	if current != nil && current.Text != "" {
		cues = append(cues, *current)
	}
	return cues, sc.Err()
}

func parseTimestamp(s string) time.Duration {
	// HH:MM:SS.mmm or MM:SS.mmm
	var h, m, sec, ms int
	parts := strings.Split(s, ":")
	switch len(parts) {
	case 3:
		fmt.Sscanf(parts[0], "%d", &h)
		fmt.Sscanf(parts[1], "%d", &m)
		parseSecMs(parts[2], &sec, &ms)
	case 2:
		fmt.Sscanf(parts[0], "%d", &m)
		parseSecMs(parts[1], &sec, &ms)
	}
	return time.Duration(h)*time.Hour +
		time.Duration(m)*time.Minute +
		time.Duration(sec)*time.Second +
		time.Duration(ms)*time.Millisecond
}

func parseSecMs(s string, sec, ms *int) {
	p := strings.SplitN(s, ".", 2)
	fmt.Sscanf(p[0], "%d", sec)
	if len(p) > 1 {
		// pad/truncate to 3 digits
		msStr := p[1]
		for len(msStr) < 3 {
			msStr += "0"
		}
		fmt.Sscanf(msStr[:3], "%d", ms)
	}
}
