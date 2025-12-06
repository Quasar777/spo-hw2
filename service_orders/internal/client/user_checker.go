package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type UsersClient struct {
	baseURL string
	client  *http.Client
}

func NewUsersClient(baseURL string) *UsersClient {
	return &UsersClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

func (c *UsersClient) UserExists(ctx context.Context, userID int) (bool, error) {
	url := fmt.Sprintf("%s/users/%d", c.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status from users service: %d", resp.StatusCode)
	}

	var body any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return false, err
	}

	return true, nil
}