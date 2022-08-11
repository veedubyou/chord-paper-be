package transfer_test

import (
	"context"
	"fmt"
	"github.com/veedubyou/chord-paper-be/worker/src/internal/application/cloud_storage/store"
	dummy2 "github.com/veedubyou/chord-paper-be/worker/src/internal/application/integration_test/dummy"
	"github.com/veedubyou/chord-paper-be/worker/src/internal/application/jobs/job_message"
	transfer2 "github.com/veedubyou/chord-paper-be/worker/src/internal/application/jobs/transfer"
	download2 "github.com/veedubyou/chord-paper-be/worker/src/internal/application/jobs/transfer/download"
	"github.com/veedubyou/chord-paper-be/worker/src/internal/application/tracks/entity"

	"encoding/json"
	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("Download Job Handler", func() {
	var (
		youtubeDLBinPath string
		bucketName       string

		dummyTrackStore *dummy2.TrackStore
		dummyFileStore  *dummy2.FileStore
		dummyExecutor   *dummy2.YoutubeDLExecutor

		handler transfer2.JobHandler

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

			dummyTrackStore = dummy2.NewDummyTrackStore()
			dummyFileStore = dummy2.NewDummyFileStore()
			dummyExecutor = dummy2.NewDummyYoutubeDLExecutor()
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
			youtubeDownloader := download2.NewYoutubeDLer(youtubeDLBinPath, dummyExecutor)
			genericDownloader := download2.NewGenericDLer()
			selectDownloader := download2.NewSelectDLer(youtubeDownloader, genericDownloader)

			trackDownloader, err := transfer2.NewTrackTransferrer(selectDownloader, dummyTrackStore, dummyFileStore, bucketName, workingDir)
			Expect(err).NotTo(HaveOccurred())

			handler = transfer2.NewJobHandler(trackDownloader)
		})
	})

	Describe("Well formed message", func() {
		var job transfer2.JobParams
		BeforeEach(func() {
			job = transfer2.JobParams{
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
			var jobParams transfer2.JobParams
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
			job := transfer2.JobParams{
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
