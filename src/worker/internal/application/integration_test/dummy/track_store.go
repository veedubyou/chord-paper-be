package dummy

import (
	"context"
	trackentity "github.com/veedubyou/chord-paper-be/src/shared/track/entity"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/lib/cerr"
	"sync"
)

var _ trackentity.Store = &TrackStore{}

func NewDummyTrackStore() *TrackStore {
	return &TrackStore{
		Unavailable: false,
		State:       make(map[string]trackentity.TrackList),
	}
}

type TrackStore struct {
	Unavailable bool
	State       map[string]trackentity.TrackList
	mutex       sync.RWMutex
}

func (t *TrackStore) GetTrackList(ctx context.Context, songID string) (trackentity.TrackList, error) {
	if t.Unavailable {
		return trackentity.TrackList{}, NetworkFailure
	}

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	tracklist, ok := t.State[songID]
	if !ok {
		return trackentity.TrackList{}, NotFound
	}

	return tracklist, nil
}

func (t *TrackStore) SetTrackList(ctx context.Context, tracklist trackentity.TrackList) error {
	if t.Unavailable {
		return NetworkFailure
	}

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	t.State[tracklist.Defined.SongID] = tracklist
	return nil
}

func (t *TrackStore) UpdateTrack(ctx context.Context, trackListID string, trackID string, updater trackentity.TrackUpdater) error {
	if t.Unavailable {
		return NetworkFailure
	}

	tracklist, err := t.GetTrackList(ctx, trackListID)
	if err != nil {
		return cerr.Wrap(err).Error("Failed to get tracklist from DB")
	}

	for i, track := range tracklist.Defined.Tracks {
		if track.GetID() == trackID {
			updatedTrack, err := updater(track)
			if err != nil {
				return cerr.Wrap(err).Error("Track update function failed")
			}

			tracklist.Defined.Tracks[i] = updatedTrack
			return t.SetTrackList(ctx, tracklist)
		}
	}

	return cerr.Error("Track ID not found in tracklist")
}
