package trackstorage_test

import (
	"context"
	"fmt"
	"github.com/cockroachdb/errors/markers"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/veedubyou/chord-paper-be/src/server/internal/track/entity"
	"github.com/veedubyou/chord-paper-be/src/server/internal/track/storage"
	. "github.com/veedubyou/chord-paper-be/src/shared/testing"
)

var _ = Describe("Track DB", func() {
	var (
		trackDB trackstorage.DB
	)

	BeforeEach(func() {
		ResetDB(db)
		trackDB = trackstorage.NewDB(db)
	})

	Describe("UpdateTrack", func() {
		var (
			newTrack trackentity.Track
		)

		BeforeEach(func() {
			newTrack = trackentity.Track{}
			newTrack.Defined.TrackType = "unrecognizable"
			newTrack.Extra = map[string]interface{}{
				"super-secret-data": "burp",
			}
		})

		Describe("With an existing tracklist", func() {
			var (
				tracklist   trackentity.TrackList
				setTrackErr error
			)

			var (
				trackSize = 3
			)

			BeforeEach(func() {
				setTrackErr = nil

				tracklist = trackentity.TrackList{}
				tracklist.Defined.SongID = uuid.New().String()
				tracklist.Extra = map[string]interface{}{
					"extra": "metadata",
				}

				track1 := trackentity.Track{}
				track1.CreateID()
				track1.Defined.TrackType = "4stems"
				track1.Extra = map[string]interface{}{}

				track2 := trackentity.Track{}
				track2.CreateID()
				track2.Defined.TrackType = "accompaniment"
				track2.Extra = map[string]interface{}{
					"accompaniment_url": "accompaniment.mp3",
				}

				track3 := trackentity.Track{}
				track3.CreateID()
				track3.Defined.TrackType = "split_2stems"
				track3.Extra = map[string]interface{}{
					"url": "song.mp3",
				}
				track3.InitializeSplitJob()

				tracklist.Defined.Tracks = []trackentity.Track{
					track1, track2, track3,
				}

				Expect(trackSize).To(Equal(3))

				err := trackDB.SetTrackList(context.Background(), tracklist)
				Expect(err).NotTo(HaveOccurred())
			})

			JustBeforeEach(func() {
				setTrackErr = trackDB.UpdateTrack(context.Background(), tracklist.Defined.SongID, newTrack)
			})

			AfterEach(func() {
				setTrackErr = nil
			})

			for trackIndex := 0; trackIndex < trackSize; trackIndex++ {
				trackIndex := trackIndex
				Describe(fmt.Sprintf("Setting the track in the index %d", trackIndex), func() {
					BeforeEach(func() {
						newTrack.Defined.ID = tracklist.Defined.Tracks[trackIndex].Defined.ID
					})

					It("succeeds", func() {
						Expect(setTrackErr).NotTo(HaveOccurred())
					})

					It("sets the track", func() {
						fetchedTracklist := ExpectSuccess(trackDB.GetTrackList(context.Background(), tracklist.Defined.SongID))
						expectedTrackList := tracklist
						expectedTrackList.Defined.Tracks[trackIndex] = newTrack

						Expect(fetchedTracklist).To(Equal(expectedTrackList))
					})
				})
			}

			Describe("Setting the track without an existing matching track", func() {
				BeforeEach(func() {
					newTrack.CreateID()
				})

				It("fails", func() {
					Expect(setTrackErr).To(HaveOccurred())
					Expect(markers.Is(setTrackErr, trackstorage.TrackNotFound)).To(BeTrue())
				})
			})
		})

		Describe("Without an existing tracklist", func() {
			BeforeEach(func() {
				newTrack.CreateID()
			})

			It("fails", func() {
				err := trackDB.UpdateTrack(context.Background(), uuid.New().String(), newTrack)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
