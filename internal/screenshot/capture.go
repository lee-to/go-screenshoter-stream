package screenshot

import (
	"context"
	"github.com/chromedp/chromedp"
)

type ICapture interface {
	Capture(ctx context.Context, url string) ([]byte, error)
}

type Capture struct{}

func (r *Capture) Capture(ctx context.Context, url string) ([]byte, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var screenshot []byte

	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(400, 400),
		chromedp.Navigate(url),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.FullScreenshot(&screenshot, 80),
	)

	if err != nil {
		return nil, err
	}

	return screenshot, nil
}
