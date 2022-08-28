package song_test

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/veedubyou/chord-paper-be/src/server/api_error"
	"github.com/veedubyou/chord-paper-be/src/server/internal/errors/auth"
	"github.com/veedubyou/chord-paper-be/src/server/internal/shared_tests/auth"
	"github.com/veedubyou/chord-paper-be/src/server/internal/song/errors"
	"github.com/veedubyou/chord-paper-be/src/server/internal/song/gateway"
	"github.com/veedubyou/chord-paper-be/src/server/internal/song/storage"
	"github.com/veedubyou/chord-paper-be/src/server/internal/song/usecase"
	"github.com/veedubyou/chord-paper-be/src/server/internal/user/storage"
	"github.com/veedubyou/chord-paper-be/src/server/internal/user/usecase"
	"github.com/veedubyou/chord-paper-be/src/shared/testing"
	"net/http"
	"net/http/httptest"
	"time"
)

var _ = Describe("Song", func() {
	var (
		songGateway songgateway.Gateway
		validator   testing.Validator
	)

	BeforeEach(func() {
		validator = testing.Validator{}
		userStorage := userstorage.NewDB(db)
		userUsecase := userusecase.NewUsecase(userStorage, validator)

		songStorage := songstorage.NewDB(db)
		songUsecase := songusecase.NewUsecase(songStorage, userUsecase)
		songGateway = songgateway.NewGateway(songUsecase)
	})

	var getSong = func(songID string) map[string]any {
		getRequest := testing.RequestFactory{
			Method:  "GET",
			Target:  fmt.Sprintf("/songs/%s", songID),
			JSONObj: nil,
		}.MakeFake()

		getResponse := httptest.NewRecorder()
		c := testing.PrepareEchoContext(getRequest, getResponse)
		err := songGateway.GetSong(c, songID)
		Expect(err).NotTo(HaveOccurred())

		return testing.DecodeJSON[map[string]any](getResponse.Body)
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

	BeforeEach(func() {
		testing.ResetDB(db)
	})

	Describe("Get Song", func() {
		Describe("For non-existing songs", func() {
			var (
				response *httptest.ResponseRecorder
				songID   string
			)

			BeforeEach(func() {
				songID = ""
				createSong(testing.LoadDemoSong())
			})

			JustBeforeEach(func() {
				requestFactory := testing.RequestFactory{
					Method:  "GET",
					Target:  "/songs/:id",
					JSONObj: nil,
				}

				request := requestFactory.MakeFake()
				response = httptest.NewRecorder()
				c := testing.PrepareEchoContext(request, response)

				err := songGateway.GetSong(c, songID)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				songID = ""
			})

			Describe("For an empty ID", func() {
				BeforeEach(func() {
					songID = ""
				})

				It("fails with the right error code", func() {
					resErr := testing.DecodeJSONError(response.Body)
					Expect(resErr.Code).To(BeEquivalentTo(songerrors.SongNotFoundCode))
				})

				It("fails with the right status code", func() {
					Expect(response.Code).To(Equal(http.StatusNotFound))
				})
			})

			Describe("For a malformed ID", func() {
				BeforeEach(func() {
					songID = "boat"
				})

				It("fails with the right error code", func() {
					resErr := testing.DecodeJSONError(response.Body)
					Expect(resErr.Code).To(BeEquivalentTo(songerrors.SongNotFoundCode))
				})

				It("fails with the right status code", func() {
					Expect(response.Code).To(Equal(http.StatusNotFound))
				})
			})

			Describe("For a song that's not there", func() {
				BeforeEach(func() {
					songID = uuid.New().String()
				})

				It("fails with the right error code", func() {
					resErr := testing.DecodeJSONError(response.Body)
					Expect(resErr.Code).To(BeEquivalentTo(songerrors.SongNotFoundCode))
				})

				It("fails with the right status code", func() {
					Expect(response.Code).To(Equal(http.StatusNotFound))
				})
			})
		})
	})

	Describe("Get Song Summaries for User", func() {
		Describe("Unauthorized", func() {
			BeforeEach(func() {
				authtest.Endpoint = func(c echo.Context) error {
					return songGateway.GetSongSummariesForUser(c, testing.PrimaryUser.ID)
				}
			})

			authtest.ItRejectsUnpermittedRequests("GET", "/users/:id/songs")
		})

		Describe("Authorized", func() {
			var (
				response       *httptest.ResponseRecorder
				requestFactory testing.RequestFactory
			)

			BeforeEach(func() {
				requestFactory = testing.RequestFactory{
					Method:  "GET",
					Target:  fmt.Sprintf("/users/%s/songs", testing.PrimaryUser.ID),
					JSONObj: nil,
					Mods:    testing.RequestModifiers{testing.WithUserCred(testing.PrimaryUser)},
				}
			})

			JustBeforeEach(func() {
				request := requestFactory.MakeFake()
				response = httptest.NewRecorder()
				c := testing.PrepareEchoContext(request, response)

				err := songGateway.GetSongSummariesForUser(c, testing.PrimaryUser.ID)
				Expect(err).NotTo(HaveOccurred())
			})

			Describe("With no songs", func() {
				It("returns success", func() {
					Expect(response.Code).To(Equal(http.StatusOK))
				})

				It("returns an empty array", func() {
					result := testing.DecodeJSON[[]any](response.Body)
					Expect(result).To(BeEmpty())
				})
			})

			Describe("With some songs", func() {
				var (
					songs   [3]map[string]any
					songIDs [3]string
				)

				var (
					makeExpectedSummary = func(song map[string]any) map[string]any {
						summary := map[string]any{}

						summary["id"] = song["id"]
						summary["owner"] = song["owner"]
						summary["lastSavedAt"] = song["lastSavedAt"]
						summary["metadata"] = song["metadata"]
						return summary
					}
				)

				BeforeEach(func() {
					songs = [3]map[string]any{}
					songIDs = [3]string{}

					songs[0] = testing.LoadDemoSong()
					Expect(songs[0]["elements"]).NotTo(BeNil())
					songIDs[0], songs[0] = createSong(songs[0])

					songs[1] = testing.LoadDemoSong()
					Expect(songs[1]["elements"]).NotTo(BeNil())
					song2Metadata := testing.ExpectType[map[string]any](songs[1]["metadata"])
					song2Metadata["title"] = "Ocean Wide Canyon Deep"
					song2Metadata["composedBy"] = "Jacob Collier"
					song2Metadata["performedBy"] = "Jacob Collier"
					songIDs[1], songs[1] = createSong(songs[1])

					songs[2] = testing.LoadDemoSong()
					Expect(songs[2]["elements"]).NotTo(BeNil())
					song3Metadata := testing.ExpectType[map[string]any](songs[2]["metadata"])
					song3Metadata["title"] = "苺"
					song3Metadata["composedBy"] = "荒谷翔大"
					song3Metadata["performedBy"] = "yonawo"
					songIDs[2], songs[2] = createSong(songs[2])
				})

				It("returns success", func() {
					Expect(response.Code).To(Equal(http.StatusOK))
				})

				It("doesn't return the body of the song", func() {
					songSummaries := testing.DecodeJSON[[]map[string]any](response.Body)
					for _, summary := range songSummaries {
						Expect(summary).NotTo(HaveKey("elements"))
					}
				})

				It("returns the other data of the song besides the body", func() {
					expectedSummaries := []any{}
					for _, song := range songs {
						expectedSummaries = append(expectedSummaries, makeExpectedSummary(song))
					}

					songSummaries := testing.DecodeJSON[[]map[string]any](response.Body)
					Expect(songSummaries).To(ConsistOf(expectedSummaries...))
				})
			})
		})
	})

	Describe("Create Song", func() {
		var (
			createSongPayload map[string]any
		)

		BeforeEach(func() {
			createSongPayload = testing.LoadDemoSong()
		})

		Describe("Unpermitted requests", func() {
			BeforeEach(func() {
				authtest.Endpoint = songGateway.CreateSong
				authtest.JSONBody = createSongPayload
			})

			authtest.ItRejectsUnpermittedRequests("POST", "/songs")
		})

		Describe("Authorized", func() {
			var (
				response       *httptest.ResponseRecorder
				requestFactory testing.RequestFactory
			)

			BeforeEach(func() {
				requestFactory = testing.RequestFactory{
					Method:  "POST",
					Target:  "/songs",
					JSONObj: createSongPayload,
				}
			})

			JustBeforeEach(func() {
				request := requestFactory.MakeFake()
				response = httptest.NewRecorder()
				c := testing.PrepareEchoContext(request, response)

				err := songGateway.CreateSong(c)
				Expect(err).NotTo(HaveOccurred())
			})

			Describe("For an authorized owner", func() {
				var (
					requestTime time.Time
				)

				BeforeEach(func() {
					requestFactory.Mods.Add(testing.WithUserCred(testing.PrimaryUser))
					requestTime = time.Now().UTC().Truncate(time.Second)
				})

				Describe("Song fields are accepted", func() {
					It("returns success", func() {
						Expect(response.Code).To(Equal(http.StatusOK))
					})

					It("returns a new song with an ID", func() {
						responseBody := testing.DecodeJSON[map[string]any](response.Body)
						Expect(responseBody["id"]).NotTo(BeEmpty())
					})

					It("returns an updated lastSavedAt", func() {
						responseBody := testing.DecodeJSON[map[string]any](response.Body)
						lastSavedAtStr := testing.ExpectType[string](responseBody["lastSavedAt"])
						lastSavedAt := testing.ExpectSuccess(time.Parse(time.RFC3339, lastSavedAtStr))
						Expect(lastSavedAt).To(BeTemporally(">=", requestTime))
						Expect(lastSavedAt).To(BeTemporally("<=", time.Now()))
					})

					It("returns the same song object", func() {
						responseBody := testing.DecodeJSON[map[string]any](response.Body)

						responseBody["id"] = ""
						testing.ExpectJSONEqualExceptLastSavedAt(responseBody, createSongPayload)
					})

					It("persists and can be retrieved after", func() {
						createResponseBody := testing.DecodeJSON[map[string]any](response.Body)
						songID := testing.ExpectType[string](createResponseBody["id"])

						getResponseBody := getSong(songID)
						Expect(getResponseBody).To(Equal(createResponseBody))
					})
				})

				Describe("For a song payload that already has an ID", func() {
					BeforeEach(func() {
						createSongPayload["id"] = uuid.New().String()
					})

					It("fails with the right error code", func() {
						resErr := testing.DecodeJSONError(response.Body)
						Expect(resErr.Code).To(BeEquivalentTo(songerrors.ExistingSongCode))
					})

					It("fails with the right status code", func() {
						Expect(response.Code).To(Equal(http.StatusBadRequest))
					})
				})

				Describe("For a song payload that doesn't have an owner field", func() {
					BeforeEach(func() {
						createSongPayload["owner"] = ""
					})

					It("fails with the right error code", func() {
						resErr := testing.DecodeJSONError(response.Body)
						Expect(resErr.Code).To(BeEquivalentTo(auth.WrongOwnerCode))
					})

					It("fails with the right status code", func() {
						Expect(response.Code).To(Equal(http.StatusForbidden))
					})
				})

				Describe("For a malformed song payload", func() {
					BeforeEach(func() {
						createSongPayload["id"] = 5
					})

					It("fails with the right error code", func() {
						resErr := testing.DecodeJSONError(response.Body)
						Expect(resErr.Code).To(BeEquivalentTo(songerrors.BadSongDataCode))
					})

					It("fails with the right status code", func() {
						Expect(response.Code).To(Equal(http.StatusBadRequest))
					})
				})
			})
		})
	})

	Describe("Update Song", func() {
		var (
			songID     string
			songUpdate map[string]any
		)

		var getFirstBlock = func(song map[string]any) map[string]any {
			lines := testing.ExpectType[[]any](song["elements"])
			Expect(lines).NotTo(BeEmpty())
			firstLine := testing.ExpectType[map[string]any](lines[0])
			firstLineElements := testing.ExpectType[[]any](firstLine["elements"])
			Expect(firstLineElements).NotTo(BeEmpty())
			firstBlock := testing.ExpectType[map[string]any](firstLineElements[0])
			return firstBlock
		}

		var getMetadata = func(song map[string]any) map[string]any {
			return testing.ExpectType[map[string]any](songUpdate["metadata"])
		}

		BeforeEach(func() {
			songID, _ = createSong(testing.LoadDemoSong())

			songUpdate = testing.LoadDemoSong()

			songUpdate["id"] = songID
			metadata := getMetadata(songUpdate)
			metadata["title"] = "Totally gonna give you up"

			firstBlock := getFirstBlock(songUpdate)
			Expect(firstBlock["chord"]).To(Equal("G^"))
			firstBlock["chord"] = "Dm7"
		})

		Describe("Unpermitted requests", func() {
			BeforeEach(func() {
				authtest.Endpoint = func(c echo.Context) error {
					return songGateway.UpdateSong(c, songID)
				}
				authtest.JSONBody = songUpdate
			})

			authtest.ItRejectsUnpermittedRequests("PUT", "/songs/:id")
		})

		Describe("Authorized", func() {
			var (
				response       *httptest.ResponseRecorder
				requestFactory testing.RequestFactory
			)

			BeforeEach(func() {
				requestFactory = testing.RequestFactory{
					Method:  "PUT",
					Target:  "/songs/:id",
					JSONObj: songUpdate,
				}
			})

			BeforeEach(func() {
				requestFactory.Mods.Add(testing.WithUserCred(testing.PrimaryUser))
			})

			JustBeforeEach(func() {
				request := requestFactory.MakeFake()
				response = httptest.NewRecorder()
				c := testing.PrepareEchoContext(request, response)

				err := songGateway.UpdateSong(c, songID)
				Expect(err).NotTo(HaveOccurred())
			})

			Describe("With an acceptable last saved at time", func() {
				var (
					previousLastSavedAt time.Time
				)

				BeforeEach(func() {
					previousLastSavedAt = time.Now().UTC().Truncate(time.Second)
					songUpdate["lastSavedAt"] = previousLastSavedAt.Format(time.RFC3339)
				})

				Describe("With mostly normal stuff", func() {
					It("succeeds", func() {
						Expect(response.Code).To(Equal(http.StatusOK))
					})

					It("updated the last saved at time", func() {
						updatedSong := testing.DecodeJSON[map[string]any](response.Body)
						updatedTimeStr := testing.ExpectType[string](updatedSong["lastSavedAt"])
						Expect(updatedTimeStr).NotTo(BeZero())
						updatedTime := testing.ExpectSuccess(time.Parse(time.RFC3339, updatedTimeStr))
						Expect(updatedTime).To(BeTemporally(">=", previousLastSavedAt))
						Expect(updatedTime).To(BeTemporally("<=", time.Now()))
					})

					It("returns the updated song", func() {
						updatedSong := testing.DecodeJSON[map[string]any](response.Body)
						testing.ExpectJSONEqualExceptLastSavedAt(updatedSong, songUpdate)
					})

					It("is updated and can be fetched", func() {
						fetchedSong := getSong(songID)
						updatedSong := testing.DecodeJSON[map[string]any](response.Body)
						Expect(fetchedSong).To(Equal(updatedSong))
					})
				})

				Describe("With a tampered ID", func() {
					BeforeEach(func() {
						randomID := uuid.New().String()
						Expect(randomID).NotTo(Equal(songID))
						songUpdate["id"] = randomID
					})

					It("succeeds", func() {
						Expect(response.Code).To(Equal(http.StatusOK))
					})

					It("doesn't overwrite the ID", func() {
						updatedSong := testing.DecodeJSON[map[string]any](response.Body)
						Expect(updatedSong["id"]).To(Equal(songID))
					})
				})

				Describe("With a tampered owner", func() {
					BeforeEach(func() {
						songUpdate["owner"] = testing.OtherUser.ID
					})

					It("succeeds", func() {
						Expect(response.Code).To(Equal(http.StatusOK))
					})

					It("doesn't overwrite the owner", func() {
						updatedSong := testing.DecodeJSON[map[string]any](response.Body)
						Expect(updatedSong["owner"]).To(Equal(testing.PrimaryUser.ID))
					})
				})
			})

			Describe("For a song with no last saved at timestamp", func() {
				BeforeEach(func() {
					delete(songUpdate, "lastSavedAt")
				})

				It("rejects the save with the right error code", func() {
					resErr := testing.DecodeJSONError(response.Body)
					Expect(resErr.Code).To(BeEquivalentTo(songerrors.SongOverwriteCode))
				})

				It("rejects the save with the right status code", func() {
					Expect(response.Code).To(Equal(http.StatusBadRequest))
				})
			})

			Describe("For a song that doesn't exist", func() {
				BeforeEach(func() {
					songID = uuid.New().String()
					songUpdate["id"] = songID
				})

				It("fails with the right error code", func() {
					resErr := testing.DecodeJSONError(response.Body)
					Expect(resErr.Code).To(BeEquivalentTo(songerrors.SongNotFoundCode))
				})

				It("fails with the right status code", func() {
					Expect(response.Code).To(Equal(http.StatusNotFound))
				})
			})

			Describe("For an empty ID", func() {
				BeforeEach(func() {
					songID = ""
					songUpdate["id"] = ""
				})

				It("fails with the right error code", func() {
					resErr := testing.DecodeJSONError(response.Body)
					Expect(resErr.Code).To(BeEquivalentTo(songerrors.SongNotFoundCode))
				})

				It("fails with the right status code", func() {
					Expect(response.Code).To(Equal(http.StatusNotFound))
				})
			})

			Describe("For a song that has an old last saved at timestamp", func() {
				BeforeEach(func() {
					anHourAgo := time.Now().UTC().Truncate(time.Second).Add(-time.Hour)
					songUpdate["lastSavedAt"] = anHourAgo.Format(time.RFC3339)
				})

				It("rejects the save with the right error code", func() {
					resErr := testing.DecodeJSONError(response.Body)
					Expect(resErr.Code).To(BeEquivalentTo(songerrors.SongOverwriteCode))
				})

				It("rejects the save with the right status code", func() {
					Expect(response.Code).To(Equal(http.StatusBadRequest))
				})
			})
		})
	})

	Describe("Delete Song", func() {
		var (
			songID string
		)

		BeforeEach(func() {
			createSongPayload := testing.LoadDemoSong()
			songID, _ = createSong(createSongPayload)
		})

		Describe("Unpermitted", func() {
			BeforeEach(func() {
				authtest.Endpoint = func(c echo.Context) error {
					Expect(songID).NotTo(BeEmpty())
					return songGateway.DeleteSong(c, songID)
				}
			})

			authtest.ItRejectsUnpermittedRequests("DELETE", "/songs/:id")
		})

		Describe("Authorized", func() {
			var (
				response       *httptest.ResponseRecorder
				requestFactory testing.RequestFactory
			)

			BeforeEach(func() {
				requestFactory = testing.RequestFactory{
					Method:  "DELETE",
					Target:  "/songs/:id",
					JSONObj: nil,
				}
			})

			BeforeEach(func() {
				requestFactory.Mods.Add(testing.WithUserCred(testing.PrimaryUser))
			})

			JustBeforeEach(func() {
				request := requestFactory.MakeFake()
				response = httptest.NewRecorder()
				c := testing.PrepareEchoContext(request, response)

				err := songGateway.DeleteSong(c, songID)
				Expect(err).NotTo(HaveOccurred())
			})

			Describe("For a nonexisting song", func() {
				Describe("For a valid ID", func() {
					BeforeEach(func() {
						songID = uuid.New().String()
					})

					It("fails with the right error code", func() {
						resErr := testing.DecodeJSONError(response.Body)
						Expect(resErr.Code).To(BeEquivalentTo(songerrors.SongNotFoundCode))
					})

					It("fails with the right status code", func() {
						Expect(response.Code).To(Equal(http.StatusNotFound))
					})
				})

				Describe("For an invalid ID", func() {
					BeforeEach(func() {
						songID = "boat"
					})

					It("fails with the right error code", func() {
						resErr := testing.DecodeJSONError(response.Body)
						Expect(resErr.Code).To(BeEquivalentTo(songerrors.SongNotFoundCode))
					})

					It("fails with the right status code", func() {
						Expect(response.Code).To(Equal(http.StatusNotFound))
					})
				})

				Describe("For an empty ID", func() {
					BeforeEach(func() {
						songID = ""
					})

					It("fails with the right error code", func() {
						resErr := testing.DecodeJSONError(response.Body)
						Expect(resErr.Code).To(BeEquivalentTo(songerrors.SongNotFoundCode))
					})

					It("fails with the right status code", func() {
						Expect(response.Code).To(Equal(http.StatusNotFound))
					})
				})

			})

			Describe("For an existing song", func() {
				It("returns success", func() {
					Expect(response.Code).To(Equal(http.StatusOK))
				})

				It("returns no content", func() {
					Expect(response.Body.Len()).To(BeZero())
				})

				It("removes the song and can't be retrieved after", func() {
					getResponse := httptest.NewRecorder()

					By("Making a request to Get Song", func() {
						getRequest := testing.RequestFactory{
							Method:  "GET",
							Target:  fmt.Sprintf("/songs/%s", songID),
							JSONObj: nil,
						}.MakeFake()

						c := testing.PrepareEchoContext(getRequest, getResponse)
						err := songGateway.GetSong(c, songID)
						Expect(err).NotTo(HaveOccurred())
					})

					By("Inspecting the error from Get Song", func() {
						Expect(getResponse.Code).To(Equal(http.StatusNotFound))
						getResponseError := testing.DecodeJSON[api_error.JSONAPIError](getResponse.Body)
						Expect(getResponseError.Code).To(BeEquivalentTo(songerrors.SongNotFoundCode))
					})
				})
			})
		})
	})
})
