package handlers

import (
	"github.com/smartfor/metrics/internal/logger"
	"github.com/smartfor/metrics/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testRequest(
	t *testing.T,
	ts *httptest.Server,
	method string,
	path string,
) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestRouter(t *testing.T) {
	type want struct {
		code     int
		response string
	}

	err := logger.Initialize("Info")
	if err != nil {
		log.Fatalf("Error initialize logger ")
	}

	s := storage.NewMemStorage()
	ts := httptest.NewServer(Router(s, logger.Log))
	defer ts.Close()

	tests := []struct {
		name       string
		method     string
		requestURL string
		want       want
	}{
		{
			name:       "Positive #1",
			requestURL: "/update/gauge/key1/1",
			method:     http.MethodPost,
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name:       "Positive - Get gauge metric value",
			requestURL: "/value/gauge/key1",
			method:     http.MethodGet,
			want: want{
				code:     http.StatusOK,
				response: "1",
			},
		},
		{
			name:       "Positive #2",
			requestURL: "/update/counter/key1/2",
			method:     http.MethodPost,
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name:       "Positive - Get counter metric value",
			requestURL: "/value/counter/key1",
			method:     http.MethodGet,
			want: want{
				code:     http.StatusOK,
				response: "2",
			},
		},
		{
			name:       "Negative - Unknown metric type",
			requestURL: "/value/foo/key1",
			method:     http.MethodGet,
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:       "Negative - Not found gauge metric",
			requestURL: "/value/counter/foo",
			method:     http.MethodGet,
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:       "Negative - Not found counter metric",
			requestURL: "/value/counter/foo",
			method:     http.MethodGet,
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:       "Negative - Get method not allowed",
			requestURL: "/update/counter/key1/2",
			method:     http.MethodGet,
			want: want{
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			name:       "Negative - Patch method not allowed",
			requestURL: "/update/counter/key1/2",
			method:     http.MethodPatch,
			want: want{
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			name:       "Negative - Put method not allowed",
			requestURL: "/update/counter/key1/2",
			method:     http.MethodPut,
			want: want{
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			name:       "Negative - Delete method not allowed",
			requestURL: "/update/counter/key1/2",
			method:     http.MethodDelete,
			want: want{
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			name:       "Negative - Metric type not passed",
			requestURL: "/update/key1/2",
			method:     http.MethodPost,
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:       "Negative - Unknown metric type passed",
			requestURL: "/update/foo/key1/2",
			method:     http.MethodPost,
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:       "Negative - invalid value passed",
			requestURL: "/update/gauge/key1/asdasd",
			method:     http.MethodPost,
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:       "Negative - invalid value passed",
			requestURL: "/update/counter/key1/asdasd",
			method:     http.MethodPost,
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:       "Positive - Get metrics page",
			requestURL: "/",
			method:     http.MethodGet,
			want: want{
				code: http.StatusOK,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, test.method, test.requestURL)
			defer resp.Body.Close()

			assert.Equal(t, test.want.code, resp.StatusCode)

			if test.want.response != "" {
				assert.Equal(t, test.want.response, body)
			}
		})
	}
}
