package transfer

import (
	trackentity "github.com/veedubyou/chord-paper-be/src/shared/track/entity"
	cloudstorage "github.com/veedubyou/chord-paper-be/src/worker/internal/application/cloud_storage/entity"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/transfer/download"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/lib/cerr"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/lib/storagepath"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/lib/working_dir"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/apex/log"

	"context"
)

func NewTrackTransferrer(downloader download.SelectDLer, trackStore trackentity.Store, fileStore cloudstorage.FileStore, pathGenerator storagepath.Generator, workingDirStr string) (TrackTransferrer, error) {
	workingDir, err := working_dir.NewWorkingDir(workingDirStr)
	if err != nil {
		return TrackTransferrer{}, cerr.Field("working_dir_str", workingDirStr).Wrap(err).Error("Failed to create working dir")
	}

	return TrackTransferrer{
		fileStore:     fileStore,
		trackStore:    trackStore,
		downloader:    downloader,
		pathGenerator: pathGenerator,
		workingDir:    workingDir,
	}, nil
}

type TrackTransferrer struct {
	fileStore     cloudstorage.FileStore
	trackStore    trackentity.Store
	downloader    download.SelectDLer
	pathGenerator storagepath.Generator
	workingDir    working_dir.WorkingDir
}

func (t TrackTransferrer) Download(tracklistID string, trackID string) (string, error) {
	errctx := cerr.Field("tracklist_id", tracklistID).Field("track_id", trackID)

	tracklist, err := t.trackStore.GetTrackList(context.Background(), tracklistID)
	if err != nil {
		return "", errctx.Wrap(err).Error("Failed to GetTrackList")
	}

	track, err := tracklist.GetTrack(trackID)
	if err != nil {
		return "", errctx.Wrap(err).Error("Failed to GetTrack")
	}

	splitStemTrack, ok := track.(*trackentity.SplitRequestTrack)
	if !ok {
		return "", errctx.Wrap(err).Error("Unexpected - track is not a split request")
	}

	tempFilePath, cleanUpTempDir, err := t.makeTempOutFilePath()
	if err != nil {
		return "", errctx.Wrap(err).Error("Failed to make a temp file path")
	}

	defer cleanUpTempDir()

	err = t.downloader.Download(splitStemTrack.OriginalURL, tempFilePath)
	if err != nil {
		return "", errctx.Field("original_url", splitStemTrack.OriginalURL).
			Wrap(err).Error("Failed to download track to cloud")
	}

	log.Info("Reading output file to memory")
	fileContent, err := os.ReadFile(tempFilePath)
	if err != nil {
		return "", errctx.Wrap(err).Error("Failed to read outputed youtubedl mp3")
	}

	destinationURL := t.pathGenerator.GeneratePath(tracklistID, trackID, "original/original.mp3")

	log.Info("Writing file to remote file store")
	err = t.fileStore.WriteFile(context.Background(), destinationURL, fileContent)
	if err != nil {
		return "", errctx.Wrap(err).Error("Failed to write file to the cloud")
	}

	return destinationURL, nil
}

func (t TrackTransferrer) makeTempOutFilePath() (string, func(), error) {
	log.Info("Creating temp dir to store downloaded source file temporarily")
	tempDir, err := ioutil.TempDir(t.workingDir.TempDir(), "transfer-*")
	if err != nil {
		return "", nil, cerr.Field("temp_dir", t.workingDir.TempDir()).
			Wrap(err).Error("Failed to create temp dir to download to")
	}

	tempDir, err = filepath.Abs(tempDir)
	if err != nil {
		return "", nil, cerr.Field("temp_dir", tempDir).
			Wrap(err).Error("Failed to turn temp dir into absolute format")
	}

	outputPath := filepath.Join(tempDir, "original.mp3")

	return outputPath, func() { os.RemoveAll(tempDir) }, nil
}
