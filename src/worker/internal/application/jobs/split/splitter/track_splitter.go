package splitter

import (
	"context"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/tracks/entity"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/lib/cerr"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/lib/storagepath"
)

var splitDirNames = map[SplitType]string{
	SplitTwoStemsType:  "2stems",
	SplitFourStemsType: "4stems",
	SplitFiveStemsType: "5stems",
}

type TrackSplitter struct {
	trackStore    entity.TrackStore
	splitter      FileSplitter
	pathGenerator storagepath.Generator
}

func NewTrackSplitter(splitter FileSplitter, trackStore entity.TrackStore, pathGenerator storagepath.Generator) TrackSplitter {
	return TrackSplitter{
		trackStore:    trackStore,
		splitter:      splitter,
		pathGenerator: pathGenerator,
	}
}

func (t TrackSplitter) SplitTrack(ctx context.Context, tracklistID string, trackID string, savedOriginalURL string) (StemFilePaths, error) {
	errctx := cerr.Fields(cerr.F{
		"tracklist_id":       tracklistID,
		"track_id":           trackID,
		"saved_original_url": savedOriginalURL,
	})

	track, err := t.trackStore.GetTrack(ctx, tracklistID, trackID)
	if err != nil {
		return nil, errctx.Wrap(err).Error("Failed to get track from track store")
	}

	splitStemTrack, ok := track.(entity.SplitStemTrack)
	if !ok {
		return nil, errctx.Error("Unexpected: track is not a split request")
	}

	errctx = errctx.Field("track", splitStemTrack)

	splitType, err := ConvertToSplitType(splitStemTrack.TrackType)
	if err != nil {
		return nil, errctx.Wrap(err).Error("Failed to recognize track type as split type")
	}

	destPath, err := t.generatePath(tracklistID, trackID, splitType)
	if err != nil {
		return nil, errctx.Field("split_type", splitType).
			Wrap(err).Error("Failed to generate a destination path for stem tracks")
	}

	return t.splitter.SplitFile(ctx, savedOriginalURL, destPath, splitType)
}

func (t TrackSplitter) generatePath(tracklistID string, trackID string, splitType SplitType) (string, error) {
	splitDir, ok := splitDirNames[splitType]
	if !ok {
		return "", cerr.Error("Invalid split type provided")
	}

	return t.pathGenerator.GeneratePath(tracklistID, trackID, splitDir), nil
}
