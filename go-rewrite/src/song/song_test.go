package song_test

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/errors/auth"
	"github.com/veedubyou/chord-paper-be/go-rewrite/src/errors/gateway"
	. "github.com/veedubyou/chord-paper-be/go-rewrite/src/lib/testing"
	authtest "github.com/veedubyou/chord-paper-be/go-rewrite/src/shared_tests/auth"
	songerrors "github.com/veedubyou/chord-paper-be/go-rewrite/src/song/errors"
	songgateway "github.com/veedubyou/chord-paper-be/go-rewrite/src/song/gateway"
	songstorage "github.com/veedubyou/chord-paper-be/go-rewrite/src/song/storage"
	songusecase "github.com/veedubyou/chord-paper-be/go-rewrite/src/song/usecase"
	userstorage "github.com/veedubyou/chord-paper-be/go-rewrite/src/user/storage"
	userusecase "github.com/veedubyou/chord-paper-be/go-rewrite/src/user/usecase"
	"net/http"
	"net/http/httptest"
	"os"
)

func loadSong() map[string]interface{} {
	file := ExpectSuccess(os.Open("./demo_song_test.json"))

	output := map[string]interface{}{}
	err := json.NewDecoder(file).Decode(&output)
	Expect(err).NotTo(HaveOccurred())

	output["id"] = ""
	output["owner"] = PrimaryUser.ID

	return output
}

var _ = Describe("Song", func() {
	var (
		songGateway songgateway.Gateway
		validator   TestingValidator
	)

	BeforeEach(func() {
		validator = TestingValidator{}
		userStorage := userstorage.NewDB(db)
		userUsecase := userusecase.NewUsecase(userStorage, validator)

		songStorage := songstorage.NewDB(db)
		songUsecase := songusecase.NewUsecase(songStorage, userUsecase)
		songGateway = songgateway.NewGateway(songUsecase)
	})

	BeforeEach(func() {
		ResetDB(db)
	})

	Describe("Get Song", func() {
		Describe("For non-existing songs", func() {
			var (
				response *httptest.ResponseRecorder
				songID   string
			)

			JustBeforeEach(func() {
				requestFactory := RequestFactory{
					Method:  "GET",
					Path:    "/songs/:id",
					JSONObj: nil,
				}

				request := requestFactory.Make()
				response = httptest.NewRecorder()
				c := PrepareEchoContext(request, response)

				Expect(songID).NotTo(BeEmpty())
				err := songGateway.GetSong(c, songID)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				songID = ""
			})

			Describe("For a malformed ID", func() {
				BeforeEach(func() {
					songID = "boat"
				})

				It("fails with the right error code", func() {
					resErr := DecodeJSONError(response)
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
					resErr := DecodeJSONError(response)
					Expect(resErr.Code).To(BeEquivalentTo(songerrors.SongNotFoundCode))
				})

				It("fails with the right status code", func() {
					Expect(response.Code).To(Equal(http.StatusNotFound))
				})
			})
		})
	})

	Describe("Create Song", func() {
		var (
			createSongPayload map[string]interface{}
		)

		BeforeEach(func() {
			createSongPayload = loadSong()
		})

		Describe("Shared auth tests", func() {
			BeforeEach(func() {
				authtest.Endpoint = songGateway.CreateSong
				authtest.JSONBody = createSongPayload
			})

			authtest.ItRejectsUnauthorizedRequests("POST", "/songs")
		})

		Describe("Unshared tests", func() {
			var (
				response       *httptest.ResponseRecorder
				requestFactory RequestFactory
			)

			BeforeEach(func() {
				requestFactory = RequestFactory{
					Method:  "POST",
					Path:    "/songs",
					JSONObj: createSongPayload,
				}
			})

			JustBeforeEach(func() {
				request := requestFactory.Make()
				response = httptest.NewRecorder()
				c := PrepareEchoContext(request, response)

				err := songGateway.CreateSong(c)
				Expect(err).NotTo(HaveOccurred())
			})

			Describe("For an authorized owner", func() {
				BeforeEach(func() {
					requestFactory.Mods.Add(WithUserCred(PrimaryUser))
				})

				Describe("Song fields are accepted", func() {
					It("returns success", func() {
						Expect(response.Code).To(Equal(http.StatusOK))
					})

					It("returns a new song with an ID", func() {
						responseBody := DecodeJSON[map[string]interface{}](response)
						Expect(responseBody["id"]).NotTo(BeEmpty())
					})

					It("returns the same song object", func() {
						responseBody := DecodeJSON[map[string]interface{}](response)

						responseBody["id"] = ""
						ExpectEqualSongJSON(responseBody, createSongPayload)
					})

					It("persists and can be retrieved after", func() {
						By("Decoding the create song response")
						createResponseBody := DecodeJSON[map[string]interface{}](response)
						songID, ok := createResponseBody["id"].(string)
						Expect(ok).To(BeTrue())

						By("Making a request to Get Song")
						getRequest := RequestFactory{
							Method:  "GET",
							Path:    fmt.Sprintf("/songs/%s", songID),
							JSONObj: nil,
						}.Make()
						getResponse := httptest.NewRecorder()
						c := PrepareEchoContext(getRequest, getResponse)
						err := songGateway.GetSong(c, songID)
						Expect(err).NotTo(HaveOccurred())

						By("Comparing the response from Get Song")
						getResponseBody := DecodeJSON[map[string]interface{}](getResponse)
						Expect(getResponseBody).To(Equal(createResponseBody))
					})
				})

				Describe("For a song payload that already has an ID", func() {
					BeforeEach(func() {
						createSongPayload["id"] = uuid.New().String()
					})

					It("fails with the right error code", func() {
						resErr := DecodeJSONError(response)
						Expect(resErr.Code).To(BeEquivalentTo(songerrors.ExistingSongCode))
					})

					It("fails with the right status code", func() {
						Expect(response.Code).To(Equal(http.StatusUnprocessableEntity))
					})
				})

				Describe("For a song payload that doesn't have an owner field", func() {
					BeforeEach(func() {
						createSongPayload["owner"] = ""
					})

					It("fails with the right error code", func() {
						resErr := DecodeJSONError(response)
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
						resErr := DecodeJSONError(response)
						Expect(resErr.Code).To(BeEquivalentTo(songerrors.BadSongDataCode))
					})

					It("fails with the right status code", func() {
						Expect(response.Code).To(Equal(http.StatusBadRequest))
					})
				})
			})

			Describe("For an unauthorized owner", func() {
				BeforeEach(func() {
					requestFactory.Mods.Add(WithUserCred(OtherUser))
				})

				It("fails with the right error code", func() {
					resErr := DecodeJSONError(response)
					Expect(resErr.Code).To(BeEquivalentTo(auth.WrongOwnerCode))
				})

				It("fails with the right status code", func() {
					Expect(response.Code).To(Equal(http.StatusForbidden))
				})
			})
		})
	})

	Describe("Delete Song", func() {
		var (
			songID string
		)

		BeforeEach(func() {
			By("First creating a song")
			createSongPayload := loadSong()

			request := RequestFactory{
				Method:  "POST",
				Path:    "/songs",
				JSONObj: createSongPayload,
				Mods:    RequestModifiers{WithUserCred(PrimaryUser)},
			}.Make()
			response := httptest.NewRecorder()
			c := PrepareEchoContext(request, response)

			err := songGateway.CreateSong(c)
			Expect(err).NotTo(HaveOccurred())

			By("Extracting the ID from the created song")
			song := DecodeJSON[map[string]interface{}](response)

			var ok bool
			songID, ok = song["id"].(string)
			Expect(ok).To(BeTrue())
			Expect(songID).NotTo(BeEmpty())
		})

		Describe("Shared auth tests", func() {
			BeforeEach(func() {
				authtest.Endpoint = func(c echo.Context) error {
					Expect(songID).NotTo(BeEmpty())
					return songGateway.DeleteSong(c, songID)
				}
			})

			authtest.ItRejectsUnauthorizedRequests("DELETE", "/songs/:id")
		})

		Describe("Unshared tests", func() {
			var (
				response       *httptest.ResponseRecorder
				requestFactory RequestFactory
			)

			BeforeEach(func() {
				requestFactory = RequestFactory{
					Method:  "DELETE",
					Path:    "/songs/:id",
					JSONObj: nil,
				}
			})

			JustBeforeEach(func() {
				request := requestFactory.Make()
				response = httptest.NewRecorder()
				c := PrepareEchoContext(request, response)

				err := songGateway.DeleteSong(c, songID)
				Expect(err).NotTo(HaveOccurred())
			})

			Describe("For a nonexisting song", func() {
				BeforeEach(func() {
					requestFactory.Mods.Add(WithUserCred(PrimaryUser))
				})

				Describe("For a valid ID", func() {
					BeforeEach(func() {
						songID = uuid.New().String()
					})

					It("fails with the right error code", func() {
						resErr := DecodeJSONError(response)
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
						resErr := DecodeJSONError(response)
						Expect(resErr.Code).To(BeEquivalentTo(songerrors.SongNotFoundCode))
					})

					It("fails with the right status code", func() {
						Expect(response.Code).To(Equal(http.StatusNotFound))
					})
				})
			})

			Describe("For an authorized owner", func() {
				BeforeEach(func() {
					requestFactory.Mods.Add(WithUserCred(PrimaryUser))
				})

				It("returns success", func() {
					Expect(response.Code).To(Equal(http.StatusOK))
				})

				It("returns no content", func() {
					Expect(response.Body.Len()).To(BeZero())
				})

				It("removes the song and can't be retrieved after", func() {
					By("Making a request to Get Song")
					getRequest := RequestFactory{
						Method:  "GET",
						Path:    fmt.Sprintf("/songs/%s", songID),
						JSONObj: nil,
					}.Make()
					getResponse := httptest.NewRecorder()
					c := PrepareEchoContext(getRequest, getResponse)
					err := songGateway.GetSong(c, songID)
					Expect(err).NotTo(HaveOccurred())

					By("Inspecting the error from Get Song")
					Expect(getResponse.Code).To(Equal(http.StatusNotFound))
					getResponseError := DecodeJSON[gateway.JSONAPIError](getResponse)
					Expect(getResponseError.Code).To(BeEquivalentTo(songerrors.SongNotFoundCode))
				})
			})

			Describe("For an unauthorized owner", func() {
				BeforeEach(func() {
					requestFactory.Mods.Add(WithUserCred(OtherUser))
				})

				It("fails with the right error code", func() {
					resErr := DecodeJSONError(response)
					Expect(resErr.Code).To(BeEquivalentTo(auth.WrongOwnerCode))
				})

				It("fails with the right status code", func() {
					Expect(response.Code).To(Equal(http.StatusForbidden))
				})
			})
		})

	})
})
