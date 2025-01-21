package test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"os"
	"screenshoter/internal"
	"screenshoter/internal/screenshot"
	"testing"
)

const successEndpoint = "https://google.com"
const failEndpoint = "https://cutcode.dev"

type requestCase struct {
	name       string
	method     string
	authHeader string
	body       internal.Request
	wantStatus int
	wantBody   internal.ErrorResponse
}

var requestCases = []requestCase{
	{
		name:       "Bad method",
		method:     http.MethodGet,
		authHeader: "Bearer " + os.Getenv("APP_TOKEN"),
		body:       internal.Request{URL: successEndpoint},
		wantStatus: http.StatusMethodNotAllowed,
		wantBody:   internal.ErrorResponse{Error: internal.OnlyPostAllowed},
	},
	{
		name:       "Invalid token",
		method:     http.MethodPost,
		authHeader: "Bearer invalid",
		body:       internal.Request{URL: successEndpoint},
		wantStatus: http.StatusUnauthorized,
		wantBody:   internal.ErrorResponse{Error: internal.TokenInvalid},
	},
}

var requestWithScreenCases = []requestCase{
	{
		name:       "Timeout",
		method:     http.MethodPost,
		authHeader: "Bearer " + os.Getenv("APP_TOKEN"),
		body:       internal.Request{URL: failEndpoint},
		wantStatus: http.StatusGatewayTimeout,
	},
	{
		name:       "Success",
		method:     http.MethodPost,
		authHeader: "Bearer " + os.Getenv("APP_TOKEN"),
		body:       internal.Request{URL: successEndpoint},
		wantStatus: http.StatusOK,
	},
}

func TestRequestHandler(t *testing.T) {
	for _, tt := range requestCases {
		t.Run(tt.name, func(t *testing.T) {
			bodyJson, err := json.Marshal(tt.body)

			if err != nil {
				t.Fatalf("Failed to marshall: %v", err)
			}

			wr := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, "/make-request", bytes.NewBuffer(bodyJson))
			req.Header.Set("Authorization", tt.authHeader)

			internal.Handle(wr, req)

			if wr.Code != tt.wantStatus {
				t.Errorf("got HTTP  status code %d, expected %d", wr.Code, tt.wantStatus)
			}

			var gotBody internal.ErrorResponse

			if err := json.NewDecoder(wr.Body).Decode(&gotBody); err != nil {
				t.Fatalf("Failed to decode response body: %v", err)
			}

			if gotBody != tt.wantBody {
				t.Errorf("got body %v, expected %v", gotBody, tt.wantBody)
			}
		})
	}
}

func TestCaptureScreenshot(t *testing.T) {
	mockCapture := new(screenshot.MockCapture)

	mockCapture.On("Capture", mock.Anything, successEndpoint).
		Return([]byte("fake_screenshot"), nil)

	mockCapture.On("Capture", mock.Anything, failEndpoint).
		Return([]byte{}, context.DeadlineExceeded)

	internal.SetCapture(mockCapture)

	t.Cleanup(func() {
		internal.SetCapture(nil)
	})

	for _, tt := range requestWithScreenCases {
		t.Run(tt.name, func(t *testing.T) {
			bodyJson, err := json.Marshal(tt.body)

			if err != nil {
				t.Fatalf("Failed to marshall: %v", err)
			}

			wr := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, "/make-request", bytes.NewBuffer(bodyJson))
			req.Header.Set("Authorization", tt.authHeader)

			internal.Handle(wr, req)

			assert.Equal(t, tt.wantStatus, wr.Code, "unexpected HTTP status code")
		})
	}

	mockCapture.AssertExpectations(t)
}
