package googleBook

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	GoogleBooksAPIHost     = "https://www.googleapis.com"
	GoogleBooksAPIBasePath = "/v1/volumes"
	ImageNotFoundPath      = "images/image_not_found.png"
)

type Client struct {
	host     string
	basePath string
	client   http.Client
}

func New(host string) Client {
	if host == "" {
		host = GoogleBooksAPIHost
	}
	return Client{
		host:     host,
		basePath: GoogleBooksAPIBasePath,
		client: http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) Books(searchMsg string) (GoogleBookResponse, error) {
	req, err := http.NewRequest(http.MethodGet, c.host+c.basePath, nil)
	if err != nil {
		return GoogleBookResponse{}, fmt.Errorf("can't create request to %s: %w", c.host+c.basePath, err)
	}

	params := url.Values{}
	params.Add("q", searchMsg)

	req.URL.RawQuery = params.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return GoogleBookResponse{}, fmt.Errorf("can't do request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return GoogleBookResponse{}, fmt.Errorf("unexpected status code %d from %s: %s", resp.StatusCode, req.URL.String(), strings.TrimSpace(string(snippet)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GoogleBookResponse{}, fmt.Errorf("failed to read response body from %s: %w", req.URL.String(), err)
	}

	var googleResponce GoogleBookResponse
	if json.Valid(body) {
		err = json.Unmarshal(body, &googleResponce)
		if err != nil {
			return GoogleBookResponse{}, fmt.Errorf("json unmarshall error: %w", err)
		}
	} else {
		return GoogleBookResponse{}, fmt.Errorf("invalid JSON response from %s", req.URL.String())
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
		notFoundImg, err := os.ReadFile(ImageNotFoundPath)
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
