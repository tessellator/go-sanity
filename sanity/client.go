package sanity

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// NewBool accepts a bool and returns a pointer to a bool with the same value.
//
// The Sanity client uses bool pointers when bool values are optional parameters
// to distinguish between unset and falsy values.
func NewBool(val bool) *bool {
	b := new(bool)
	*b = val

	return b
}

type service struct {
	client *Client
}

// Client is a client for the Sanity HTTP API.
type Client struct {
	// Projects is the client for the Projects API.
	Projects *ProjectsService

	client *http.Client

	baseURL string

	common service
}

// NewClient creates a new Sanity client.
//
// If `httpClient` is nil, the `http.DefaultClient` will be used.
// The `httpClient` is expected to provide authentication.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	client := &Client{
		client:  httpClient,
		baseURL: "https://api.sanity.io",
	}
	client.common.client = client
	client.Projects = (*ProjectsService)(&client.common)

	return client
}

func do(ctx context.Context, client *http.Client, url string, method string, body any, result any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		type errorMessage struct {
			Message string `json:"message"`
		}
		var msg errorMessage
		err = json.NewDecoder(resp.Body).Decode(&msg)
		if err != nil {
			return err
		}
		return errors.New(msg.Message)
	}

	return json.NewDecoder(resp.Body).Decode(result)
}
