package handlers

import (
	"github.com/smartfor/metrics/cmd/server/storage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMakeUpdateHandler(t *testing.T) {
	type want struct {
		code int
	}

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
			name:       "Positive #2",
			requestURL: "/update/counter/key1/2",
			method:     http.MethodPost,
			want: want{
				code: http.StatusOK,
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := storage.NewMemStorage()
			request := httptest.NewRequest(test.method, test.requestURL, nil)
			w := httptest.NewRecorder()
			MakeUpdateHandler(s)(w, request)
			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)
		})
	}
}
