package save_stems_to_db_test

import (
	"context"
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	trackentity "github.com/veedubyou/chord-paper-be/src/shared/track/entity"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/integration_test/dummy"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/job_message"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/save_stems_to_db"
)

var _ = Describe("JobHandler", func() {
	var (
		tracklistID string
		trackID     string
		trackType   trackentity.SplitRequestType

		dummyTrackStore *dummy.TrackStore
		handler         save_stems_to_db.JobHandler
	)

	BeforeEach(func() {
		tracklistID = "tracklist-ID"
		trackID = "track-ID"
		trackType = ""

		dummyTrackStore = dummy.NewDummyTrackStore()
		handler = save_stems_to_db.NewJobHandler(dummyTrackStore)
	})

	JustBeforeEach(func() {
		Expect(trackType).NotTo(BeZero())

		prevUnavailable := dummyTrackStore.Unavailable
		dummyTrackStore.Unavailable = false

		tracklist := trackentity.TrackList{}
		tracklist.Defined.SongID = tracklistID
		tracklist.Defined.Tracks = trackentity.Tracks{
			&trackentity.SplitRequestTrack{
				TrackFields: trackentity.TrackFields{ID: trackID},
				TrackType:   trackType,
				OriginalURL: "https://whocares",
			},
		}

		err := dummyTrackStore.SetTrackList(context.Background(), tracklist)

		dummyTrackStore.Unavailable = prevUnavailable
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Handle message", func() {
		var messageBytes []byte

		BeforeEach(func() {
			messageBytes = nil
		})

		Describe("Well formed job message", func() {
			var (
				stemURLs  map[string]string
				jobParams save_stems_to_db.JobParams
			)

			Describe("2stems", func() {
				BeforeEach(func() {
					stemURLs = map[string]string{
						"vocals":        "vocals.mp3",
						"accompaniment": "accompaniment.mp3",
					}
					jobParams = save_stems_to_db.JobParams{
						TrackIdentifier: job_message.TrackIdentifier{
							TrackListID: tracklistID,
							TrackID:     trackID,
						},
						StemURLS: stemURLs,
					}
					trackType = trackentity.SplitTwoStemsType

					var err error
					messageBytes, err = json.Marshal(jobParams)
					Expect(err).NotTo(HaveOccurred())
				})

				Describe("Store saves successfully", func() {
					It("does not error", func() {
						err := handler.HandleSaveStemsToDBJob(messageBytes)
						Expect(err).NotTo(HaveOccurred())
					})

					It("updates the track store", func() {
						_ = handler.HandleSaveStemsToDBJob(messageBytes)

						tracklist, err := dummyTrackStore.GetTrackList(context.Background(), tracklistID)
						Expect(err).NotTo(HaveOccurred())

						track, err := tracklist.GetTrack(trackID)
						Expect(err).NotTo(HaveOccurred())

						stemTrack, ok := track.(*trackentity.StemTrack)
						Expect(ok).To(BeTrue())
						Expect(stemTrack.TrackType).To(Equal(trackentity.TwoStemsType))
						Expect(stemTrack.StemURLs).To(Equal(stemURLs))
					})
				})

				Describe("Store is unavailable", func() {
					BeforeEach(func() {
						dummyTrackStore.Unavailable = true
					})

					It("also returns a failure", func() {
						err := handler.HandleSaveStemsToDBJob(messageBytes)
						Expect(err).To(HaveOccurred())
					})
				})
			})

			Describe("4stems", func() {
				BeforeEach(func() {
					stemURLs = map[string]string{
						"vocals": "vocals.mp3",
						"other":  "other.mp3",
						"bass":   "bass.mp3",
						"drums":  "drums.mp3",
					}
					jobParams = save_stems_to_db.JobParams{
						TrackIdentifier: job_message.TrackIdentifier{
							TrackListID: tracklistID,
							TrackID:     trackID,
						},
						StemURLS: stemURLs,
					}
					trackType = trackentity.SplitFourStemsType

					var err error
					messageBytes, err = json.Marshal(jobParams)
					Expect(err).NotTo(HaveOccurred())
				})

				Describe("Store saves successfully", func() {
					It("does not error", func() {
						err := handler.HandleSaveStemsToDBJob(messageBytes)
						Expect(err).NotTo(HaveOccurred())
					})

					It("updates the track store", func() {
						_ = handler.HandleSaveStemsToDBJob(messageBytes)
						tracklist, err := dummyTrackStore.GetTrackList(context.Background(), tracklistID)
						Expect(err).NotTo(HaveOccurred())

						track, err := tracklist.GetTrack(trackID)
						Expect(err).NotTo(HaveOccurred())

						stemTrack, ok := track.(*trackentity.StemTrack)
						Expect(ok).To(BeTrue())
						Expect(stemTrack.TrackType).To(Equal(trackentity.FourStemsType))
						Expect(stemTrack.StemURLs).To(Equal(stemURLs))
					})
				})

				Describe("Store is unavailable", func() {
					BeforeEach(func() {
						dummyTrackStore.Unavailable = true
					})

					It("also returns a failure", func() {
						err := handler.HandleSaveStemsToDBJob(messageBytes)
						Expect(err).To(HaveOccurred())
					})
				})
			})

			Describe("5stems", func() {
				BeforeEach(func() {
					stemURLs = map[string]string{
						"vocals": "vocals.mp3",
						"other":  "other.mp3",
						"bass":   "bass.mp3",
						"drums":  "drums.mp3",
						"piano":  "piano.mp3",
					}
					jobParams = save_stems_to_db.JobParams{
						TrackIdentifier: job_message.TrackIdentifier{
							TrackListID: tracklistID,
							TrackID:     trackID,
						},
						StemURLS: stemURLs,
					}
					trackType = trackentity.SplitFiveStemsType

					var err error
					messageBytes, err = json.Marshal(jobParams)
					Expect(err).NotTo(HaveOccurred())
				})

				Describe("Store saves successfully", func() {
					It("does not error", func() {
						err := handler.HandleSaveStemsToDBJob(messageBytes)
						Expect(err).NotTo(HaveOccurred())
					})

					It("updates the track store", func() {
						_ = handler.HandleSaveStemsToDBJob(messageBytes)
						tracklist, err := dummyTrackStore.GetTrackList(context.Background(), tracklistID)
						Expect(err).NotTo(HaveOccurred())

						track, err := tracklist.GetTrack(trackID)
						Expect(err).NotTo(HaveOccurred())

						stemTrack, ok := track.(*trackentity.StemTrack)
						Expect(ok).To(BeTrue())
						Expect(stemTrack.TrackType).To(Equal(trackentity.FiveStemsType))
						Expect(stemTrack.StemURLs).To(Equal(stemURLs))
					})
				})

				Describe("Store is unavailable", func() {
					BeforeEach(func() {
						dummyTrackStore.Unavailable = true
					})

					It("also returns a failure", func() {
						err := handler.HandleSaveStemsToDBJob(messageBytes)
						Expect(err).To(HaveOccurred())
					})
				})
			})

		})

		Describe("Malformed job message", func() {
			BeforeEach(func() {
				jobParams := save_stems_to_db.JobParams{
					TrackIdentifier: job_message.TrackIdentifier{
						TrackListID: tracklistID,
						TrackID:     trackID,
					},
				}

				var err error
				messageBytes, err = json.Marshal(jobParams)
				Expect(err).NotTo(HaveOccurred())

				trackType = trackentity.SplitRequestType(trackentity.TwoStemsType)
			})

			It("returns error", func() {
				err := handler.HandleSaveStemsToDBJob(messageBytes)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
