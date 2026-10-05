package transcribe

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Backend string

const (
	BackendMLX     Backend = "mlx"
	BackendWhisper Backend = "whisper"
	// BackendWhisperCpp is whisper.cpp, whose binary is named whisper-cpp.
	BackendWhisperCpp Backend = "whisper-cpp"
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
	if _, err := exec.LookPath("whisper"); err == nil {
		return BackendWhisper
	}
	if _, err := exec.LookPath("whisper-cpp"); err == nil {
		return BackendWhisperCpp
	}
	return BackendWhisper
}

// resolveCppModel returns the path to a ggml model file for whisper.cpp.
// model may be a path to a file, or a short name (e.g. "base") looked up as
// ggml-<name>.bin in $WHISPER_CPP_MODEL_DIR, then a few conventional dirs.
func resolveCppModel(model string) (string, error) {
	if st, err := os.Stat(model); err == nil && !st.IsDir() {
		return model, nil
	}
	var dirs []string
	if d := os.Getenv("WHISPER_CPP_MODEL_DIR"); d != "" {
		dirs = append(dirs, d)
	}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs,
			filepath.Join(home, ".cache", "whisper-cpp"),
			filepath.Join(home, ".local", "share", "whisper-cpp"),
		)
	}
	dirs = append(dirs, "/opt/homebrew/share/whisper-cpp", "/usr/local/share/whisper-cpp")

	name := "ggml-" + model + ".bin"
	for _, d := range dirs {
		p := filepath.Join(d, name)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("whisper.cpp model %q not found: pass --model /path/to/%s or set WHISPER_CPP_MODEL_DIR", model, name)
}

// runCpp extracts 16kHz mono audio with ffmpeg (whisper.cpp cannot read video
// containers) and runs whisper-cpp on it, writing <outputDir>/<stem>.vtt.
func runCpp(videoPath, outputDir, model, language string) error {
	modelPath, err := resolveCppModel(model)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp("", "vidprep-*.wav")
	if err != nil {
		return err
	}
	tmp.Close()
	defer os.Remove(tmp.Name())

	ff := exec.Command("ffmpeg", "-y", "-loglevel", "error", "-i", videoPath,
		"-vn", "-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le", "-f", "wav", tmp.Name())
	ff.Stderr = os.Stderr
	if err := ff.Run(); err != nil {
		return fmt.Errorf("ffmpeg audio extraction: %w", err)
	}

	stem := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))
	cmd := exec.Command("whisper-cpp", "-m", modelPath, "-f", tmp.Name(),
		"-ovtt", "-of", filepath.Join(outputDir, stem), "-l", language)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Run transcribes videoPath and writes a .vtt file into outputDir.
// mlx_whisper, whisper and whisper-cpp all name the output <stem>.vtt.
func Run(videoPath, outputDir, model, language string, backend Backend) error {
	var cmd *exec.Cmd
	switch backend {
	case BackendWhisperCpp:
		return runCpp(videoPath, outputDir, model, language)
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
