package api

import (
	"bytes"
	"encoding/json"
	"errors"
	mockstore "github.com/HyperGAI/serving-webhook/storage/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func mustOpen(filename string) *os.File {
	if f, err := os.Open(filename); err != nil {
		panic("not reached")
	} else {
		return f
	}
}

func buildRequestBody(values map[string]io.Reader) (*bytes.Buffer, string, error) {
	var b bytes.Buffer
	var err error

	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		if x, ok := r.(*os.File); ok {
			// Add a file
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return nil, "", err
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return nil, "", err
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return nil, "", err
		}
	}
	w.Close()
	return &b, w.FormDataContentType(), nil
}

func TestUpload(t *testing.T) {
	testCases := []struct {
		name          string
		body          map[string]io.Reader
		buildStubs    func(store *mockstore.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: map[string]io.Reader{
				"file": mustOpen("file_test.go"),
			},
			buildStubs: func(store *mockstore.MockStore) {
				/*
					store.EXPECT().
						Upload(gomock.Any(), gomock.Any()).
						Times(1).
						Return("test_url", nil)
				*/
				store.EXPECT().
					PutObject(gomock.Any(), gomock.Any()).
					Times(1).
					Return("test_url", nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				var res map[string]string
				data, err := io.ReadAll(recorder.Body)
				require.NoError(t, err)
				err = json.Unmarshal(data, &res)
				require.NoError(t, err)
				require.Equal(t, res["url"], "test_url")
			},
		},
		{
			name: "Upload failed",
			body: map[string]io.Reader{
				"file": mustOpen("file_test.go"),
			},
			buildStubs: func(store *mockstore.MockStore) {
				/*
					store.EXPECT().
						Upload(gomock.Any(), gomock.Any()).
						Times(1).
						Return("", errors.New("upload failed"))
				*/
				store.EXPECT().
					PutObject(gomock.Any(), gomock.Any()).
					Times(1).
					Return("", errors.New("upload failed"))
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "File field not exist",
			body: map[string]io.Reader{
				"other": strings.NewReader("test"),
			},
			buildStubs: func(store *mockstore.MockStore) {
				store.EXPECT().
					Upload(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					PutObject(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockstore.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, nil, nil)
			recorder := httptest.NewRecorder()

			requestBody, contentType, err := buildRequestBody(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(
				http.MethodPost, "/upload", requestBody)
			require.NoError(t, err)
			request.Header.Set("Content-Type", contentType)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}
