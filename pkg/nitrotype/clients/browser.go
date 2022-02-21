package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"nt-folly-xmaxx-comp/pkg/nitrotype"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type APIClientBrowser struct {
	options []chromedp.ExecAllocatorOption
}

func NewAPIClientBrowser(userAgent string) *APIClientBrowser {
	return &APIClientBrowser{
		options: []chromedp.ExecAllocatorOption{
			chromedp.UserAgent(userAgent),
			chromedp.NoFirstRun,
			chromedp.NoDefaultBrowserCheck,
			chromedp.NoSandbox,
			chromedp.Headless,
			chromedp.DisableGPU,
		},
	}
}

func (c *APIClientBrowser) getRequest(url string, timeout int) ([]byte, error) {
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), c.options...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(
		ctx,
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	var requestID network.RequestID
	downloadComplete := make(chan bool)

	chromedp.ListenTarget(ctx, func(v interface{}) {
		switch ev := v.(type) {
		case *network.EventRequestWillBeSent:
			if ev.Request.URL == url {
				requestID = ev.RequestID
			}
		case *network.EventLoadingFinished:
			if ev.RequestID == requestID {
				close(downloadComplete)
			}
		}
	})

	err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate("view-source:"+url),
	)
	if err != nil {
		return nil, err
	}

	// This will block until the chromedp listener closes the channel
	<-downloadComplete

	// get the downloaded bytes for the request id
	var downloadBytes []byte
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		downloadBytes, err = network.GetResponseBody(requestID).Do(ctx)
		return err
	})); err != nil {
		log.Fatal(err)
	}

	return downloadBytes, nil
}

func (c *APIClientBrowser) GetTeam(tagName string) (*nitrotype.TeamV2APIResponse, error) {
	resp, err := c.getRequest("https://www.nitrotype.com/api/v2/teams/"+tagName, 30)
	if err != nil {
		return nil, fmt.Errorf("failed to request api team data: %w", err)
	}
	var output nitrotype.TeamV2APIResponse
	if err := json.Unmarshal(resp, &output); err != nil {
		return nil, fmt.Errorf("unmarshal nt team api response failed: %w", err)
	}
	return &output, nil
}

func (c *APIClientBrowser) GetProfile(username string) (*nitrotype.UserProfile, error) {
	resp, err := c.getRequest("https://www.nitrotype.com/racer/"+username, 30)
	if err != nil {
		return nil, fmt.Errorf("failed to request racer profile: %w", err)
	}
	matches := nitrotype.NTUserProfileExtractRegExp.FindSubmatch(resp)
	if len(matches) != 2 {
		return nil, nitrotype.ErrNTUserProfileNotFound
	}
	var output nitrotype.UserProfile
	if err := json.Unmarshal(matches[1], &output); err != nil {
		return nil, fmt.Errorf("unmarshal nt racer data failed: %w", err)
	}
	return &output, nil
}

func init() {
	chromedp.Flag("disable-setuid-sandbox", true)
}
