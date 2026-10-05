package scanner

import (
	"os"
	"path/filepath"
	"strings"
)

var videoMIME = map[string]string{
	".mp4":  "video/mp4",
	".mov":  "video/quicktime",
	".mkv":  "video/x-matroska",
	".webm": "video/webm",
	".avi":  "video/x-msvideo",
}

func FindVideos(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var videos []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if _, ok := videoMIME[ext]; ok {
			videos = append(videos, filepath.Join(dir, e.Name()))
		}
	}
	return videos, nil
}

func MIMEType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if m, ok := videoMIME[ext]; ok {
		return m
	}
	return "video/mp4"
}
