package clients

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"nt-folly-xmaxx-comp/pkg/nitrotype"
)

type APIClientHTTP struct {
	client *http.Client
}

func NewAPIClientHTTP(client *http.Client) *APIClientHTTP {
	return &APIClientHTTP{client}
}

func (c *APIClientHTTP) GetTeam(tagName string) (*nitrotype.TeamV2APIResponse, error) {
	resp, err := c.client.Get("https://www.nitrotype.com/api/teams/" + tagName)
	if err != nil {
		return nil, fmt.Errorf("failed to http get: %w", err)
	}
	defer resp.Body.Close()

	var output nitrotype.TeamV2APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
		return nil, fmt.Errorf("unmarshal nt team api response failed: %w", err)
	}
	return &output, nil
}

func (c *APIClientHTTP) GetProfile(username string) (*nitrotype.UserProfile, error) {
	resp, err := c.client.Get("https://www.nitrotype.com/racer/" + username)
	if err != nil {
		return nil, fmt.Errorf("failed to http get: %w", err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	matches := nitrotype.NTUserProfileExtractRegExp.FindSubmatch(data)
	if len(matches) != 2 {
		return nil, nitrotype.ErrNTUserProfileNotFound
	}

	var output nitrotype.UserProfile
	if err := json.Unmarshal(matches[1], &output); err != nil {
		return nil, fmt.Errorf("unmarshal nt racer data failed: %w", err)
	}
	return &output, nil
}
