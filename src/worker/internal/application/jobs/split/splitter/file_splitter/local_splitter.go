package file_splitter

import (
	"context"
	"fmt"
	trackentity "github.com/veedubyou/chord-paper-be/src/shared/track/entity"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/executor"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/split/splitter"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/lib/cerr"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/lib/working_dir"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
)

var _ splitter.FileSplitter = LocalFileSplitter{}

var spleeterParamMap = map[splitter.SplitType]string{
	splitter.SplitTwoStemsType:  "spleeter:2stems-16kHz",
	splitter.SplitFourStemsType: "spleeter:4stems-16kHz",
	splitter.SplitFiveStemsType: "spleeter:5stems-16kHz",
}

func NewLocalFileSplitter(workingDirStr string, spleeterBinPath string, demucsBinPath string, executor executor.Executor) (LocalFileSplitter, error) {
	workingDir, err := working_dir.NewWorkingDir(workingDirStr)
	if err != nil {
		return LocalFileSplitter{}, cerr.Wrap(err).Error("Failed to convert working dir to absolute format")
	}
	return LocalFileSplitter{
		workingDir:      workingDir,
		spleeterBinPath: spleeterBinPath,
		demucsBinPath:   demucsBinPath,
		executor:        executor,
	}, nil
}

type LocalFileSplitter struct {
	workingDir      working_dir.WorkingDir
	spleeterBinPath string
	demucsBinPath   string
	executor        executor.Executor
}

func (l LocalFileSplitter) SplitFile(ctx context.Context, originalTrackFilePath string, stemsOutputDir string, splitType splitter.SplitType, engineType trackentity.SplitEngineType) (splitter.StemFilePaths, error) {
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

	switch engineType {
	case trackentity.SpleeterType:
		if err := l.runSpleeter(absOriginalTrackFilePath, absStemsOutputDir, splitType); err != nil {
			return nil, cerr.Field("output_dir", absStemsOutputDir).
				Wrap(err).Error("Failed to execute spleeter")
		}

	case trackentity.DemucsType:
		if err := l.runDemucs(absOriginalTrackFilePath, absStemsOutputDir, splitType); err != nil {
			return nil, cerr.Field("output_dir", absStemsOutputDir).
				Wrap(err).Error("Failed to execute demucs")
		}
	}

	return collectStemFilePaths(absStemsOutputDir)
}

func (l LocalFileSplitter) runDemucs(sourcePath string, destPath string, splitType splitter.SplitType) error {
	logger := log.WithFields(log.Fields{
		"sourcePath": sourcePath,
		"destPath":   destPath,
		"splitType":  splitType,
		"workingDir": l.workingDir,
	})

	logger.Info("Running demucs command")

	args := []string{"-o", destPath, "--mp3", "-d", "cpu", "--filename", "{stem}.{ext}", sourcePath}

	errctx := cerr.Field("demucs_bin_path", l.demucsBinPath).Field("demucs_args", args)

	cmd := l.executor.Command(l.demucsBinPath, args...)
	cmd.SetDir(l.workingDir.Root())

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errctx.Field("demucs_output", string(output)).
			Wrap(err).
			Error(fmt.Sprintf("Error occurred while running demucs: %s", string(output)))
	}

	logger.Debug(string(output))
	logger.Info("Finished demucs command")

	return nil
}

func (l LocalFileSplitter) runSpleeter(sourcePath string, destPath string, splitType splitter.SplitType) error {
	logger := log.WithFields(log.Fields{
		"sourcePath": sourcePath,
		"destPath":   destPath,
		"splitType":  splitType,
		"workingDir": l.workingDir,
	})

	splitParam, ok := spleeterParamMap[splitType]
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

func collectStemFilePaths(dir string) (splitter.StemFilePaths, error) {
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

	outputs := splitter.StemFilePaths{}

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
