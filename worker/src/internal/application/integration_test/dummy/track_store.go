package dummy

import (
	"context"
	entity2 "github.com/veedubyou/chord-paper-be/worker/src/internal/application/tracks/entity"
	"github.com/veedubyou/chord-paper-be/worker/src/internal/lib/cerr"
	"sync"
)

var _ entity2.TrackStore = &TrackStore{}

func NewDummyTrackStore() *TrackStore {
	return &TrackStore{
		Unavailable: false,
		State:       make(map[string]map[string]entity2.Track),
	}
}

type TrackStore struct {
	Unavailable bool
	State       map[string]map[string]entity2.Track
	mutex       sync.RWMutex
}

func (t *TrackStore) GetTrack(_ context.Context, tracklistID string, trackID string) (entity2.Track, error) {
	if t.Unavailable {
		return entity2.BaseTrack{}, NetworkFailure
	}

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	trackMap, ok := t.State[tracklistID]
	if !ok {
		return entity2.BaseTrack{}, NotFound
	}

	track, ok := trackMap[trackID]
	if !ok {
		return entity2.BaseTrack{}, NotFound
	}

	return track, nil
}

func (t *TrackStore) SetTrack(_ context.Context, tracklistID string, trackID string, track entity2.Track) error {
	if t.Unavailable {
		return NetworkFailure
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.State[tracklistID] == nil {
		t.State[tracklistID] = make(map[string]entity2.Track)
	}

	t.State[tracklistID][trackID] = track

	return nil
}

func (t *TrackStore) UpdateTrack(ctx context.Context, trackListID string, trackID string, updater entity2.TrackUpdater) error {
	if t.Unavailable {
		return NetworkFailure
	}

	track, err := t.GetTrack(ctx, trackListID, trackID)
	if err != nil {
		return cerr.Wrap(err).Error("Failed to get track from DB")
	}

	updatedTrack, err := updater(track)
	if err != nil {
		return cerr.Wrap(err).Error("Track update function failed")
	}

	err = t.SetTrack(ctx, trackListID, trackID, updatedTrack)
	if err != nil {
		return cerr.Wrap(err).Error("Failed to set the updated track")
	}

	return nil
}
