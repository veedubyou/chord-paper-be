package config

import (
	"fmt"
	"os/exec"
	"strings"
)

func FindBin(bin string) string {
	cmd := exec.Command("which", bin)
	output, err := cmd.CombinedOutput()

	stringOutput := string(output)
	if err != nil {
		panic(fmt.Sprintf("Failed to find %s: %s", bin, stringOutput))
	}

	trimmedOutput := strings.TrimSpace(stringOutput)
	if trimmedOutput == "" {
		panic(fmt.Sprintf("No bin found for %s", bin))
	}

	return trimmedOutput
}

func SpleeterPath() string {
	return FindBin("spleeter")
}

func DemucsPath() string {
	return FindBin("demucs")
}

func YoutubeDLPath() string {
	return FindBin("yt-dlp")
}
