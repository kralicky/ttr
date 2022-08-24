package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	apiEndpoint           = `https://www.toontownrewritten.com/api`
	patchManifestEndpoint = `https://cdn.toontownrewritten.com/content/patchmanifest.txt`
	downloadEndpoint      = `https://download.toontownrewritten.com/patches/`
)

type SuccessKind string

const (
	SuccessTrue    SuccessKind = "true"
	SuccessFalse   SuccessKind = "false"
	SuccessPartial SuccessKind = "partial"
	SuccessDelayed SuccessKind = "delayed"
)

type LoginResponse struct {
	Success SuccessKind `json:"success"`
	Message string      `json:"banner"`

	// Will be set if success == "true"
	*LoginSuccessPayload `json:",inline,omitempty"`

	// Will be set if success == "partial"
	*LoginPartialSuccessPayload `json:",inline,omitempty"`

	// Will be set if success == "delayed"
	*LoginDelayedSuccessPayload `json:",inline,omitempty"`
}

type LoginSuccessPayload struct {
	Gameserver string `json:"gameserver,omitempty"`
	Cookie     string `json:"cookie,omitempty"`
}

type LoginPartialSuccessPayload struct {
	ResponseToken string `json:"responseToken,omitempty"`
}

type LoginDelayedSuccessPayload struct {
	QueueToken string `json:"queueToken,omitempty"`
	ETA        int    `json:"eta,string,omitempty"`
	Position   int    `json:"position,string,omitempty"`
}

type PatchManifest map[string]*PatchSpec

type PatchSpec struct {
	Download       string   `json:"dl"`
	Only           []string `json:"only"`
	Hash           string   `json:"hash"`
	CompressedHash string   `json:"compHash"`
}

type LoginClient interface {
	Login(ctx context.Context, username, password string) (*LoginResponse, error)
	RetryDelayedLogin(ctx context.Context, queueToken string) (*LoginResponse, error)
}

type DownloadClient interface {
	DownloadPatchManifest(ctx context.Context) (PatchManifest, error)
	DownloadFile(ctx context.Context, name string) (io.ReadCloser, error)
}

type Client interface {
	LoginClient
	DownloadClient
}

type client struct {
	httpClient *http.Client
}

func NewClient() Client {
	return &client{
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{},
			},
		},
	}
}

func (c *client) Login(ctx context.Context, username, password string) (*LoginResponse, error) {
	form := url.Values{}
	form.Add("username", username)
	form.Add("password", password)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiEndpoint+"/login?format=json", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return nil, err
	}
	return &loginResp, nil
}

func (c *client) RetryDelayedLogin(ctx context.Context, queueToken string) (*LoginResponse, error) {
	form := url.Values{}
	form.Add("queueToken", queueToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiEndpoint+"/login?format=json", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return nil, err
	}
	return &loginResp, nil
}

func (c *client) DownloadPatchManifest(ctx context.Context) (PatchManifest, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, patchManifestEndpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var patchManifest PatchManifest
	if err := json.NewDecoder(resp.Body).Decode(&patchManifest); err != nil {
		return nil, err
	}
	return patchManifest, nil
}

func (c *client) DownloadFile(ctx context.Context, name string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadEndpoint+name, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("download failed: unexpected status: %s", resp.Status)
	}
	return resp.Body, nil
}
