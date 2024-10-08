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

type PatchManifest map[string]*ManifestEntry

type ManifestEntry struct {
	Download       string                `json:"dl"`
	Only           []string              `json:"only"`
	Hash           string                `json:"hash"`
	CompressedHash string                `json:"compHash"`
	Patches        map[string]*PatchSpec `json:"patches"`
}

type PatchSpec struct {
	Filename            string `json:"filename"`
	PatchHash           string `json:"patchHash"`
	CompressedPatchHash string `json:"compPatchHash"`
}

type StatusSpec struct {
	Open               bool   `json:"open"`
	Banner             string `json:"banner"`
	LastCookieIssuedAt int64  `json:"lastCookieIssuedAt"`
	LastGameAuthAt     int64  `json:"lastGameAuthAt"`
}

type LoginClient interface {
	Login(ctx context.Context, username, password string) (*LoginResponse, error)
	RetryDelayedLogin(ctx context.Context, queueToken string) (*LoginResponse, error)
	CompleteTwoFactorAuth(ctx context.Context, responseToken, code string) (*LoginResponse, error)
}

type DownloadClient interface {
	DownloadPatchManifest(ctx context.Context) (PatchManifest, error)
	DownloadFile(ctx context.Context, name string) (io.ReadCloser, error)
}

type Client interface {
	LoginClient
	DownloadClient

	Status(ctx context.Context) (StatusSpec, error)
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
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("API error: %s", string(respData))
	}
	var loginResp LoginResponse
	if err := json.Unmarshal(respData, &loginResp); err != nil {
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
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("API error: %s", string(respData))
	}
	var loginResp LoginResponse
	if err := json.Unmarshal(respData, &loginResp); err != nil {
		return nil, err
	}
	return &loginResp, nil
}

func (c *client) CompleteTwoFactorAuth(ctx context.Context, responseToken, code string) (*LoginResponse, error) {
	form := url.Values{}
	form.Add("authToken", responseToken)
	form.Add("appToken", code)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiEndpoint+"/login?format=json", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("API error: %s", string(respData))
	}
	var loginResp LoginResponse
	if err := json.Unmarshal(respData, &loginResp); err != nil {
		return nil, err
	}
	if loginResp.Success == "partial" {
		return nil, fmt.Errorf("API error submitting 2FA code (try logging in to the website once): %s", loginResp.Message)
	}
	return &loginResp, nil
}

func (c *client) DownloadPatchManifest(ctx context.Context) (PatchManifest, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, patchManifestEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
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

func (c *client) Status(ctx context.Context) (StatusSpec, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiEndpoint+"/status", nil)
	if err != nil {
		return StatusSpec{}, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return StatusSpec{}, err
	}
	defer resp.Body.Close()
	var status StatusSpec
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return StatusSpec{}, err
	}
	return status, nil
}
