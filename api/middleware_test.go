package api

import (
	"github.com/HyperGAI/serving-webhook/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func addAuthorization(
	request *http.Request,
	token string,
) {
	request.Header.Set(authorizationHeaderKey, token)
}

func TestAuthMiddleware(t *testing.T) {
	config := utils.Config{
		SecretAPIKey: "123456",
	}

	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(request, "123456")
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "No authorization",
			setupAuth: func(t *testing.T, request *http.Request) {
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "No permission",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(request, "000000")
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "Invalid authorization format",
			setupAuth: func(t *testing.T, request *http.Request) {
				addAuthorization(request, "000000 000")
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			server := newTestServer(t, nil, nil, nil)
			authPath := "/auth"
			server.router.GET(
				authPath,
				authMiddleware(config),
				func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{})
				},
			)

			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}
