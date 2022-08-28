package start_test

import (
	"context"
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	trackentity "github.com/veedubyou/chord-paper-be/src/shared/track/entity"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/integration_test/dummy"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/job_message"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/start"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/transfer"
)

var _ = Describe("Start", func() {
	var (
		dummyTrackStore *dummy.TrackStore

		handler start.JobHandler

		message []byte

		tracklistID string
		trackID     string
	)

	BeforeEach(func() {
		By("Initializing all variables", func() {
			message = nil

			tracklistID = "tracklist-id"
			trackID = "track-id"

			dummyTrackStore = dummy.NewDummyTrackStore()
		})

		By("Setting up the dummy track store data", func() {
			tracklist := trackentity.TrackList{}
			tracklist.Defined.SongID = tracklistID
			tracklist.Defined.Tracks = trackentity.Tracks{
				&trackentity.SplitRequestTrack{
					TrackFields: trackentity.TrackFields{ID: trackID},
					TrackType:   trackentity.SplitFourStemsType,
					OriginalURL: "",
					Status:      trackentity.RequestedStatus,
				},
			}

			err := dummyTrackStore.SetTrackList(context.Background(), tracklist)
			Expect(err).NotTo(HaveOccurred())
		})

		By("Instantiating the handler", func() {
			handler = start.NewJobHandler(dummyTrackStore)
		})
	})

	Describe("Well formed message", func() {
		var job start.JobParams
		BeforeEach(func() {
			job = start.JobParams{
				TrackIdentifier: job_message.TrackIdentifier{
					TrackListID: tracklistID,
					TrackID:     trackID,
				},
			}

			var err error
			message, err = json.Marshal(job)
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("Happy path", func() {
			var err error
			var jobParams start.JobParams

			BeforeEach(func() {
				jobParams, err = handler.HandleStartJob(message)
			})

			It("doesn't return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("updates the track status", func() {
				tracklist, err := dummyTrackStore.GetTrackList(context.Background(), tracklistID)
				Expect(err).NotTo(HaveOccurred())

				track, err := tracklist.GetTrack(trackID)
				Expect(err).NotTo(HaveOccurred())

				splitStemTrack, ok := track.(*trackentity.SplitRequestTrack)
				Expect(ok).To(BeTrue())

				Expect(splitStemTrack.Status).To(Equal(trackentity.ProcessingStatus))
			})

			It("returns the processed data", func() {
				Expect(jobParams.TrackListID).To(Equal(job.TrackListID))
				Expect(jobParams.TrackID).To(Equal(job.TrackID))
			})
		})

		Describe("Can't reach track store", func() {
			BeforeEach(func() {
				dummyTrackStore.Unavailable = true
			})

			It("returns an error", func() {
				_, err := handler.HandleStartJob(message)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Poorly formed message", func() {
		BeforeEach(func() {
			job := transfer.JobParams{
				TrackIdentifier: job_message.TrackIdentifier{
					TrackListID: tracklistID,
				},
			}

			var err error
			message, err = json.Marshal(job)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns error", func() {
			_, err := handler.HandleStartJob(message)
			Expect(err).To(HaveOccurred())
		})
	})
})
