// wrapper for HTTP API for scheduler, kubelet, contoller manager and CLI

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	v1 "github.com/therandombyte/mini-k8s/pkg/api/v1"
)

type Client struct {
	BaseURL string
	HTTP *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTP: &http.Client{},
	}
}

// context to control the timeout, cancellation, deadlines
// path is appened to baseurl
func (c *Client) do(ctx context.Context, method, path string, in, out any) error {
	// use pointer to avoid copying large struct, and to represent "no value" using nil which can be initialized later
	// bytes.Buffer is a struct from bytes package, that implements a growable buffer of of bytes with Read/Write methods
	var body *bytes.Buffer

	// initialize body whether input is present or not
	if in != nil {
		body = new(bytes.Buffer)
		if err := json.NewEncoder(body).Encode(in); err != nil {
			return err
		}
	} else {
		body = bytes.NewBuffer(nil)
	}

	// creating HTTP Request
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+ path, body)
	if err != nil {
		return err
	}

	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// execute the HTTP request
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// handle status codes
	if resp.StatusCode >= 300 {
		return fmt.Errorf("request failed: %s", resp.Status)
	}

	// decoding response body
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}

	return nil
}

func (c *Client) GetPod(ctx context.Context, name string) (*v1.Pod, error) {
	var out v1.Pod
	return &out, c.do(ctx, http.MethodGet, "/api/v1/pods"+name, nil, &out)
}

func (c *Client) ListPods(ctx context.Context) (*v1.PodList, error) {
	var out v1.PodList
	return &out, c.do(ctx, http.MethodGet, "/api/v1/pods", nil, &out)
}

func (c *Client) CreatePod(ctx context.Context, pod *v1.Pod) error {
	return c.do(ctx, http.MethodPost, "/api/v1/pods", pod, nil)
}

func (c *Client) DeletePod(ctx context.Context, name string) error {
	return c.do(ctx, http.MethodDelete, "/api/v1/pods"+name, nil, nil)
}

func (c *Client) CreateNode(ctx context.Context, node *v1.Node) error {
	return nil
}

func (c *Client) UpdatePod(ctx context.Context, pod *v1.Pod) error {
	return c.do(ctx, http.MethodPut, "/api/v1/pods/"+ pod.Metadata.Name, pod, nil)
}

func (c *Client) UpdatePodStatus(ctx context.Context, name string, status *v1.PodStatus) error {
	return c.do(ctx, http.MethodPut, "/api/v1/pods/" + name + "/status", status, nil)
}

func (c *Client) UpdateNodeStatus(ctx context.Context, name string, status *v1.NodeStatus) error {
	return nil
}

func (c *Client) ListNodes(ctx context.Context) (*v1.NodeList, error) {
	return nil, nil
}
