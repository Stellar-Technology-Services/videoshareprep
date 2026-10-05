package transcribe

import (
	"os"
	"os/exec"
)

type Backend string

const (
	BackendMLX     Backend = "mlx"
	BackendWhisper Backend = "whisper"
)

// mlxModels maps short model names to mlx-community HuggingFace repos.
var mlxModels = map[string]string{
	"tiny":    "mlx-community/whisper-tiny-mlx",
	"base":    "mlx-community/whisper-base-mlx",
	"small":   "mlx-community/whisper-small-mlx",
	"medium":  "mlx-community/whisper-medium-mlx",
	"large":   "mlx-community/whisper-large-v3-mlx",
	"large-v3": "mlx-community/whisper-large-v3-mlx",
}

func DetectBackend() Backend {
	if _, err := exec.LookPath("mlx_whisper"); err == nil {
		return BackendMLX
	}
	return BackendWhisper
}

// Run transcribes videoPath and writes a .vtt file into outputDir.
// Both mlx_whisper and whisper name the output <stem>.vtt.
func Run(videoPath, outputDir, model, language string, backend Backend) error {
	var cmd *exec.Cmd
	switch backend {
	case BackendMLX:
		repo := model
		if mapped, ok := mlxModels[model]; ok {
			repo = mapped
		}
		args := []string{videoPath, "--output-format", "vtt", "--output-dir", outputDir, "--model", repo}
		if language != "auto" {
			args = append(args, "--language", language)
		}
		cmd = exec.Command("mlx_whisper", args...)
	default:
		args := []string{videoPath, "--output_format", "vtt", "--output_dir", outputDir, "--model", model}
		if language != "auto" {
			args = append(args, "--language", language)
		}
		cmd = exec.Command("whisper", args...)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
