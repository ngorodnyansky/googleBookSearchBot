package googleBook

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Client struct {
	host     string
	basePath string
	client   http.Client
}

func New(host string) Client {
	return Client{
		host:     host,
		basePath: "/v1/volumes",
		client: http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) Books(searchMsg string) (GoogleBookResponce, error) {
	req, err := http.NewRequest(http.MethodGet, c.host+c.basePath, nil)
	if err != nil {
		return GoogleBookResponce{}, fmt.Errorf("can't create request: %w", err)
	}

	params := url.Values{}
	params.Add("q", searchMsg)

	req.URL.RawQuery = params.Encode()
	resp, err := c.client.Do(req)
	if err != nil {
		return GoogleBookResponce{}, fmt.Errorf("can't do request: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GoogleBookResponce{}, fmt.Errorf("can't read responce body: %w", err)
	}

	var googleResponce GoogleBookResponce
	if json.Valid(body) {
		err = json.Unmarshal(body, &googleResponce)
		if err != nil {
			return GoogleBookResponce{}, fmt.Errorf("json unmarshall error: %w", err)
		}
	} else {
		return GoogleBookResponce{}, fmt.Errorf("invalid json response")
	}

	return googleResponce, nil
}

func (c *Client) BookImage(links ImageLinks) ([]byte, error) {
	var imageLink string
	if len(links.Medium) != 0 {
		imageLink = links.Medium
	} else if len(links.Small) != 0 {
		imageLink = links.Small
	} else if len(links.Thumbnail) != 0 {
		imageLink = links.Thumbnail
	} else {
		notFoundImg, err := os.ReadFile("images/image_not_found.png")
		if err != nil {
			return []byte{}, fmt.Errorf("can't read image_not_found.png: %w", err)
		}
		return notFoundImg, nil
	}

	resp, err := c.client.Get(imageLink)
	if err != nil {
		return []byte{}, fmt.Errorf("image download failed: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("can't read responce body: %w", err)
	}

	return body, nil
}
