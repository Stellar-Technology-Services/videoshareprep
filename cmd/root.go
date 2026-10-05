package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Stellar-Technology-Services/vidprep/internal/htmlgen"
	"github.com/Stellar-Technology-Services/vidprep/internal/markdown"
	"github.com/Stellar-Technology-Services/vidprep/internal/scanner"
	"github.com/Stellar-Technology-Services/vidprep/internal/transcribe"
	"github.com/Stellar-Technology-Services/vidprep/internal/vtt"
)

var (
	flagBackend  string
	flagModel    string
	flagForce    bool
	flagDryRun   bool
	flagLanguage string
)

var rootCmd = &cobra.Command{
	Use:   "vidprep [directory]",
	Short: "Transcribe videos to VTT, Markdown, and HTML",
	Long: `vidprep scans a directory for video files and for each one:
  1. Transcribes it to .vtt (WebVTT) using mlx_whisper or whisper
  2. Converts .vtt to a .md file with title and transcript
  3. Generates a .html page with embedded video, closed captions, and description`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCmd,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVar(&flagBackend, "backend", "auto", "transcription backend: auto|mlx|whisper")
	rootCmd.Flags().StringVar(&flagModel, "model", "base", "model size: tiny|base|small|medium|large")
	rootCmd.Flags().BoolVar(&flagForce, "force", false, "regenerate files even if they already exist")
	rootCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "print planned actions without writing any files")
	rootCmd.Flags().StringVar(&flagLanguage, "language", "auto", "language code (e.g. en, fr, ja) or auto")
}

func runCmd(_ *cobra.Command, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	backend := transcribe.Backend(flagBackend)
	if flagBackend == "auto" {
		backend = transcribe.DetectBackend()
		fmt.Printf("detected backend: %s\n", backend)
	}

	videos, err := scanner.FindVideos(dir)
	if err != nil {
		return err
	}
	if len(videos) == 0 {
		fmt.Println("no video files found in", dir)
		return nil
	}
	fmt.Printf("found %d video(s)\n", len(videos))

	for _, v := range videos {
		if err := processVideo(v, backend); err != nil {
			fmt.Fprintf(os.Stderr, "✗ %s: %v\n", filepath.Base(v), err)
		}
	}
	return nil
}

func processVideo(videoPath string, backend transcribe.Backend) error {
	stem := strings.TrimSuffix(videoPath, filepath.Ext(videoPath))
	vttPath  := stem + ".vtt"
	mdPath   := stem + ".md"
	htmlPath := stem + ".html"

	fmt.Printf("\n▶ %s\n", filepath.Base(videoPath))

	// 1. Transcribe → .vtt
	if flagForce || !exists(vttPath) {
		if flagDryRun {
			fmt.Printf("  [dry-run] transcribe → %s\n", filepath.Base(vttPath))
		} else {
			fmt.Printf("  transcribing → %s\n", filepath.Base(vttPath))
			if err := transcribe.Run(videoPath, filepath.Dir(videoPath), flagModel, flagLanguage, backend); err != nil {
				return fmt.Errorf("transcription: %w", err)
			}
		}
	} else {
		fmt.Printf("  skip (exists) %s\n", filepath.Base(vttPath))
	}

	// 2. Parse .vtt → .md
	if flagForce || !exists(mdPath) {
		if flagDryRun {
			fmt.Printf("  [dry-run] generate → %s\n", filepath.Base(mdPath))
		} else {
			cues, err := vtt.Parse(vttPath)
			if err != nil {
				return fmt.Errorf("vtt parse: %w", err)
			}
			fmt.Printf("  generating → %s (%d cues)\n", filepath.Base(mdPath), len(cues))
			if err := markdown.Generate(videoPath, cues, mdPath); err != nil {
				return fmt.Errorf("markdown: %w", err)
			}
		}
	} else {
		fmt.Printf("  skip (exists) %s\n", filepath.Base(mdPath))
	}

	// 3. .md + .vtt → .html
	if flagForce || !exists(htmlPath) {
		if flagDryRun {
			fmt.Printf("  [dry-run] generate → %s\n", filepath.Base(htmlPath))
		} else {
			mime := scanner.MIMEType(videoPath)
			fmt.Printf("  generating → %s\n", filepath.Base(htmlPath))
			if err := htmlgen.Generate(videoPath, vttPath, mdPath, htmlPath, mime); err != nil {
				return fmt.Errorf("html: %w", err)
			}
		}
	} else {
		fmt.Printf("  skip (exists) %s\n", filepath.Base(htmlPath))
	}

	fmt.Printf("  ✓ done\n")
	return nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
