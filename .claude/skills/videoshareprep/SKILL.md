---
name: transcribing-videos
description: Transcribes local video files to WebVTT captions, Markdown transcripts, and self-contained HTML player pages using vidprep. Use when the user wants to transcribe videos, generate closed captions, create shareable video pages, produce .vtt files, or convert local video files into HTML with embedded players.
allowed-tools: Bash
---

# Transcribing Videos with vidprep

`vidprep` scans a directory for video files and produces three output files per video:
- `video.vtt` — WebVTT closed captions
- `video.md` — Markdown transcript (`# Title` + prose paragraphs)
- `video.html` — Self-contained page: embedded player + captions + description

The `#` heading in `.md` becomes the HTML `<title>` and `<h1>`. All steps are idempotent — existing files are skipped unless `--force` is passed.

## Quick Start

```bash
# Install (once)
brew tap Stellar-Technology-Services/vidprep
brew install vidprep
pip install mlx-whisper          # Apple Silicon backend (recommended)

# Run in any directory containing videos
cd /path/to/videos
vidprep
```

## Installation

### Homebrew (recommended)

```bash
brew tap Stellar-Technology-Services/vidprep
brew install vidprep
```

Then install a transcription backend (vidprep needs one):

| Backend | Install | Best for |
|---------|---------|----------|
| mlx-whisper | `pip install mlx-whisper` | Apple Silicon — fast, local, free |
| openai-whisper | `brew install openai-whisper` | Intel Mac / Linux fallback |

vidprep auto-detects whichever is available, preferring mlx-whisper.

### Build from source

```bash
git clone https://github.com/Stellar-Technology-Services/vidprep
cd vidprep
go build -o vidprep .
# move binary to a directory on $PATH
mv vidprep /usr/local/bin/
```

## Usage

```bash
vidprep [directory] [flags]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--backend` | `auto` | `auto`, `mlx`, or `whisper` |
| `--model` | `base` | `tiny`, `base`, `small`, `medium`, `large` |
| `--language` | `auto` | Language code (`en`, `fr`, `ja`, …) or `auto` |
| `--force` | false | Regenerate files even if they already exist |
| `--dry-run` | false | Print what would be done without writing files |

## Examples

**Transcribe all videos in the current directory:**
```bash
vidprep
```

**Process a specific folder, higher accuracy model:**
```bash
vidprep ~/recordings --model small
```

**Force-regenerate everything with a known language:**
```bash
vidprep --force --language en
```

**Preview without writing anything:**
```bash
vidprep --dry-run
```

**Use openai-whisper explicitly (e.g. on Linux):**
```bash
vidprep --backend whisper
```

## Output Structure

Given `demo-walkthrough.mp4`, vidprep produces:

```
demo-walkthrough.mp4
demo-walkthrough.vtt    ← WebVTT captions with timestamps
demo-walkthrough.md     ← Markdown: # title + transcript paragraphs
demo-walkthrough.html   ← HTML page: video player + captions + description
```

The `.html` file references `.mp4` and `.vtt` by relative filename — keep all three files in the same directory when serving.

## Supported Video Formats

`.mp4`, `.mov`, `.mkv`, `.webm`, `.avi`

## Troubleshooting

**"no video files found"** — Check the directory contains a supported video format.

**mlx-whisper model download on first run** — The first transcription downloads the model from HuggingFace (~150 MB for `base`). Subsequent runs are instant.

**mlx-whisper not found, falling back to whisper** — Install `pip install mlx-whisper` for the faster Apple Silicon backend.

**VTT already exists but looks wrong** — Re-run with `--force` to regenerate, or try a larger `--model`.
