package file_splitter

import (
	"context"
	"fmt"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/executor"
	splitter2 "github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/split/splitter"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/lib/cerr"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/lib/working_dir"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
)

var _ splitter2.FileSplitter = LocalFileSplitter{}

var paramMap = map[splitter2.SplitType]string{
	splitter2.SplitTwoStemsType:  "spleeter:2stems-16kHz",
	splitter2.SplitFourStemsType: "spleeter:4stems-16kHz",
	splitter2.SplitFiveStemsType: "spleeter:5stems-16kHz",
}

func NewLocalFileSplitter(workingDirStr string, spleeterBinPath string, executor executor.Executor) (LocalFileSplitter, error) {
	workingDir, err := working_dir.NewWorkingDir(workingDirStr)
	if err != nil {
		return LocalFileSplitter{}, cerr.Wrap(err).Error("Failed to convert working dir to absolute format")
	}
	return LocalFileSplitter{
		workingDir:      workingDir,
		spleeterBinPath: spleeterBinPath,
		executor:        executor,
	}, nil
}

type LocalFileSplitter struct {
	workingDir      working_dir.WorkingDir
	spleeterBinPath string
	executor        executor.Executor
}

func (l LocalFileSplitter) SplitFile(ctx context.Context, originalTrackFilePath string, stemsOutputDir string, splitType splitter2.SplitType) (splitter2.StemFilePaths, error) {
	absOriginalTrackFilePath, err := filepath.Abs(originalTrackFilePath)
	if err != nil {
		return nil, cerr.Wrap(err).Error("Cannot convert source path to absolute format")
	}

	errctx := cerr.Field("original_filepath", absOriginalTrackFilePath)

	absStemsOutputDir, err := filepath.Abs(stemsOutputDir)
	if err != nil {
		return nil, errctx.Wrap(err).Error("Cannot convert destination path to absolute format")
	}

	// splitting is a lengthy process, if we want to halt now is the time
	if ctx.Err() != nil {
		return nil, cerr.Wrap(ctx.Err()).Error("Context cancelled before splitting could happen")
	}

	if err := l.runSpleeter(absOriginalTrackFilePath, absStemsOutputDir, splitType); err != nil {
		return nil, cerr.Field("output_dir", absStemsOutputDir).
			Wrap(err).Error("Failed to execute spleeter")
	}

	return collectStemFilePaths(absStemsOutputDir)
}

func (l LocalFileSplitter) runSpleeter(sourcePath string, destPath string, splitType splitter2.SplitType) error {
	logger := log.WithFields(log.Fields{
		"sourcePath": sourcePath,
		"destPath":   destPath,
		"splitType":  splitType,
		"workingDir": l.workingDir,
	})

	splitParam, ok := paramMap[splitType]
	if !ok {
		return cerr.Error("Invalid split type passed in!")
	}

	logger.Info("Running spleeter command")

	args := []string{"separate", "-p", splitParam, "-o", destPath, "-c", "mp3", "-b", "320k", "-f", "{instrument}.mp3", sourcePath}

	errctx := cerr.Field("spleeter_bin_path", l.spleeterBinPath).Field("spleeter_args", args)

	cmd := l.executor.Command(l.spleeterBinPath, args...)
	cmd.SetDir(l.workingDir.Root())

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errctx.Field("spleeter_output", string(output)).
			Wrap(err).
			Error(fmt.Sprintf("Error occurred while running spleeter: %s", string(output)))
	}

	logger.Debug(string(output))
	logger.Info("Finished spleeter command")

	return nil
}

func collectStemFilePaths(dir string) (splitter2.StemFilePaths, error) {
	logger := log.WithFields(log.Fields{
		"dir": dir,
	})

	logger.Info("Reading directory to collect file paths")
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, cerr.Wrap(err).Error("Error reading output directory")
	}

	if len(dirEntries) == 0 {
		return nil, cerr.Error("No files in output directory")
	}

	outputs := splitter2.StemFilePaths{}

	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			continue
		}

		fileName := dirEntry.Name()
		relFilePath := filepath.Join(dir, fileName)
		filePath, err := filepath.Abs(relFilePath)
		if err != nil {
			return nil, cerr.Field("relative_file_path", relFilePath).
				Wrap(err).Error("Failed to convert file path to absolute format")
		}

		stemName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		outputs[stemName] = filePath
	}

	return outputs, nil
}
