package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chromedp/chromedp"
	"log"
	"net/http"
	"strings"
	"time"
)

type ScreenshotRequest struct {
	URL string `json:"url"`
}

func main() {
	http.HandleFunc("/make-screenshot", handle)
	port := "8383"
	fmt.Printf("Server started %s", port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handle(w http.ResponseWriter, r *http.Request) {
	const token = "secret"

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")

	if strings.TrimPrefix(authHeader, "Bearer ") != token {
		http.Error(w, "Token invalid", http.StatusUnauthorized)
		return
	}

	var screenshotRequest ScreenshotRequest

	if err := json.NewDecoder(r.Body).Decode(&screenshotRequest); err != nil {
		http.Error(w, "Payload invalid", http.StatusInternalServerError)
		return
	}

	screenshot, err := captureScreenshot(screenshotRequest.URL)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", `attachment; filename="screenshot.png"`)
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(screenshot)

	if err != nil {
		return
	}
}

func captureScreenshot(url string) ([]byte, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var screenshot []byte

	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(400, 400),
		chromedp.Navigate(url),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.FullScreenshot(&screenshot, 80),
	)

	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		return nil, err
	}

	return screenshot, nil
}
