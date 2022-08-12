package integration_test_test

import (
	"bytes"
	"context"
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rabbitmq/amqp091-go"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/integration_test/dummy"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/job_message"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/job_router"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/save_stems_to_db"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/split"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/split/splitter"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/split/splitter/file_splitter"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/start"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/transfer"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/transfer/download"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/tracks/entity"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/worker"
)

var _ = Describe("IntegrationTest", func() {
	var (
		tracklistID       string
		trackID           string
		originalURL       string
		originalTrackData []byte
		bucketName        string

		rabbitMQ          *dummy.RabbitMQ
		fileStore         *dummy.FileStore
		trackStore        *dummy.TrackStore
		youtubeDLExecutor *dummy.YoutubeDLExecutor
		spleeterExecutor  *dummy.SpleeterExecutor

		queueWorker worker.QueueWorker
		run         func()
	)

	BeforeEach(func() {
		By("Assigning data to variables", func() {
			tracklistID = "track-list-ID"
			trackID = "track-ID"
			originalURL = "https://www.youtube.com/jams.mp3"
			originalTrackData = []byte("cool-jamz")
			bucketName = "bucket-head"
		})

		By("Instantiating all dummies", func() {
			rabbitMQ = dummy.NewRabbitMQ()
			fileStore = dummy.NewDummyFileStore()
			trackStore = dummy.NewDummyTrackStore()
			youtubeDLExecutor = dummy.NewDummyYoutubeDLExecutor()
			spleeterExecutor = dummy.NewDummySpleeterExecutor()
		})

		By("Setting up the track store", func() {
			track := entity.SplitStemTrack{
				BaseTrack: entity.BaseTrack{
					TrackType: entity.SplitFourStemsType,
				},
				OriginalURL: originalURL,
				JobStatus:   entity.RequestedStatus,
			}
			err := trackStore.SetTrack(context.Background(), tracklistID, trackID, track)
			Expect(err).NotTo(HaveOccurred())
		})

		By("Setting up the youtubeDL executor", func() {
			youtubeDLExecutor.AddURL(originalURL, originalTrackData)
		})

		var startHandler start.JobHandler
		By("Creating the start job handler", func() {
			startHandler = start.NewJobHandler(trackStore)
		})

		var transferHandler transfer.JobHandler
		By("Creating the download job handler", func() {
			youtubedler := download.NewYoutubeDLer("/whatever/youtube-dl", youtubeDLExecutor)
			genericdler := download.NewGenericDLer()
			selectdler := download.NewSelectDLer(youtubedler, genericdler)

			trackDownloader, err := transfer.NewTrackTransferrer(selectdler, trackStore, fileStore, bucketName, workingDir)
			Expect(err).NotTo(HaveOccurred())

			transferHandler = transfer.NewJobHandler(trackDownloader)
		})

		var splitHandler split.JobHandler
		By("Creating the split job handler", func() {
			localFileSplitter, err := file_splitter.NewLocalFileSplitter(workingDir, "/whatever/spleeter", spleeterExecutor)
			Expect(err).NotTo(HaveOccurred())
			remoteFileSplitter, err := file_splitter.NewRemoteFileSplitter(workingDir, fileStore, localFileSplitter)
			Expect(err).NotTo(HaveOccurred())
			trackSplitter := splitter.NewTrackSplitter(remoteFileSplitter, trackStore, bucketName)
			splitHandler = split.NewJobHandler(trackSplitter)
		})

		var saveHandler save_stems_to_db.JobHandler
		By("Creating the save stems to DB job handler", func() {
			saveHandler = save_stems_to_db.NewJobHandler(trackStore)
		})

		By("Instantiating the worker", func() {
			router := job_router.NewJobRouter(
				trackStore,
				rabbitMQ,
				startHandler,
				transferHandler,
				splitHandler,
				saveHandler,
			)
			queueWorker = worker.NewQueueWorker(rabbitMQ, "test-queue", router)
		})

		By("Setting up the run routine", func() {
			run = func() {
				go func() {
					err := queueWorker.Start()
					Expect(err).NotTo(HaveOccurred())
				}()

				startJobParams := start.JobParams{
					TrackIdentifier: job_message.TrackIdentifier{
						TrackListID: tracklistID,
						TrackID:     trackID,
					},
				}

				jsonBytes, err := json.Marshal(startJobParams)
				Expect(err).NotTo(HaveOccurred())

				message := amqp091.Publishing{
					Type: start.JobType,
					Body: jsonBytes,
				}
				err = rabbitMQ.Publish(message)
				Expect(err).NotTo(HaveOccurred())
			}
		})
	})

	Describe("All jobs run successfully", func() {
		It("gets 4 acks", func() {
			run()

			Eventually(func() int {
				return rabbitMQ.AckCounter
			}).Should(Equal(4))
		})

		It("gets no nacks", func() {
			run()

			Consistently(func() int {
				return rabbitMQ.NackCounter
			}).Should(Equal(0))
		})

		It("uploads the data and converts the track", func() {
			run()

			Eventually(func() bool {
				track, err := trackStore.GetTrack(context.Background(), tracklistID, trackID)
				if err != nil {
					return false
				}

				stemTrack, ok := track.(entity.StemTrack)
				if !ok {
					return false
				}

				if stemTrack.TrackType != entity.FourStemsType {
					return false
				}

				if len(stemTrack.StemURLs) != 4 {
					return false
				}

				for stemName, stemURL := range stemTrack.StemURLs {
					contents, err := fileStore.GetFile(context.Background(), stemURL)
					if err != nil {
						return false
					}

					expectedContent := []byte(string(originalTrackData) + "-" + stemName)
					if bytes.Compare(contents, expectedContent) != 0 {
						return false
					}
				}

				return true
			}).Should(BeTrue())
		})
	})

	Describe("File storage is down", func() {
		BeforeEach(func() {
			fileStore.Unavailable = true
		})

		It("gets 1 ack for the start job", func() {
			run()

			Eventually(func() int {
				return rabbitMQ.AckCounter
			}).Should(Equal(1))
		})

		It("gets 1 nack for the transfer/download job failing", func() {
			run()

			Eventually(func() int {
				return rabbitMQ.NackCounter
			}).Should(Equal(1))
		})

		It("reports the error status", func() {
			run()

			Eventually(func() bool {
				track, err := trackStore.GetTrack(context.Background(), tracklistID, trackID)
				if err != nil {
					return false
				}

				stemTrack, ok := track.(entity.SplitStemTrack)
				if !ok {
					return false
				}

				if stemTrack.TrackType != entity.SplitFourStemsType {
					return false
				}

				if stemTrack.JobStatus != entity.ErrorStatus {
					return false
				}

				return true
			}).Should(BeTrue())
		})
	})
})
