package split_test

import (
	"context"
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/veedubyou/chord-paper-be/src/shared/config/prod"
	trackentity "github.com/veedubyou/chord-paper-be/src/shared/track/entity"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/integration_test/dummy"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/job_message"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/split"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/split/splitter"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/split/splitter/file_splitter"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/lib/storagepath"
)

var _ = Describe("Split handler", func() {
	var (
		bucketName string

		dummyTrackStore *dummy.TrackStore
		dummyFileStore  *dummy.FileStore
		dummyExecutor   *dummy.SpleeterExecutor

		handler split.JobHandler

		message           []byte
		savedOriginalURL  string
		remoteURLBase     string
		originalTrackData []byte

		tracklistID string
		trackID     string
		trackType   trackentity.SplitRequestType
	)

	BeforeEach(func() {
		By("Assigning all the variables data", func() {
			tracklistID = "tracklist-ID"
			trackID = "track-ID"
			trackType = ""
			bucketName = "bucket-head"

			remoteURLBase = fmt.Sprintf("%s/%s/%s/%s", prod.GOOGLE_STORAGE_HOST, bucketName, tracklistID, trackID)
			savedOriginalURL = fmt.Sprintf("%s/original/original.mp3", remoteURLBase)
			originalTrackData = []byte("cool_jamz")
		})

		By("Instantiating all mocks", func() {
			dummyTrackStore = dummy.NewDummyTrackStore()
			dummyFileStore = dummy.NewDummyFileStore()
			dummyExecutor = dummy.NewDummySpleeterExecutor()
		})

		By("Setting up file on the file store", func() {
			err := dummyFileStore.WriteFile(context.Background(), savedOriginalURL, originalTrackData)
			Expect(err).NotTo(HaveOccurred())
		})

		By("Instantiating the handler", func() {
			localSplitter, err := file_splitter.NewLocalFileSplitter(workingDir, "/somewhere/spleeter", "/somewhere/demucs", dummyExecutor)
			Expect(err).NotTo(HaveOccurred())

			remoteSplitter, err := file_splitter.NewRemoteFileSplitter(workingDir, dummyFileStore, localSplitter)
			Expect(err).NotTo(HaveOccurred())

			pathGenerator := storagepath.Generator{
				Host:   prod.GOOGLE_STORAGE_HOST,
				Bucket: bucketName,
			}
			trackSplitter := splitter.NewTrackSplitter(remoteSplitter, dummyTrackStore, pathGenerator)
			handler = split.NewJobHandler(trackSplitter)
		})
	})

	JustBeforeEach(func() {
		Expect(trackType).NotTo(Equal(BeZero()))

		prevUnavailable := dummyTrackStore.Unavailable
		dummyTrackStore.Unavailable = false

		tracklist := trackentity.TrackList{}
		tracklist.Defined.SongID = tracklistID
		tracklist.Defined.Tracks = trackentity.Tracks{
			&trackentity.SplitRequestTrack{
				TrackFields: trackentity.TrackFields{ID: trackID},
				TrackType:   trackType,
				EngineType:  trackentity.SpleeterType,
				OriginalURL: "https://whocares",
			},
		}

		err := dummyTrackStore.SetTrackList(context.Background(), tracklist)
		dummyTrackStore.Unavailable = prevUnavailable

		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Well formed message", func() {
		BeforeEach(func() {
			job := split.JobParams{
				TrackIdentifier: job_message.TrackIdentifier{
					TrackListID: tracklistID,
					TrackID:     trackID,
				},
				SavedOriginalURL: savedOriginalURL,
			}

			var err error
			message, err = json.Marshal(job)
			Expect(err).NotTo(HaveOccurred())

			// just setting something for now so that other paths
			// don't run into an error
			trackType = trackentity.SplitRequestType(trackentity.TwoStemsType)
		})

		Describe("Happy path", func() {
			var (
				err               error
				returnedJobParams split.JobParams
				returnedStemUrls  splitter.StemFilePaths

				expectedReturnedStemUrls splitter.StemFilePaths
				expectedStemFileContent  map[string][]byte

				expectUploadedStemFiles = func() {
					Expect(expectedStemFileContent).NotTo(BeEmpty())
					for stemURL, stemFileContent := range expectedStemFileContent {
						storedBytes, err := dummyFileStore.GetFile(context.Background(), stemURL)
						Expect(err).NotTo(HaveOccurred())
						Expect(storedBytes).To(Equal(stemFileContent))
					}
				}

				expectReturnValues = func() {
					Expect(returnedJobParams.TrackListID).To(Equal(tracklistID))
					Expect(returnedJobParams.TrackID).To(Equal(trackID))
					Expect(returnedStemUrls).To(Equal(expectedReturnedStemUrls))
				}
			)

			BeforeEach(func() {
				err = nil
				returnedJobParams = split.JobParams{}
				returnedStemUrls = nil
			})

			JustBeforeEach(func() {
				returnedJobParams, returnedStemUrls, err = handler.HandleSplitJob(message)
			})

			Describe("2stems", func() {
				BeforeEach(func() {
					trackType = trackentity.SplitTwoStemsType

					vocalsURL := remoteURLBase + "/2stems/vocals.mp3"
					accompanimentURL := remoteURLBase + "/2stems/accompaniment.mp3"

					expectedReturnedStemUrls = map[string]string{
						"vocals":        vocalsURL,
						"accompaniment": accompanimentURL,
					}

					expectedStemFileContent = map[string][]byte{
						vocalsURL:        []byte(string(originalTrackData) + "-vocals"),
						accompanimentURL: []byte(string(originalTrackData) + "-accompaniment"),
					}
				})

				It("succeeds", func() {
					Expect(err).NotTo(HaveOccurred())
				})

				It("uploaded the stem files", expectUploadedStemFiles)

				It("returns the right values", expectReturnValues)
			})

			Describe("4stems", func() {
				BeforeEach(func() {
					trackType = trackentity.SplitFourStemsType

					vocalsURL := remoteURLBase + "/4stems/vocals.mp3"
					otherURL := remoteURLBase + "/4stems/other.mp3"
					bassURL := remoteURLBase + "/4stems/bass.mp3"
					drumsURL := remoteURLBase + "/4stems/drums.mp3"

					expectedReturnedStemUrls = map[string]string{
						"vocals": vocalsURL,
						"other":  otherURL,
						"bass":   bassURL,
						"drums":  drumsURL,
					}

					expectedStemFileContent = map[string][]byte{
						vocalsURL: []byte(string(originalTrackData) + "-vocals"),
						otherURL:  []byte(string(originalTrackData) + "-other"),
						bassURL:   []byte(string(originalTrackData) + "-bass"),
						drumsURL:  []byte(string(originalTrackData) + "-drums"),
					}
				})

				It("succeeds", func() {
					Expect(err).NotTo(HaveOccurred())
				})

				It("uploaded the stem files", expectUploadedStemFiles)

				It("returns the right values", expectReturnValues)
			})

			Describe("5stems", func() {
				BeforeEach(func() {
					trackType = trackentity.SplitFiveStemsType

					vocalsURL := remoteURLBase + "/5stems/vocals.mp3"
					otherURL := remoteURLBase + "/5stems/other.mp3"
					pianoURL := remoteURLBase + "/5stems/piano.mp3"
					bassURL := remoteURLBase + "/5stems/bass.mp3"
					drumsURL := remoteURLBase + "/5stems/drums.mp3"

					expectedReturnedStemUrls = map[string]string{
						"vocals": vocalsURL,
						"other":  otherURL,
						"piano":  pianoURL,
						"bass":   bassURL,
						"drums":  drumsURL,
					}

					expectedStemFileContent = map[string][]byte{
						vocalsURL: []byte(string(originalTrackData) + "-vocals"),
						otherURL:  []byte(string(originalTrackData) + "-other"),
						pianoURL:  []byte(string(originalTrackData) + "-piano"),
						bassURL:   []byte(string(originalTrackData) + "-bass"),
						drumsURL:  []byte(string(originalTrackData) + "-drums"),
					}
				})

				It("succeeds", func() {
					Expect(err).NotTo(HaveOccurred())
				})

				It("uploaded the stem files", expectUploadedStemFiles)

				It("returns the right values", expectReturnValues)
			})
		})

		Describe("When the file store is down", func() {
			BeforeEach(func() {
				dummyFileStore.Unavailable = true
			})

			It("returns an error", func() {
				_, _, err := handler.HandleSplitJob(message)
				Expect(err).To(HaveOccurred())
			})
		})

		Describe("When the track store is down", func() {
			BeforeEach(func() {
				dummyTrackStore.Unavailable = true
			})

			It("returns an error", func() {
				_, _, err := handler.HandleSplitJob(message)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Malformed message", func() {
		BeforeEach(func() {
			job := split.JobParams{
				TrackIdentifier: job_message.TrackIdentifier{
					TrackListID: tracklistID,
					TrackID:     trackID,
				},
			}

			var err error
			message, err = json.Marshal(job)
			Expect(err).NotTo(HaveOccurred())

			trackType = trackentity.SplitRequestType(trackentity.TwoStemsType)
		})

		It("failaroo", func() {
			_, _, err := handler.HandleSplitJob(message)
			Expect(err).To(HaveOccurred())
		})
	})
})
