package transfer_test

import (
	"context"
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/cloud_storage/store"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/integration_test/dummy"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/job_message"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/transfer"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/transfer/download"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/tracks/entity"
)

var _ = Describe("Download Job Handler", func() {
	var (
		youtubeDLBinPath string
		bucketName       string

		dummyTrackStore *dummy.TrackStore
		dummyFileStore  *dummy.FileStore
		dummyExecutor   *dummy.YoutubeDLExecutor

		handler transfer.JobHandler

		message           []byte
		originalURL       string
		originalTrackData []byte

		tracklistID string
		trackID     string
	)

	BeforeEach(func() {
		By("Initializing all variables", func() {
			message = nil
			youtubeDLBinPath = "/bin/youtube-dl"
			bucketName = "bucket-head"

			tracklistID = "tracklist-id"
			trackID = "track-id"
			originalURL = "https://youtube.com/coolsong.mp3"
			originalTrackData = []byte("cool_jamz")

			dummyTrackStore = dummy.NewDummyTrackStore()
			dummyFileStore = dummy.NewDummyFileStore()
			dummyExecutor = dummy.NewDummyYoutubeDLExecutor()
		})

		By("Setting up the dummy track store data", func() {
			err := dummyTrackStore.SetTrack(context.Background(), tracklistID, trackID, entity.SplitStemTrack{
				BaseTrack: entity.BaseTrack{
					TrackType: entity.SplitFourStemsType,
				},
				OriginalURL: originalURL,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		By("Setting up the dummy executor", func() {
			dummyExecutor.AddURL(originalURL, originalTrackData)
		})

		By("Instantiating the handler", func() {
			youtubeDownloader := download.NewYoutubeDLer(youtubeDLBinPath, dummyExecutor)
			genericDownloader := download.NewGenericDLer()
			selectDownloader := download.NewSelectDLer(youtubeDownloader, genericDownloader)

			trackDownloader, err := transfer.NewTrackTransferrer(selectDownloader, dummyTrackStore, dummyFileStore, bucketName, workingDir)
			Expect(err).NotTo(HaveOccurred())

			handler = transfer.NewJobHandler(trackDownloader)
		})
	})

	Describe("Well formed message", func() {
		var job transfer.JobParams
		BeforeEach(func() {
			job = transfer.JobParams{
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
			var expectedSavedURL string
			var jobParams transfer.JobParams
			var savedOriginalURL string

			BeforeEach(func() {
				jobParams, savedOriginalURL, err = handler.HandleTransferJob(message)
				expectedSavedURL = fmt.Sprintf("%s/%s/%s/%s/original/original.mp3", store.GOOGLE_STORAGE_HOST, bucketName, tracklistID, trackID)
			})

			It("doesn't return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("saved the track to the file store", func() {
				contents, err := dummyFileStore.GetFile(context.Background(), expectedSavedURL)
				Expect(err).NotTo(HaveOccurred())
				Expect(contents).To(Equal(originalTrackData))
			})

			It("returns the processed data", func() {
				Expect(savedOriginalURL).To(Equal(expectedSavedURL))
				Expect(jobParams.TrackListID).To(Equal(job.TrackListID))
				Expect(jobParams.TrackID).To(Equal(job.TrackID))
			})
		})

		Describe("Can't reach track store", func() {
			BeforeEach(func() {
				dummyTrackStore.Unavailable = true
			})

			It("returns an error", func() {
				_, _, err := handler.HandleTransferJob(message)
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
			_, _, err := handler.HandleTransferJob(message)
			Expect(err).To(HaveOccurred())
		})
	})
})
