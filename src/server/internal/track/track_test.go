package track_test

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/veedubyou/chord-paper-be/src/server/internal/shared_tests/auth"
	"github.com/veedubyou/chord-paper-be/src/server/internal/song/errors"
	"github.com/veedubyou/chord-paper-be/src/server/internal/song/gateway"
	"github.com/veedubyou/chord-paper-be/src/server/internal/song/storage"
	"github.com/veedubyou/chord-paper-be/src/server/internal/song/usecase"
	"github.com/veedubyou/chord-paper-be/src/server/internal/track/errors"
	"github.com/veedubyou/chord-paper-be/src/server/internal/track/gateway"
	"github.com/veedubyou/chord-paper-be/src/server/internal/track/usecase"
	"github.com/veedubyou/chord-paper-be/src/server/internal/user/storage"
	"github.com/veedubyou/chord-paper-be/src/server/internal/user/usecase"
	"github.com/veedubyou/chord-paper-be/src/shared/lib/jsonlib"
	"github.com/veedubyou/chord-paper-be/src/shared/lib/rabbitmq"
	"github.com/veedubyou/chord-paper-be/src/shared/testing"
	"github.com/veedubyou/chord-paper-be/src/shared/track/entity"
	"github.com/veedubyou/chord-paper-be/src/shared/track/storage"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("Track", func() {
	var (
		trackGateway trackgateway.Gateway
		songGateway  songgateway.Gateway
		publisher    rabbitmq.QueuePublisher
		validator    testing.Validator

		consumer testing.RabbitMQConsumer
	)

	BeforeEach(func() {
		validator = testing.Validator{}
		publisher = testing.MakeRabbitMQPublisher(publisherConn)
		consumer = testing.NewRabbitMQConsumer(consumerConn)

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
		testing.ResetDB(db)
		testing.ResetRabbitMQ(publisherConn)
	})

	BeforeEach(func() {
		go consumer.AsyncStart()
	})

	AfterEach(func() {
		consumer.Stop()
	})

	var getTracklist = func(tracklistID string) map[string]any {
		getRequest := testing.RequestFactory{
			Method:  "GET",
			Target:  fmt.Sprintf("/songs/%s/tracklist", tracklistID),
			JSONObj: nil,
		}.MakeFake()

		getResponse := httptest.NewRecorder()
		c := testing.PrepareEchoContext(getRequest, getResponse)
		err := trackGateway.GetTrackList(c, tracklistID)
		Expect(err).NotTo(HaveOccurred())

		return testing.DecodeJSON[map[string]any](getResponse.Body)
	}

	var getTrackSliceFromResponse = func(responseBody map[string]any) []map[string]any {
		if responseBody["tracks"] == nil {
			return nil
		}

		tracks := testing.ExpectType[[]any](responseBody["tracks"])
		trackObjs := []map[string]any{}
		for _, item := range tracks {
			track := testing.ExpectType[map[string]any](item)
			trackObjs = append(trackObjs, track)
		}

		return trackObjs
	}

	var createSong = func(songPayload map[string]any) (string, map[string]any) {
		response := httptest.NewRecorder()

		By("First creating a song", func() {
			request := testing.RequestFactory{
				Method:  "POST",
				Target:  "/songs",
				JSONObj: songPayload,
				Mods:    testing.RequestModifiers{testing.WithUserCred(testing.PrimaryUser)},
			}.MakeFake()

			c := testing.PrepareEchoContext(request, response)

			err := songGateway.CreateSong(c)
			Expect(err).NotTo(HaveOccurred())
		})

		By("Extracting the ID from the created song")
		song := testing.DecodeJSON[map[string]any](response.Body)

		songID := testing.ExpectType[string](song["id"])
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
				createSongPayload := testing.LoadDemoSong()
				songID, _ = createSong(createSongPayload)

				tracklist.Defined.SongID = songID
			})

			Describe("Unpermitted requests", func() {
				BeforeEach(func() {
					jsonBody := testing.ExpectSuccess(tracklist.ToMap())

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
					requestFactory testing.RequestFactory
				)

				BeforeEach(func() {
					requestFactory = testing.RequestFactory{
						Method:  "POST",
						Target:  "/songs/:id/tracklist",
						JSONObj: nil,
					}
				})

				var setTracklist = func() {
					requestFactory.JSONObj = testing.ExpectSuccess(tracklist.ToMap())
					request := requestFactory.MakeFake()
					response = httptest.NewRecorder()
					c := testing.PrepareEchoContext(request, response)

					err := trackGateway.SetTrackList(c, songID)
					Expect(err).NotTo(HaveOccurred())
				}

				JustBeforeEach(func() {
					setTracklist()
				})

				Describe("For an authorized owner", func() {
					BeforeEach(func() {
						requestFactory.Mods.Add(testing.WithUserCred(testing.PrimaryUser))
					})

					var ItSavesSuccessfully = func() {
						It("returns success", func() {
							Expect(response.Code).To(Equal(http.StatusOK))
						})

						It("returns the same tracklist that was sent", func() {
							responseBody := testing.DecodeJSON[map[string]any](response.Body)
							responseTracks := getTrackSliceFromResponse(responseBody)
							Expect(tracklist.Defined.Tracks).To(HaveLen(len(responseTracks)))

							By("copying over the new IDs", func() {
								for i, expectedTrack := range tracklist.Defined.Tracks {
									if expectedTrack.GetID() != "" {
										continue
									}

									actualTrack := responseTracks[i]
									actualTrackID := testing.ExpectType[string](actualTrack["id"])

									switch t := expectedTrack.(type) {
									case *trackentity.GenericTrack:
										t.Defined.ID = actualTrackID
									case *trackentity.SplitRequestTrack:
										t.ID = actualTrackID
									case *trackentity.StemTrack:
										t.ID = actualTrackID
									default:
										panic("Unrecognized track type")
									}
								}
							})

							expectedTracklist := testing.ExpectSuccess(tracklist.ToMap())
							Expect(responseBody).To(Equal(expectedTracklist))
						})

						It("returns tracks that all have IDs", func() {
							responseBody := testing.DecodeJSON[map[string]any](response.Body)
							tracks := getTrackSliceFromResponse(responseBody)

							for _, track := range tracks {
								trackID := testing.ExpectType[string](track["id"])
								Expect(trackID).NotTo(BeZero())
							}
						})

						It("persists and can be retrieved after", func() {
							setResponseBody := testing.DecodeJSON[map[string]any](response.Body)

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
							track := trackentity.StemTrack{}
							track.TrackType = "4stems"

							for i := 0; i < 11; i++ {
								tracklist.Defined.Tracks = append(tracklist.Defined.Tracks, &track)
							}
						})

						It("fails with the right error code", func() {
							resErr := testing.DecodeJSONError(response.Body)
							Expect(resErr.Code).To(BeEquivalentTo(trackerrors.TrackListSizeExceeded))
						})

						It("fails with the right status code", func() {
							Expect(response.Code).To(Equal(http.StatusBadRequest))
						})

						ItDoesntQueueMessages()
					})

					Describe("A tracklist with tracks", func() {
						var (
							track0 trackentity.StemTrack
							track1 trackentity.GenericTrack
							track2 trackentity.GenericTrack
						)

						BeforeEach(func() {
							track0 = trackentity.StemTrack{}
							track0.TrackType = "4stems"

							track1 = trackentity.GenericTrack{}
							track1.Defined.TrackType = "accompaniment"
							track1.Extra = map[string]any{
								"accompaniment_url": "accompaniment.mp3",
							}

							track2 = trackentity.GenericTrack{}
							track2.Defined.TrackType = "original"
							track2.Extra = map[string]any{
								"url": "song.mp3",
							}

							tracklist.Defined.Tracks = []trackentity.Track{
								&track0, &track1, &track2,
							}
						})

						ItSavesSuccessfully()
						ItDoesntQueueMessages()

						Describe("Updating a second time", func() {
							var (
								newTrack trackentity.GenericTrack
							)

							BeforeEach(func() {
								By("Setting the tracklist the first time", func() {
									setTracklist()

									newTrack = trackentity.GenericTrack{}
									newTrack.Defined.TrackType = "new-type"
									newTrack.Extra = map[string]any{
										"amuro": "ray.mp4",
									}

									tracklist.Defined.Tracks[1] = &newTrack
								})
							})

							ItSavesSuccessfully()
							ItDoesntQueueMessages()

							It("doesn't include the overwritten track anymore", func() {
								updatedTrackList := getTracklist(songID)
								updatedTracks := testing.ExpectType[[]any](updatedTrackList["tracks"])
								updatedTrack1 := testing.ExpectType[map[string]any](updatedTracks[1])

								originalTrack1 := testing.ExpectSuccess(track1.ToMap())
								Expect(updatedTrack1).NotTo(Equal(originalTrack1))
							})
						})

						Describe("With split requests", func() {
							var (
								splitRequestTrack trackentity.SplitRequestTrack
							)

							BeforeEach(func() {
								splitRequestTrack = trackentity.SplitRequestTrack{}
								splitRequestTrack.TrackType = "split_4stems"
								splitRequestTrack.OriginalURL = "thisplace.com/song.mp3"
								splitRequestTrack.InitializeRequest()

								tracklist.Defined.Tracks = append(tracklist.Defined.Tracks, &splitRequestTrack)
							})

							ItSavesSuccessfully()

							It("queues a start job message", func() {
								setResponseBody := testing.DecodeJSON[map[string]any](response.Body)
								tracks := getTrackSliceFromResponse(setResponseBody)
								Expect(tracks).To(HaveLen(4))
								splitRequestID := testing.ExpectType[string](tracks[3]["id"])

								expectedMessage := map[string]any{
									"tracklist_id": songID,
									"track_id":     splitRequestID,
								}

								Eventually(consumer.Unload).Should(Equal([]testing.ReceivedMessage{
									{
										Type:    "start_job",
										Message: expectedMessage,
									},
								}))

								Consistently(consumer.Unload).Should(BeEmpty())
							})

							Describe("Updating a second time with an existing split request", func() {
								BeforeEach(func() {
									By("first setting the tracklist", func() {
										setTracklist()
									})

									By("unloading any existing job messages", func() {
										Eventually(consumer.Unload).Should(HaveLen(1))
									})

									By("changing the split request job", func() {
										setResponseBody := testing.DecodeJSON[map[string]any](response.Body)
										tracks := getTrackSliceFromResponse(setResponseBody)
										tracks[3]["retry_times"] = 5
										tracklist = trackentity.TrackList{}
										err := tracklist.FromMap(setResponseBody)
										Expect(err).NotTo(HaveOccurred())
									})
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

				requestFactory := testing.RequestFactory{
					Method:  "POST",
					Target:  "/songs/:id/tracklist",
					JSONObj: testing.ExpectSuccess(tracklist.ToMap()),
				}

				requestFactory.Mods.Add(testing.WithUserCred(testing.PrimaryUser))

				request := requestFactory.MakeFake()
				response = httptest.NewRecorder()
				c := testing.PrepareEchoContext(request, response)

				err := trackGateway.SetTrackList(c, tracklist.Defined.SongID)
				Expect(err).NotTo(HaveOccurred())
			})

			It("fails with the right error code", func() {
				resErr := testing.DecodeJSONError(response.Body)
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
				jsonObj := testing.ExpectSuccess(jsonlib.StructToMap(struct {
					SongID string   `json:"song_id"`
					Tracks []string `json:"tracks"`
				}{
					SongID: "hmm",
					Tracks: []string{"track1", "track2"},
				}))

				requestFactory := testing.RequestFactory{
					Method:  "POST",
					Target:  "/songs/:id/tracklist",
					JSONObj: jsonObj,
				}

				requestFactory.Mods.Add(testing.WithUserCred(testing.PrimaryUser))

				request := requestFactory.MakeFake()
				response = httptest.NewRecorder()
				c := testing.PrepareEchoContext(request, response)

				err := trackGateway.SetTrackList(c, uuid.New().String())
				Expect(err).NotTo(HaveOccurred())
			})

			It("fails with the right error code", func() {
				resErr := testing.DecodeJSONError(response.Body)
				Expect(resErr.Code).To(BeEquivalentTo(trackerrors.BadTracklistDataCode))
			})

			It("fails with the right status code", func() {
				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})

			ItDoesntQueueMessages()
		})
	})
})
