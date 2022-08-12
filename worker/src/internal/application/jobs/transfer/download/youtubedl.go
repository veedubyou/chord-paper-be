package download

import (
	"fmt"
	"github.com/veedubyou/chord-paper-be/worker/src/internal/application/executor"
	"github.com/veedubyou/chord-paper-be/worker/src/internal/lib/cerr"

	"github.com/apex/log"
)

var _ Downloader = YoutubeDLer{}

func NewYoutubeDLer(youtubedlBinPath string, commandExecutor executor.Executor) YoutubeDLer {
	return YoutubeDLer{
		youtubedlBinPath: youtubedlBinPath,
		commandExecutor:  commandExecutor,
	}
}

type YoutubeDLer struct {
	youtubedlBinPath string
	commandExecutor  executor.Executor
}

func (y YoutubeDLer) Download(sourceURL string, outFilePath string) error {
	err := y.download(sourceURL, outFilePath)
	// error may be fixable by clearing the cache dir
	// so try again in case that's the issue
	if err != nil {
		y.clearCache()
		return y.download(sourceURL, outFilePath)
	}

	return nil
}

func (y YoutubeDLer) download(sourceURL string, outFilePath string) error {
	log.Info("Running youtube-dl")

	cmd := y.commandExecutor.Command(y.youtubedlBinPath, "-o", outFilePath, "-x", "--audio-format", "mp3", "--audio-quality", "0", sourceURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return cerr.Field("error_msg", string(output)).
			Wrap(err).
			Error(fmt.Sprintf("Failed to run youtube-dl: %s", string(output)))
	}

	return nil
}

func (y YoutubeDLer) clearCache() {
	log.Info("Clearing youtube-dl cache")
	cmd := y.commandExecutor.Command(y.youtubedlBinPath, "--rm-cache-dir")
	output, err := cmd.CombinedOutput()
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to clear cache: %s", string(output))
		log.Error(errorMsg)
	}
}
