package internal

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"screenshoter/internal/screenshot"
	"strings"
	"time"
)

type Request struct {
	URL string `json:"url"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func sendErrorResponse(w http.ResponseWriter, status int, message string) {
	response := ErrorResponse{
		Error: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(response)

	if err != nil {
		log.Printf("Failed to send response: %v", err)
	}
}

var capture screenshot.ICapture = &screenshot.Capture{}

func SetCapture(c screenshot.ICapture) {
	capture = c
}

func Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendErrorResponse(w, http.StatusMethodNotAllowed, OnlyPostAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")

	if strings.TrimPrefix(authHeader, "Bearer ") != os.Getenv("APP_TOKEN") {
		sendErrorResponse(w, http.StatusUnauthorized, TokenInvalid)
		return
	}

	var screenshotRequest Request

	if err := json.NewDecoder(r.Body).Decode(&screenshotRequest); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, PayloadInvalid)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	res, err := capture.Capture(ctx, screenshotRequest.URL)

	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			sendErrorResponse(w, http.StatusGatewayTimeout, ScreenshotTimeout)
		default:
			sendErrorResponse(w, http.StatusInternalServerError, ScreenshotCaptureFailed)
		}

		log.Printf("Error capturing: %v", err)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", `attachment; filename="screenshot.png"`)
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(res)

	if err != nil {
		return
	}
}
