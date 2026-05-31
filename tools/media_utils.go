package tools

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func SaveMedia(data []byte, filename string) (string, error) {
	mediaDir := "/data/media"
	if err := os.MkdirAll(mediaDir, 0755); err != nil {
		return "", err
	}
	path := filepath.Join(mediaDir, filename)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", err
	}
	return "/media/" + filename, nil
}

func StitchFramesToVideo(frameDir string, outputFilename string) (string, error) {
	mediaDir := "/data/media"
	os.MkdirAll(mediaDir, 0755)
	outputPath := filepath.Join(mediaDir, outputFilename)

	// ffmpeg -framerate 10 -i frame_%03d.png -c:v libx264 -pix_fmt yuv420p output.mp4
	cmd := exec.Command("ffmpeg", "-y", "-framerate", "10", "-i", filepath.Join(frameDir, "frame_%03d.png"), "-c:v", "libx264", "-pix_fmt", "yuv420p", outputPath)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg failed: %v", err)
	}

	return "/media/" + outputFilename, nil
}

func DownloadToMedia(url string, filename string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return SaveMedia(data, filename)
}
