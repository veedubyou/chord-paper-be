package track_test

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/veedubyou/chord-paper-be/src/server/internal/lib/jsonlib"
	"github.com/veedubyou/chord-paper-be/src/server/internal/lib/rabbitmq"
	. "github.com/veedubyou/chord-paper-be/src/server/internal/lib/testing"
	"github.com/veedubyou/chord-paper-be/src/server/internal/shared_tests/auth"
	"github.com/veedubyou/chord-paper-be/src/server/internal/song/errors"
	"github.com/veedubyou/chord-paper-be/src/server/internal/song/gateway"
	"github.com/veedubyou/chord-paper-be/src/server/internal/song/storage"
	"github.com/veedubyou/chord-paper-be/src/server/internal/song/usecase"
	"github.com/veedubyou/chord-paper-be/src/server/internal/track/entity"
	"github.com/veedubyou/chord-paper-be/src/server/internal/track/errors"
	"github.com/veedubyou/chord-paper-be/src/server/internal/track/gateway"
	"github.com/veedubyou/chord-paper-be/src/server/internal/track/storage"
	"github.com/veedubyou/chord-paper-be/src/server/internal/track/usecase"
	"github.com/veedubyou/chord-paper-be/src/server/internal/user/storage"
	"github.com/veedubyou/chord-paper-be/src/server/internal/user/usecase"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("Track", func() {
	var (
		trackGateway trackgateway.Gateway
		songGateway  songgateway.Gateway
		publisher    rabbitmq.Publisher
		validator    TestingValidator

		consumer RabbitMQConsumer
	)

	BeforeEach(func() {
		validator = TestingValidator{}
		publisher = MakeRabbitMQPublisher(publisherConn)
		consumer = NewRabbitMQConsumer(consumerConn)

		userStorage := userstorage.NewDB(db)
		userUsecase := userusecase.NewUsecase(userStorage, validator)

		songStorage := songstorage.NewDB(db)
		songUsecase := songusecase.NewUsecase(songStorage, userUsecase)
		songGateway = songgateway.NewGateway(songUsecase)

		trackStorage := trackstorage.NewDB(db)
		trackUsecase := trackusecase.NewUsecase(trackStorage, songUsecase, publisher)
		trackGateway = trackgateway.NewGateway(trackUsecase)
	})

	BeforeEach(func() {
		ResetDB(db)
		ResetRabbitMQ(publisherConn)
	})

	BeforeEach(func() {
		go consumer.AsyncStart()
	})

	AfterEach(func() {
		consumer.Stop()
	})

	var getTracklist = func(tracklistID string) map[string]interface{} {
		getRequest := RequestFactory{
			Method:  "GET",
			Path:    fmt.Sprintf("/songs/%s/tracklist", tracklistID),
			JSONObj: nil,
		}.Make()

		getResponse := httptest.NewRecorder()
		c := PrepareEchoContext(getRequest, getResponse)
		err := trackGateway.GetTrackList(c, tracklistID)
		Expect(err).NotTo(HaveOccurred())

		return DecodeJSON[map[string]interface{}](getResponse)
	}

	var getTrackSliceFromResponse = func(responseBody map[string]interface{}) []map[string]interface{} {
		if responseBody["tracks"] == nil {
			return nil
		}

		tracks := ExpectType[[]interface{}](responseBody["tracks"])
		trackObjs := []map[string]interface{}{}
		for _, item := range tracks {
			track := ExpectType[map[string]interface{}](item)
			trackObjs = append(trackObjs, track)
		}

		return trackObjs
	}

	var createSong = func(songPayload map[string]interface{}) (string, map[string]interface{}) {
		By("First creating a song")

		request := RequestFactory{
			Method:  "POST",
			Path:    "/songs",
			JSONObj: songPayload,
			Mods:    RequestModifiers{WithUserCred(PrimaryUser)},
		}.Make()
		response := httptest.NewRecorder()
		c := PrepareEchoContext(request, response)

		err := songGateway.CreateSong(c)
		Expect(err).NotTo(HaveOccurred())

		By("Extracting the ID from the created song")
		song := DecodeJSON[map[string]interface{}](response)

		songID := ExpectType[string](song["id"])
		Expect(songID).NotTo(BeEmpty())
		return songID, song
	}

	var ItDoesntQueueMessages = func() {
		It("doesn't queue any messages in rabbitmq", func() {
			Consistently(consumer.Unload).Should(BeEmpty())
		})
	}

	Describe("Set Tracklist", func() {
		var (
			tracklist trackentity.TrackList
		)

		BeforeEach(func() {
			tracklist = trackentity.TrackList{}
			tracklist.Defined.SongID = ""
			tracklist.Defined.Tracks = nil
		})

		Describe("With an existing attached song", func() {
			var (
				songID string
			)

			BeforeEach(func() {
				createSongPayload := LoadDemoSong()
				songID, _ = createSong(createSongPayload)

				tracklist.Defined.SongID = songID
			})

			Describe("Unpermitted requests", func() {
				BeforeEach(func() {
					jsonBody := ExpectSuccess(tracklist.ToMap())

					authtest.Endpoint = func(c echo.Context) error {
						return trackGateway.SetTrackList(c, songID)
					}
					authtest.JSONBody = jsonBody
				})

				authtest.ItRejectsUnpermittedRequests("POST", "/songs/:id/tracklist")
			})

			Describe("Authorized", func() {
				var (
					response       *httptest.ResponseRecorder
					requestFactory RequestFactory
				)

				BeforeEach(func() {
					requestFactory = RequestFactory{
						Method:  "POST",
						Path:    "/songs/:id/tracklist",
						JSONObj: nil,
					}
				})

				var setTracklist = func() {
					requestFactory.JSONObj = ExpectSuccess(tracklist.ToMap())
					request := requestFactory.Make()
					response = httptest.NewRecorder()
					c := PrepareEchoContext(request, response)

					err := trackGateway.SetTrackList(c, songID)
					Expect(err).NotTo(HaveOccurred())
				}

				JustBeforeEach(func() {
					setTracklist()
				})

				Describe("For an authorized owner", func() {
					BeforeEach(func() {
						requestFactory.Mods.Add(WithUserCred(PrimaryUser))
					})

					var ItSavesSuccessfully = func() {
						It("returns success", func() {
							Expect(response.Code).To(Equal(http.StatusOK))
						})

						It("returns the same tracklist that was sent", func() {
							responseBody := DecodeJSON[map[string]interface{}](response)
							responseTracks := getTrackSliceFromResponse(responseBody)
							Expect(tracklist.Defined.Tracks).To(HaveLen(len(responseTracks)))

							By("copying over the new IDs")
							for i := range tracklist.Defined.Tracks {
								expectedTrack := &tracklist.Defined.Tracks[i]
								if expectedTrack.Defined.ID != "" {
									continue
								}

								actualTrack := responseTracks[i]
								expectedTrack.Defined.ID = ExpectType[string](actualTrack["id"])
							}

							expectedTracklist := ExpectSuccess(tracklist.ToMap())
							Expect(responseBody).To(Equal(expectedTracklist))
						})

						It("returns tracks that all have IDs", func() {
							responseBody := DecodeJSON[map[string]interface{}](response)
							tracks := getTrackSliceFromResponse(responseBody)

							for _, track := range tracks {
								trackID := ExpectType[string](track["id"])
								Expect(trackID).NotTo(BeZero())
							}
						})

						It("persists and can be retrieved after", func() {
							setResponseBody := DecodeJSON[map[string]interface{}](response)

							getResponseBody := getTracklist(songID)
							Expect(getResponseBody).To(Equal(setResponseBody))
						})
					}

					Describe("An empty tracklist", func() {
						ItSavesSuccessfully()
						ItDoesntQueueMessages()
					})

					Describe("Too many tracks in the tracklist", func() {
						BeforeEach(func() {
							track := trackentity.Track{}
							track.Defined.TrackType = "4stems"

							for i := 0; i < 11; i++ {
								tracklist.Defined.Tracks = append(tracklist.Defined.Tracks, track)
							}
						})

						It("fails with the right error code", func() {
							resErr := DecodeJSONError(response)
							Expect(resErr.Code).To(BeEquivalentTo(trackerrors.TrackListSizeExceeded))
						})

						It("fails with the right status code", func() {
							Expect(response.Code).To(Equal(http.StatusBadRequest))
						})

						ItDoesntQueueMessages()
					})

					Describe("A tracklist with tracks", func() {
						var (
							track0 trackentity.Track
							track1 trackentity.Track
							track2 trackentity.Track
						)

						BeforeEach(func() {
							track0 = trackentity.Track{}
							track0.Defined.TrackType = "4stems"
							track0.Extra = map[string]interface{}{}

							track1 = trackentity.Track{}
							track1.Defined.TrackType = "accompaniment"
							track1.Extra = map[string]interface{}{
								"accompaniment_url": "accompaniment.mp3",
							}

							track2 = trackentity.Track{}
							track2.Defined.TrackType = "original"
							track2.Extra = map[string]interface{}{
								"url": "song.mp3",
							}

							tracklist.Defined.Tracks = []trackentity.Track{
								track0, track1, track2,
							}
						})

						ItSavesSuccessfully()
						ItDoesntQueueMessages()

						Describe("Updating a second time", func() {
							var (
								newTrack trackentity.Track
							)

							BeforeEach(func() {
								By("Setting the tracklist the first time")
								setTracklist()

								newTrack = trackentity.Track{}
								newTrack.Defined.TrackType = "new-type"
								newTrack.Extra = map[string]interface{}{
									"amuro": "ray.mp4",
								}

								tracklist.Defined.Tracks[1] = newTrack
							})

							ItSavesSuccessfully()
							ItDoesntQueueMessages()

							It("doesn't include the overwritten track anymore", func() {
								updatedTrackList := getTracklist(songID)
								updatedTracks := ExpectType[[]interface{}](updatedTrackList["tracks"])
								updatedTrack1 := ExpectType[map[string]interface{}](updatedTracks[1])

								originalTrack1 := ExpectSuccess(track1.ToMap())
								Expect(updatedTrack1).NotTo(Equal(originalTrack1))
							})
						})

						Describe("With split requests", func() {
							var (
								splitRequestTrack trackentity.Track
							)

							BeforeEach(func() {
								splitRequestTrack = trackentity.Track{}
								splitRequestTrack.Defined.TrackType = "split_4stems"
								splitRequestTrack.Extra = map[string]interface{}{
									"original_url": "thisplace.com/song.mp3",
								}
								splitRequestTrack.InitializeSplitJob()

								tracklist.Defined.Tracks = append(tracklist.Defined.Tracks, splitRequestTrack)
							})

							ItSavesSuccessfully()

							It("queues a start job message", func() {
								setResponseBody := DecodeJSON[map[string]interface{}](response)
								tracks := getTrackSliceFromResponse(setResponseBody)
								Expect(tracks).To(HaveLen(4))
								splitRequestID := ExpectType[string](tracks[3]["id"])

								expectedMessage := map[string]interface{}{
									"tracklist_id": songID,
									"track_id":     splitRequestID,
								}

								Eventually(consumer.Unload).Should(Equal([]ReceivedMessage{
									{
										Type:    "start_job",
										Message: expectedMessage,
									},
								}))

								Consistently(consumer.Unload).Should(BeEmpty())
							})

							Describe("Updating a second time with an existing split request", func() {
								BeforeEach(func() {
									By("first setting the tracklist")
									setTracklist()

									By("unloading any existing job messages")
									Eventually(consumer.Unload).Should(HaveLen(1))

									By("changing the split request job")
									setResponseBody := DecodeJSON[map[string]interface{}](response)
									tracks := getTrackSliceFromResponse(setResponseBody)
									tracks[3]["retry_times"] = 5
									tracklist = trackentity.TrackList{}
									err := tracklist.FromMap(setResponseBody)
									Expect(err).NotTo(HaveOccurred())
								})

								ItSavesSuccessfully()
								ItDoesntQueueMessages()
							})
						})
					})
				})
			})
		})

		Describe("Without an existing attached song", func() {
			var (
				response *httptest.ResponseRecorder
			)

			BeforeEach(func() {
				tracklist.Defined.SongID = uuid.New().String()

				requestFactory := RequestFactory{
					Method:  "POST",
					Path:    "/songs/:id/tracklist",
					JSONObj: ExpectSuccess(tracklist.ToMap()),
				}

				requestFactory.Mods.Add(WithUserCred(PrimaryUser))

				request := requestFactory.Make()
				response = httptest.NewRecorder()
				c := PrepareEchoContext(request, response)

				err := trackGateway.SetTrackList(c, tracklist.Defined.SongID)
				Expect(err).NotTo(HaveOccurred())
			})

			It("fails with the right error code", func() {
				resErr := DecodeJSONError(response)
				Expect(resErr.Code).To(BeEquivalentTo(songerrors.SongNotFoundCode))
			})

			It("fails with the right status code", func() {
				Expect(response.Code).To(Equal(http.StatusNotFound))
			})

			ItDoesntQueueMessages()
		})

		Describe("Bad tracklist data", func() {
			var (
				response *httptest.ResponseRecorder
			)

			BeforeEach(func() {
				By("making a deliberately wrongly typed tracks field")
				jsonObj := ExpectSuccess(jsonlib.StructToMap(struct {
					SongID string   `json:"song_id"`
					Tracks []string `json:"tracks"`
				}{
					SongID: "hmm",
					Tracks: []string{"track1", "track2"},
				}))

				requestFactory := RequestFactory{
					Method:  "POST",
					Path:    "/songs/:id/tracklist",
					JSONObj: jsonObj,
				}

				requestFactory.Mods.Add(WithUserCred(PrimaryUser))

				request := requestFactory.Make()
				response = httptest.NewRecorder()
				c := PrepareEchoContext(request, response)

				err := trackGateway.SetTrackList(c, uuid.New().String())
				Expect(err).NotTo(HaveOccurred())
			})

			It("fails with the right error code", func() {
				resErr := DecodeJSONError(response)
				Expect(resErr.Code).To(BeEquivalentTo(trackerrors.BadTracklistDataCode))
			})

			It("fails with the right status code", func() {
				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})

			ItDoesntQueueMessages()
		})
	})
})
