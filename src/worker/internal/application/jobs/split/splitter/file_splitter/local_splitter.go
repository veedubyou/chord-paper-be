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
	"path"
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

func NewLocalFileSplitter(spleeterWorkingDirStr string, spleeterBinPath string, demucsWorkingDirStr string, demucsBinPath string, executor executor.Executor) (LocalFileSplitter, error) {
	spleeterWorkingDir, err := working_dir.NewWorkingDir(spleeterWorkingDirStr)
	if err != nil {
		return LocalFileSplitter{}, cerr.Wrap(err).Error("Failed to convert working dir to absolute format")
	}

	demucsWorkingDir, err := working_dir.NewWorkingDir(demucsWorkingDirStr)
	if err != nil {
		return LocalFileSplitter{}, cerr.Wrap(err).Error("Failed to convert working dir to absolute format")
	}

	return LocalFileSplitter{
		spleeterWorkingDir: spleeterWorkingDir,
		spleeterBinPath:    spleeterBinPath,
		demucsWorkingDir:   demucsWorkingDir,
		demucsBinPath:      demucsBinPath,
		executor:           executor,
	}, nil
}

type LocalFileSplitter struct {
	spleeterWorkingDir working_dir.WorkingDir
	spleeterBinPath    string
	demucsWorkingDir   working_dir.WorkingDir
	demucsBinPath      string
	executor           executor.Executor
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
		filePaths, err := l.runSpleeter(absOriginalTrackFilePath, absStemsOutputDir, splitType)
		if err != nil {
			return nil, cerr.Field("output_dir", absStemsOutputDir).
				Wrap(err).Error("Failed to execute spleeter")
		}

		return filePaths, nil

	case trackentity.DemucsType:
		fallthrough
	case trackentity.DemucsV3Type:
		filePaths, err := l.runDemucs(absOriginalTrackFilePath, absStemsOutputDir, splitType, engineType)
		if err != nil {
			return nil, cerr.Field("output_dir", absStemsOutputDir).
				Wrap(err).Error("Failed to execute demucs")
		}

		return filePaths, nil

	default:
		return nil, cerr.Field("engine_type", engineType).Error("Unexpected engine type")
	}
}

func (l LocalFileSplitter) runDemucs(sourcePath string, destPath string, splitType splitter.SplitType, engineType trackentity.SplitEngineType) (splitter.StemFilePaths, error) {
	logger := log.WithFields(log.Fields{
		"sourcePath":       sourcePath,
		"destPath":         destPath,
		"splitType":        splitType,
		"demucsWorkingDir": l.demucsWorkingDir,
	})

	logger.Info("Running demucs command")

	var model string

	switch engineType {
	case trackentity.DemucsType:
		model = "htdemucs"
	case trackentity.DemucsV3Type:
		model = "hdemucs_mmi"
	default:
		return nil, cerr.Field("engine_type", engineType).Error("Unsupported engine type")
	}

	args := []string{"-o", destPath, "--name", model, "--device", "cpu", "--filename", "{stem}.{ext}", "--mp3"}

	switch splitType {
	case splitter.SplitTwoStemsType:
		args = append(args, "--two-stems", "vocals")
	case splitter.SplitFourStemsType:
		// do nothing, all good
	default:
		return nil, cerr.Field("split_type", splitType).Error("Unsupported split type")
	}

	args = append(args, sourcePath)

	errctx := cerr.Field("demucs_bin_path", l.demucsBinPath).Field("demucs_args", args)

	cmd := l.executor.Command(l.demucsBinPath, args...)
	cmd.SetDir(l.demucsWorkingDir.Root())

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errctx.Field("demucs_output", string(output)).
			Wrap(err).
			Error(fmt.Sprintf("Error occurred while running demucs: %s", string(output)))
	}

	logger.Debug(string(output))
	logger.Info("Finished demucs command")

	outputDir := path.Join(destPath, model)

	filePaths, err := collectStemFilePaths(outputDir)
	if err != nil {
		return nil, errctx.Field("output_dir", outputDir).Wrap(err).Error("Failed to collect stem file paths")
	}

	return filePaths, nil
}

func (l LocalFileSplitter) runSpleeter(sourcePath string, destPath string, splitType splitter.SplitType) (splitter.StemFilePaths, error) {
	logger := log.WithFields(log.Fields{
		"sourcePath":         sourcePath,
		"destPath":           destPath,
		"splitType":          splitType,
		"spleeterWorkingDir": l.spleeterWorkingDir,
	})

	splitParam, ok := spleeterParamMap[splitType]
	if !ok {
		return nil, cerr.Error("Invalid split type passed in!")
	}

	logger.Info("Running spleeter command")

	args := []string{"separate", "-p", splitParam, "-o", destPath, "-c", "mp3", "-b", "320k", "-f", "{instrument}.mp3", sourcePath}

	errctx := cerr.Field("spleeter_bin_path", l.spleeterBinPath).Field("spleeter_args", args)

	cmd := l.executor.Command(l.spleeterBinPath, args...)
	cmd.SetDir(l.spleeterWorkingDir.Root())

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errctx.Field("spleeter_output", string(output)).
			Wrap(err).
			Error(fmt.Sprintf("Error occurred while running spleeter: %s", string(output)))
	}

	logger.Debug(string(output))
	logger.Info("Finished spleeter command")

	filePaths, err := collectStemFilePaths(destPath)
	if err != nil {
		return nil, errctx.Field("output_dir", destPath).Wrap(err).Error("Failed to collect stem file paths")
	}

	return filePaths, nil
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
