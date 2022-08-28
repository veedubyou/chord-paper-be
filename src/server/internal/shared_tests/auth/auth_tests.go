package authtest

import (
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/veedubyou/chord-paper-be/src/server/internal/errors/auth"
	"github.com/veedubyou/chord-paper-be/src/shared/testing"
	"net/http"
	"net/http/httptest"
)

// to use this shared test, all tests must set the Endpoint in the BeforeEach
// and JSONBody optionally
var (
	Endpoint func(c echo.Context) error
	JSONBody any
)

func ItRejectsUnpermittedRequests(method string, path string) {
	ItRejectsUnauthorizedRequests(method, path)
	ItRejectsWrongOwnerRequests(method, path)
}

func ItRejectsWrongOwnerRequests(method string, path string) {
	Describe("Unauthenticated requests", func() {
		var (
			response       *httptest.ResponseRecorder
			requestFactory testing.RequestFactory
		)

		BeforeEach(func() {
			requestFactory = testing.RequestFactory{
				Method:  method,
				Target:  path,
				JSONObj: JSONBody,
			}
		})

		BeforeEach(func() {
			Expect(Endpoint).NotTo(BeNil())
		})

		AfterEach(func() {
			Endpoint = nil
			JSONBody = nil
		})

		JustBeforeEach(func() {
			request := requestFactory.MakeFake()
			response = httptest.NewRecorder()
			c := testing.PrepareEchoContext(request, response)

			Expect(Endpoint).NotTo(BeNil())
			err := Endpoint(c)
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("For a user that's not the owner of the resources", func() {
			BeforeEach(func() {
				requestFactory.Mods.Add(testing.WithUserCred(testing.OtherUser))
			})

			It("fails with the right error code", func() {
				resErr := testing.DecodeJSONError(response.Body)
				Expect(resErr.Code).To(BeEquivalentTo(auth.WrongOwnerCode))
			})

			It("fails with the right status code", func() {
				Expect(response.Code).To(Equal(http.StatusForbidden))
			})
		})
	})
}

func ItRejectsUnauthorizedRequests(method string, path string) {
	Describe("Unauthorized requests", func() {
		var (
			response       *httptest.ResponseRecorder
			requestFactory testing.RequestFactory
		)

		BeforeEach(func() {
			requestFactory = testing.RequestFactory{
				Method:  method,
				Target:  path,
				JSONObj: JSONBody,
			}
		})

		BeforeEach(func() {
			Expect(Endpoint).NotTo(BeNil())
		})

		AfterEach(func() {
			Endpoint = nil
			JSONBody = nil
		})

		JustBeforeEach(func() {
			request := requestFactory.MakeFake()
			response = httptest.NewRecorder()
			c := testing.PrepareEchoContext(request, response)

			Expect(Endpoint).NotTo(BeNil())
			err := Endpoint(c)
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("With no auth header", func() {
			It("fails with the right error code", func() {
				resErr := testing.DecodeJSONError(response.Body)
				Expect(resErr.Code).To(BeEquivalentTo(auth.BadAuthorizationHeaderCode))
			})

			It("fails with the right status code", func() {
				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Describe("With a malformed token", func() {
			BeforeEach(func() {
				token := testing.TokenForUserID(testing.PrimaryUser.ID)
				requestFactory.Mods.Add(testing.WithAuthHeader(token))
			})

			It("fails with the right error code", func() {
				resErr := testing.DecodeJSONError(response.Body)
				Expect(resErr.Code).To(BeEquivalentTo(auth.BadAuthorizationHeaderCode))
			})

			It("fails with the right status code", func() {
				Expect(response.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Describe("With a Google unauthorized token", func() {
			BeforeEach(func() {
				requestFactory.Mods.Add(testing.WithUserCred(testing.GoogleUnauthorizedUser))
			})

			It("fails with the right error code", func() {
				resErr := testing.DecodeJSONError(response.Body)
				Expect(resErr.Code).To(BeEquivalentTo(auth.NotGoogleAuthorizedCode))
			})

			It("fails with the right status code", func() {
				Expect(response.Code).To(Equal(http.StatusUnauthorized))
			})
		})

		Describe("For a user that's not in the DB", func() {
			BeforeEach(func() {
				requestFactory.Mods.Add(testing.WithUserCred(testing.NoAccountUser))
			})

			It("fails with the right error code", func() {
				resErr := testing.DecodeJSONError(response.Body)
				Expect(resErr.Code).To(BeEquivalentTo(auth.NoAccountCode))
			})

			It("fails with the right status code", func() {
				Expect(response.Code).To(Equal(http.StatusUnauthorized))
			})
		})
	})
}
