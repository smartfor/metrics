package handlers

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/smartfor/metrics/internal/logger"
	"github.com/smartfor/metrics/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(
	t *testing.T,
	ts *httptest.Server,
	method string,
	path string,
	body string,
) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, strings.NewReader(body))
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestRouter(t *testing.T) {
	type want struct {
		response    string
		contentType string
		code        int
	}

	zlog, err := logger.MakeLogger("Info")
	if err != nil {
		log.Fatalf("Error initialize logger ")
	}

	fs, err := storage.NewFileStorage("/tmp/metrics.json")
	if err != nil {
		t.Fatal(err)
	}
	s, err := storage.NewMemStorage(fs, false, false)
	if err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(Router(s, zlog, "", nil, ""))
	defer ts.Close()

	tests := []struct {
		name        string
		method      string
		requestURL  string
		requestBody string
		want        want
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

		// JSON HANDLERS TESTS
		{
			name:        "JSON :: Positive  #1",
			requestURL:  "/update/",
			method:      http.MethodPost,
			requestBody: `{ "id": "key1", "type": "gauge", "value": 1 }`,
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
			},
		},
		{
			name:        "JSON :: Positive - Get gauge metric value",
			requestURL:  "/value/",
			method:      http.MethodPost,
			requestBody: `{ "id": "key1", "type": "gauge" }`,
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				response: `{
					"id": "key1",
					"type": "gauge",
					"value": 1
				}`,
			},
		},
		{
			name:        "JSON :: Positive #2",
			requestURL:  "/update/",
			requestBody: `{ "id": "counterKey1", "type": "counter", "delta": 2}`,
			method:      http.MethodPost,
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				response: `{
					"id": "counterKey1",
					"type": "counter",
					"delta": 2
				}`,
			},
		},
		{
			name:        "JSON :: Positive - Get counter metric value",
			requestURL:  "/value/",
			method:      http.MethodPost,
			requestBody: `{ "id": "counterKey1", "type": "counter"}`,
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				response:    `{ "id": "counterKey1", "type": "counter", "delta": 2}`,
			},
		},
		{
			name:        "JSON :: Negative - Unknown metric type",
			requestURL:  "/value/",
			method:      http.MethodPost,
			requestBody: `{ "id": "key1", "type": "foo"}`,
			want: want{
				code:        http.StatusNotFound,
				contentType: "application/json",
			},
		},
		{
			name:        "JSON :: Negative - Not found gauge metric",
			requestURL:  "/value/",
			requestBody: `{ "id": "foo", "type": "gauge"}`,
			method:      http.MethodPost,
			want: want{
				code:        http.StatusNotFound,
				contentType: "application/json",
			},
		},
		{
			name:        "JSON :: Negative - Not found counter metric",
			requestURL:  "/value/",
			requestBody: `{ "id": "foo", "type": "counter"}`,
			method:      http.MethodPost,
			want: want{
				code:        http.StatusNotFound,
				contentType: "application/json",
			},
		},
		{
			name:        "JSON :: Negative - Metric type not passed",
			requestURL:  "/update/",
			method:      http.MethodPost,
			requestBody: `{ "id": "key1", "value": 2, "delta": 2 }`,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "application/json",
			},
		},
		{
			name:        "JSON :: Negative - Unknown metric type passed",
			requestURL:  "/update/",
			requestBody: `{ "id": "key1", "type":"foo", "value": 2, "delta": 2 }`,
			method:      http.MethodPost,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "application/json",
			},
		},
		{
			name:        "JSON :: Negative - invalid value passed",
			requestURL:  "/update/",
			method:      http.MethodPost,
			requestBody: `{ "id": "key1", "type":"gauge", "value": "asdasd"}`,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "application/json",
			},
		},
		{
			name:        "JSON :: Negative - invalid value passed",
			requestURL:  "/update/",
			requestBody: `{ "id": "key1", "type":"counter", "delta": "asdasd"}`,
			method:      http.MethodPost,
			want: want{
				code:        http.StatusBadRequest,
				contentType: "application/json",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, test.method, test.requestURL, test.requestBody)
			defer resp.Body.Close()

			assert.Equal(t, test.want.code, resp.StatusCode)
			if test.want.contentType != "" {
				assert.Equal(t, test.want.contentType, resp.Header.Get("Content-Type"))
			}

			if resp.StatusCode >= 400 {
				return
			}

			if test.requestBody != "" {
				if test.requestBody != "" && test.want.response != "" {
					assert.JSONEq(t, test.want.response, body)
				}
			} else {
				if test.want.response != "" {
					assert.Equal(t, test.want.response, body)
				}
			}

			if test.want.contentType != "" {
				assert.Equal(t, test.want.contentType, resp.Header.Get("Content-Type"))
			}
		})
	}
}
