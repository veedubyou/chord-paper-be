package user_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/veedubyou/chord-paper-be/src/server/internal/shared_tests/auth"
	"github.com/veedubyou/chord-paper-be/src/server/internal/user/gateway"
	"github.com/veedubyou/chord-paper-be/src/server/internal/user/storage"
	"github.com/veedubyou/chord-paper-be/src/server/internal/user/usecase"
	"github.com/veedubyou/chord-paper-be/src/shared/testing"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("User", func() {
	var (
		userGateway usergateway.Gateway
		validator   testing.Validator
	)

	BeforeEach(func() {
		validator = testing.Validator{}
		userStorage := userstorage.NewDB(db)
		userUsecase := userusecase.NewUsecase(userStorage, validator)
		userGateway = usergateway.NewGateway(userUsecase)
	})

	BeforeEach(func() {
		testing.ResetDB(db)
	})

	Describe("Login", func() {
		Describe("Unauthorized", func() {
			BeforeEach(func() {
				authtest.Endpoint = userGateway.Login
			})

			authtest.ItRejectsUnauthorizedRequests("POST", "/login")
		})

		Describe("For an authorized user", func() {
			var (
				response *httptest.ResponseRecorder
			)

			BeforeEach(func() {
				request := testing.RequestFactory{
					Method:  "POST",
					Target:  "/login",
					JSONObj: nil,
					Mods:    testing.RequestModifiers{testing.WithUserCred(testing.PrimaryUser)},
				}.MakeFake()
				response = httptest.NewRecorder()

				c := testing.PrepareEchoContext(request, response)
				err := userGateway.Login(c)
				Expect(err).NotTo(HaveOccurred())
			})

			It("succeeds", func() {
				Expect(response.Code).To(Equal(http.StatusOK))
			})

			It("returns the correct user", func() {
				userResponse := testing.DecodeJSON[usergateway.UserJSON](response.Body)
				Expect(userResponse).To(BeEquivalentTo(testing.PrimaryUser))
			})
		})
	})
})
