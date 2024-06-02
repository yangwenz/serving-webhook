package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	mockdb "github.com/HyperGAI/serving-webhook/db/mock"
	db "github.com/HyperGAI/serving-webhook/db/sqlc"
	mockstore "github.com/HyperGAI/serving-webhook/storage/mock"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreate(t *testing.T) {
	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(cache *mockstore.MockCache)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"id":            "12345",
				"model_name":    "test_model",
				"model_version": "v1",
			},
			buildStubs: func(cache *mockstore.MockCache) {
				cache.EXPECT().
					SetKey(gomock.Eq("12345"), gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "OK",
			body: gin.H{
				"id":            "12345",
				"model_name":    "test_model",
				"model_version": "v1",
				"status":        "running",
			},
			buildStubs: func(cache *mockstore.MockCache) {
				cache.EXPECT().
					SetKey(gomock.Eq("12345"), gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Redis failed",
			body: gin.H{
				"id":            "12345",
				"model_name":    "test_model",
				"model_version": "v1",
			},
			buildStubs: func(cache *mockstore.MockCache) {
				cache.EXPECT().
					SetKey(gomock.Eq("12345"), gomock.Any(), gomock.Any()).
					Times(1).
					Return(errors.New("redis error"))
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cache := mockstore.NewMockCache(ctrl)
			tc.buildStubs(cache)

			server := newTestServer(t, nil, cache, nil)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(
				http.MethodPost, "/task", bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func requireBodyMatchTask(t *testing.T, body *bytes.Buffer, task TaskInfo) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotTask TaskInfo
	err = json.Unmarshal(data, &gotTask)

	require.NoError(t, err)
	require.Equal(t, gotTask.ID, task.ID)
	require.Equal(t, gotTask.ModelName, task.ModelName)
	require.Equal(t, gotTask.ModelVersion, task.ModelVersion)
	require.Equal(t, gotTask.Status, task.Status)
	require.Equal(t, gotTask.RunningTime, task.RunningTime)
	require.Equal(t, gotTask.CreatedAt, task.CreatedAt)
	require.Equal(t, gotTask.ErrorInfo, task.ErrorInfo)
	require.Equal(t, gotTask.QueueNum, task.QueueNum)
	require.Equal(t, gotTask.QueueID, task.QueueID)

	a, _ := json.Marshal(gotTask.Outputs)
	b, _ := json.Marshal(task.Outputs)
	require.Equal(t, string(a), string(b))
}

func TestGet(t *testing.T) {
	arg := TaskInfo{
		ID:           "12345",
		ModelName:    "test_model",
		ModelVersion: "v1",
		Status:       "pending",
		RunningTime:  "1s",
		CreatedAt:    time.Time{},
		Outputs:      nil,
	}

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(cache *mockstore.MockCache)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"id": "12345",
			},
			buildStubs: func(cache *mockstore.MockCache) {
				data, _ := json.Marshal(arg)
				cache.EXPECT().
					GetKey(gomock.Eq("12345")).
					Times(1).
					Return(string(data), nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTask(t, recorder.Body, arg)
			},
		},
		{
			name: "Task not found",
			body: gin.H{
				"id": "12345",
			},
			buildStubs: func(cache *mockstore.MockCache) {
				cache.EXPECT().
					GetKey(gomock.Eq("12345")).
					Times(1).
					Return("", errors.New("not found"))
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

			cache := mockstore.NewMockCache(ctrl)
			tc.buildStubs(cache)

			server := newTestServer(t, nil, cache, nil)
			recorder := httptest.NewRecorder()

			request, err := http.NewRequest(
				http.MethodGet, fmt.Sprintf("/task/%s", tc.body["id"]), nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestUpdate(t *testing.T) {
	arg := TaskInfo{
		ID:           "12345",
		ModelName:    "test_model",
		ModelVersion: "v1",
		Status:       "pending",
		RunningTime:  "",
		CreatedAt:    time.Time{},
		Outputs:      nil,
		ErrorInfo:    "",
		QueueNum:     1,
	}
	output := TaskInfo{
		ID:           "12345",
		ModelName:    "test_model",
		ModelVersion: "v1",
		Status:       "succeeded",
		RunningTime:  "5s",
		CreatedAt:    time.Time{},
		Outputs:      map[string]string{"output": "abc"},
		ErrorInfo:    "empty",
		QueueNum:     1,
		QueueID:      "1234",
	}

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(cache *mockstore.MockCache)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"id":           "12345",
				"status":       "succeeded",
				"running_time": "5s",
				"outputs":      map[string]string{"output": "abc"},
				"error_info":   "empty",
				"queue_id":     "1234",
			},
			buildStubs: func(cache *mockstore.MockCache) {
				data, _ := json.Marshal(arg)
				cache.EXPECT().
					GetKey(gomock.Eq("12345")).
					Times(1).
					Return(string(data), nil)
				cache.EXPECT().
					SetKey(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTask(t, recorder.Body, output)
			},
		},
		{
			name: "Task not found",
			body: gin.H{
				"id":           "12345",
				"status":       "failed",
				"running_time": "5s",
				"outputs":      map[string]string{"output": "abc"},
			},
			buildStubs: func(cache *mockstore.MockCache) {
				cache.EXPECT().
					GetKey(gomock.Eq("12345")).
					Times(1).
					Return("", errors.New("not found"))
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Unmarshal error",
			body: gin.H{
				"id":           "12345",
				"status":       "failed",
				"running_time": "5s",
				"outputs":      map[string]string{"output": "abc"},
			},
			buildStubs: func(cache *mockstore.MockCache) {
				cache.EXPECT().
					GetKey(gomock.Eq("12345")).
					Times(1).
					Return("abc", nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cache := mockstore.NewMockCache(ctrl)
			tc.buildStubs(cache)

			server := newTestServer(t, nil, cache, nil)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(
				http.MethodPut, "/task", bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestCreateWithDB(t *testing.T) {
	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(cache *mockstore.MockCache, database *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"id":            "12345",
				"model_name":    "test_model",
				"model_version": "v1",
			},
			buildStubs: func(cache *mockstore.MockCache, database *mockdb.MockStore) {
				cache.EXPECT().
					SetKey(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				database.EXPECT().ExecTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Creation failed",
			body: gin.H{
				"id":            "12345",
				"model_name":    "test_model",
				"model_version": "v1",
			},
			buildStubs: func(cache *mockstore.MockCache, database *mockdb.MockStore) {
				cache.EXPECT().
					SetKey(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				database.EXPECT().ExecTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(errors.New("creation error"))
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cache := mockstore.NewMockCache(ctrl)
			database := mockdb.NewMockStore(ctrl)
			tc.buildStubs(cache, database)

			server := newTestServer(t, nil, cache, database)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(
				http.MethodPost, "/task", bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateWithDB(t *testing.T) {
	arg := TaskInfo{
		ID:           "12345",
		ModelName:    "test_model",
		ModelVersion: "v1",
		Status:       "pending",
		RunningTime:  "",
		CreatedAt:    time.Time{},
		Outputs:      nil,
		ErrorInfo:    "",
		QueueNum:     1,
	}
	output := TaskInfo{
		ID:           "12345",
		ModelName:    "test_model",
		ModelVersion: "v1",
		Status:       "succeeded",
		RunningTime:  "5s",
		CreatedAt:    time.Time{},
		Outputs:      map[string]string{"output": "abc"},
		ErrorInfo:    "empty",
		QueueNum:     1,
		QueueID:      "1234",
	}

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(cache *mockstore.MockCache, database *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"id":           "12345",
				"status":       "succeeded",
				"running_time": "5s",
				"outputs":      map[string]string{"output": "abc"},
				"error_info":   "empty",
				"queue_id":     "1234",
			},
			buildStubs: func(cache *mockstore.MockCache, database *mockdb.MockStore) {
				data, _ := json.Marshal(arg)
				cache.EXPECT().
					GetKey(gomock.Eq("12345")).
					Times(1).
					Return(string(data), nil)
				cache.EXPECT().
					SetKey(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				database.EXPECT().ExecTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTask(t, recorder.Body, output)
			},
		},
		{
			name: "OK DatabaseOnly",
			body: gin.H{
				"id":            "12345",
				"status":        "succeeded",
				"running_time":  "5s",
				"outputs":       map[string]string{"output": "abc"},
				"error_info":    "empty",
				"queue_id":      "1234",
				"database_only": true,
			},
			buildStubs: func(cache *mockstore.MockCache, database *mockdb.MockStore) {
				cache.EXPECT().
					GetKey(gomock.Eq("12345")).
					Times(0)
				cache.EXPECT().
					SetKey(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				database.EXPECT().ExecTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Task not found",
			body: gin.H{
				"id":           "12345",
				"status":       "failed",
				"running_time": "5s",
				"outputs":      map[string]string{"output": "abc"},
			},
			buildStubs: func(cache *mockstore.MockCache, database *mockdb.MockStore) {
				cache.EXPECT().
					GetKey(gomock.Eq("12345")).
					Times(1).
					Return("", errors.New("not found"))
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Unmarshal error",
			body: gin.H{
				"id":           "12345",
				"status":       "failed",
				"running_time": "5s",
				"outputs":      map[string]string{"output": "abc"},
			},
			buildStubs: func(cache *mockstore.MockCache, database *mockdb.MockStore) {
				cache.EXPECT().
					GetKey(gomock.Eq("12345")).
					Times(1).
					Return("abc", nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cache := mockstore.NewMockCache(ctrl)
			database := mockdb.NewMockStore(ctrl)
			tc.buildStubs(cache, database)

			server := newTestServer(t, nil, cache, database)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(
				http.MethodPut, "/task", bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestGetTaskFromDB(t *testing.T) {
	tasks := []db.Task{
		{
			ID:        0,
			TaskID:    "12345",
			ModelName: "test",
			Status:    pgtype.Text{String: "pending", Valid: true},
		},
	}
	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(cache *mockstore.MockCache, database *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"model_name": "test",
				"status":     "pending",
			},
			buildStubs: func(cache *mockstore.MockCache, database *mockdb.MockStore) {
				database.EXPECT().
					GetTasksByModelNameAndStatus(gomock.Any(), gomock.Any()).
					Times(1).
					Return(tasks, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				fmt.Println(recorder.Body)
			},
		},
		{
			name: "Internal error",
			body: gin.H{
				"model_name": "test",
				"status":     "pending",
			},
			buildStubs: func(cache *mockstore.MockCache, database *mockdb.MockStore) {
				database.EXPECT().
					GetTasksByModelNameAndStatus(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, errors.New("failed"))
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "Record not found",
			body: gin.H{
				"model_name": "test",
				"status":     "pending",
			},
			buildStubs: func(cache *mockstore.MockCache, database *mockdb.MockStore) {
				database.EXPECT().
					GetTasksByModelNameAndStatus(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, db.ErrRecordNotFound)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cache := mockstore.NewMockCache(ctrl)
			database := mockdb.NewMockStore(ctrl)
			tc.buildStubs(cache, database)

			server := newTestServer(t, nil, cache, database)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(
				http.MethodGet, "/task/modelstatus", bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}
