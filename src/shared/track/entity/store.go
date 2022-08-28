package trackentity

import (
	"context"
)

type TrackUpdater func(track Track) (Track, error)

type Store interface {
	GetTrackList(ctx context.Context, songID string) (TrackList, error)
	SetTrackList(ctx context.Context, tracklist TrackList) error
	UpdateTrack(ctx context.Context, tracklistID string, trackID string, updater TrackUpdater) error
}
