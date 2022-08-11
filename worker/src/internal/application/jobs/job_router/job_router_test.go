package job_router_test

import (
	"context"
	"encoding/json"
	"github.com/streadway/amqp"
	dummy2 "github.com/veedubyou/chord-paper-be/worker/src/internal/application/integration_test/dummy"
	"github.com/veedubyou/chord-paper-be/worker/src/internal/application/jobs/job_message"
	"github.com/veedubyou/chord-paper-be/worker/src/internal/application/jobs/job_router"
	"github.com/veedubyou/chord-paper-be/worker/src/internal/application/jobs/save_stems_to_db"
	"github.com/veedubyou/chord-paper-be/worker/src/internal/application/jobs/save_stems_to_db/save_stems_to_dbfakes"
	"github.com/veedubyou/chord-paper-be/worker/src/internal/application/jobs/split"
	"github.com/veedubyou/chord-paper-be/worker/src/internal/application/jobs/split/splitfakes"
	"github.com/veedubyou/chord-paper-be/worker/src/internal/application/jobs/split/splitter"
	"github.com/veedubyou/chord-paper-be/worker/src/internal/application/jobs/start"
	"github.com/veedubyou/chord-paper-be/worker/src/internal/application/jobs/start/startfakes"
	"github.com/veedubyou/chord-paper-be/worker/src/internal/application/jobs/transfer"
	"github.com/veedubyou/chord-paper-be/worker/src/internal/application/jobs/transfer/transferfakes"
	"github.com/veedubyou/chord-paper-be/worker/src/internal/application/tracks/entity"
	"github.com/veedubyou/chord-paper-be/worker/src/internal/lib/cerr"

	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("JobRouter", func() {
	var (
		tracklistID string
		trackID     string

		startHandler     *startfakes.FakeStartJobHandler
		transferHandler  *transferfakes.FakeTransferJobHandler
		splitHandler     *splitfakes.FakeSplitJobHandler
		saveStemsHandler *save_stems_to_dbfakes.FakeSaveStemsJobHandler

		trackStore *dummy2.TrackStore
		rabbitMQ   *dummy2.RabbitMQ

		jobRouter job_router.JobRouter

		message     amqp.Delivery
		messageJson []byte

		// reusable tests
		WhenJobFails = func(failureSetup func()) {
			Describe("When job fails", func() {
				BeforeEach(failureSetup)

				It("updates the track to error status", func() {
					Expect(message).NotTo(BeZero())

					_ = jobRouter.HandleMessage(message)

					track, err := trackStore.GetTrack(context.Background(), tracklistID, trackID)
					Expect(err).NotTo(HaveOccurred())

					stemTrack, ok := track.(entity.SplitStemTrack)
					Expect(ok).To(BeTrue())

					Expect(stemTrack.JobStatus).To(Equal(entity.ErrorStatus))
				})

				It("returns an error", func() {
					err := jobRouter.HandleMessage(message)
					Expect(err).To(HaveOccurred())
				})

				It("doesn't publish any new jobs", func() {
					Expect(rabbitMQ.MessageChannel).To(BeEmpty())
				})
			})
		}

		ItUpdatesProgress = func() {
			It("updates the progress", func() {
				_ = jobRouter.HandleMessage(message)

				track, err := trackStore.GetTrack(context.Background(), tracklistID, trackID)
				Expect(err).NotTo(HaveOccurred())

				stemTrack, ok := track.(entity.SplitStemTrack)
				Expect(ok).To(BeTrue())

				Expect(stemTrack.JobProgress).To(BeNumerically(">", 0))
			})
		}
	)

	BeforeEach(func() {
		tracklistID = "tracklist-id"
		trackID = "track-id"
		message = amqp.Delivery{}

		var err error
		messageJson, err = json.Marshal(job_message.TrackIdentifier{
			TrackListID: tracklistID,
			TrackID:     trackID,
		})
		Expect(err).NotTo(HaveOccurred())

		By("Initializing the router", func() {
			startHandler = &startfakes.FakeStartJobHandler{}
			transferHandler = &transferfakes.FakeTransferJobHandler{}
			splitHandler = &splitfakes.FakeSplitJobHandler{}
			saveStemsHandler = &save_stems_to_dbfakes.FakeSaveStemsJobHandler{}

			trackStore = dummy2.NewDummyTrackStore()
			rabbitMQ = dummy2.NewRabbitMQ()

			jobRouter = job_router.NewJobRouter(trackStore, rabbitMQ, startHandler, transferHandler, splitHandler, saveStemsHandler)
		})

		By("Setting up the track store", func() {
			track := entity.SplitStemTrack{
				BaseTrack: entity.BaseTrack{
					TrackType: entity.SplitFourStemsType,
				},
				OriginalURL:       "",
				JobStatus:         entity.RequestedStatus,
				JobStatusMessage:  "",
				JobStatusDebugLog: "",
				JobProgress:       0,
			}
			err := trackStore.SetTrack(context.Background(), tracklistID, trackID, track)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Start job", func() {
		BeforeEach(func() {
			message = amqp.Delivery{
				Type: start.JobType,
				Body: messageJson,
			}
		})

		Describe("When job succeeds", func() {
			BeforeEach(func() {
				startHandler.HandleStartJobReturns(start.JobParams{
					TrackIdentifier: job_message.TrackIdentifier{
						TrackListID: tracklistID,
						TrackID:     trackID,
					},
				}, nil)
			})

			It("doesn't return an error", func() {
				err := jobRouter.HandleMessage(message)
				Expect(err).NotTo(HaveOccurred())
			})

			It("publishes the next job", func() {
				_ = jobRouter.HandleMessage(message)
				Expect(rabbitMQ.MessageChannel).To(HaveLen(1))

				nextJob := <-rabbitMQ.MessageChannel
				Expect(nextJob.Type).To(Equal(transfer.JobType))

				var transferJob transfer.JobParams
				err := json.Unmarshal(nextJob.Body, &transferJob)
				Expect(err).NotTo(HaveOccurred())
				Expect(transferJob.TrackListID).To(Equal(tracklistID))
				Expect(transferJob.TrackID).To(Equal(trackID))
			})

			ItUpdatesProgress()
		})

		WhenJobFails(func() {
			startHandler.HandleStartJobReturns(start.JobParams{}, cerr.Error("i failed"))
		})
	})

	Describe("Transfer job", func() {
		BeforeEach(func() {
			message = amqp.Delivery{
				Type: transfer.JobType,
				Body: messageJson,
			}
		})

		Describe("When job succeeds", func() {
			var savedOriginalURL string

			BeforeEach(func() {
				savedOriginalURL = "saved-original-url"
				transferHandler.HandleTransferJobReturns(transfer.JobParams{
					TrackIdentifier: job_message.TrackIdentifier{
						TrackListID: tracklistID,
						TrackID:     trackID,
					},
				}, savedOriginalURL, nil)
			})

			It("doesn't return an error", func() {
				err := jobRouter.HandleMessage(message)
				Expect(err).NotTo(HaveOccurred())
			})

			It("publishes the next job", func() {
				_ = jobRouter.HandleMessage(message)
				Expect(rabbitMQ.MessageChannel).To(HaveLen(1))

				nextJob := <-rabbitMQ.MessageChannel
				Expect(nextJob.Type).To(Equal(split.JobType))

				var splitJob split.JobParams
				err := json.Unmarshal(nextJob.Body, &splitJob)
				Expect(err).NotTo(HaveOccurred())
				Expect(splitJob.TrackListID).To(Equal(tracklistID))
				Expect(splitJob.TrackID).To(Equal(trackID))
				Expect(splitJob.SavedOriginalURL).To(Equal(savedOriginalURL))
			})

			ItUpdatesProgress()
		})

		WhenJobFails(func() {
			transferHandler.HandleTransferJobReturns(transfer.JobParams{}, "", cerr.Error("i failed"))
		})
	})

	Describe("Split job", func() {
		BeforeEach(func() {
			message = amqp.Delivery{
				Type: split.JobType,
				Body: messageJson,
			}
		})

		Describe("When job succeeds", func() {
			var stemURLs splitter.StemFilePaths

			BeforeEach(func() {
				stemURLs = map[string]string{
					"vocals": "vocals.mp3",
					"other":  "other.mp3",
					"bass":   "bass.mp3",
					"drums":  "drums.mp3",
				}

				splitHandler.HandleSplitJobReturns(split.JobParams{
					TrackIdentifier: job_message.TrackIdentifier{
						TrackListID: tracklistID,
						TrackID:     trackID,
					},
					SavedOriginalURL: "saved-original-url",
				}, stemURLs, nil)
			})

			It("doesn't return an error", func() {
				err := jobRouter.HandleMessage(message)
				Expect(err).NotTo(HaveOccurred())
			})

			It("publishes the next job", func() {
				_ = jobRouter.HandleMessage(message)
				Expect(rabbitMQ.MessageChannel).To(HaveLen(1))

				nextJob := <-rabbitMQ.MessageChannel
				Expect(nextJob.Type).To(Equal(save_stems_to_db.JobType))

				var saveStemsJob save_stems_to_db.JobParams
				err := json.Unmarshal(nextJob.Body, &saveStemsJob)
				Expect(err).NotTo(HaveOccurred())
				Expect(saveStemsJob.TrackListID).To(Equal(tracklistID))
				Expect(saveStemsJob.TrackID).To(Equal(trackID))
				Expect(saveStemsJob.StemURLS).To(Equal(stemURLs))
			})

			ItUpdatesProgress()
		})

		WhenJobFails(func() {
			splitHandler.HandleSplitJobReturns(split.JobParams{}, nil, cerr.Error("i failed"))
		})
	})

	Describe("Save stem tracks job", func() {
		BeforeEach(func() {
			message = amqp.Delivery{
				Type: save_stems_to_db.JobType,
				Body: messageJson,
			}
		})

		Describe("When job succeeds", func() {
			BeforeEach(func() {
				saveStemsHandler.HandleSaveStemsToDBJobReturns(nil)
			})

			It("doesn't return an error", func() {
				err := jobRouter.HandleMessage(message)
				Expect(err).NotTo(HaveOccurred())
			})

			It("doesn't publishes the next job", func() {
				_ = jobRouter.HandleMessage(message)
				Expect(rabbitMQ.MessageChannel).To(BeEmpty())
			})

			It("doesn't update progress", func() {
				track, err := trackStore.GetTrack(context.Background(), tracklistID, trackID)
				Expect(err).NotTo(HaveOccurred())

				// technically in the real run this would not be a split stem track anymore
				// because this job would have updated it to a loaded stem track
				// however, this was mocked out and we still want to assert this behaviour
				stemTrack, ok := track.(entity.SplitStemTrack)
				Expect(ok).To(BeTrue())

				Expect(stemTrack.JobProgress).To(BeZero())
			})
		})

		WhenJobFails(func() {
			saveStemsHandler.HandleSaveStemsToDBJobReturns(cerr.Error("i failed"))
		})
	})
})
