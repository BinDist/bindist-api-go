// Package bindist provides a client for the BinDist API.
package bindist

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client is the BinDist API client.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new BinDist API client.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetHTTPClient sets a custom HTTP client.
func (c *Client) SetHTTPClient(client *http.Client) {
	c.httpClient = client
}

func (c *Client) doRequest(ctx context.Context, method, path string, query url.Values, body io.Reader, opts *requestOptions) (*http.Response, error) {
	reqURL := c.baseURL + path
	if len(query) > 0 {
		reqURL += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	if opts != nil && opts.channel != "" {
		req.Header.Set("X-Channel", opts.channel)
	}

	return c.httpClient.Do(req)
}

func parseResponse[T any](resp *http.Response) (*Response[T], error) {
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp apiResponse[T]
	parseErr := json.Unmarshal(bodyBytes, &apiResp)

	out := &Response[T]{
		Success:    apiResp.Success,
		Data:       apiResp.Data,
		Error:      apiResp.Error,
		Meta:       apiResp.Meta,
		HTTPStatus: resp.StatusCode,
	}

	if resp.StatusCode >= 400 {
		// Non-2xx: ensure callers see Success=false and a populated Error,
		// even when the body isn't a standard envelope (e.g. plain
		// {"message":"Unauthorized"}, an HTML 502 page, or empty body).
		out.Success = false
		if out.Error == nil {
			out.Error = synthesizeError(resp.StatusCode, bodyBytes, parseErr)
		} else if out.Error.HTTPStatus == 0 {
			out.Error.HTTPStatus = resp.StatusCode
		}
		return out, nil
	}

	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse response: %w", parseErr)
	}

	return out, nil
}

// synthesizeError builds an ApiError for a non-2xx response whose body
// did not contain a standard error envelope. It tries to extract a useful
// message from a bare {"message":"..."} body, falling back to the HTTP
// status text.
func synthesizeError(status int, body []byte, parseErr error) *ApiError {
	message := ""
	if parseErr == nil {
		// Body parsed as JSON but didn't fill apiResp.Error. Try a bare
		// {"message": "..."} or {"error": "..."} shape.
		var bare struct {
			Message string `json:"message"`
			Error   string `json:"error"`
		}
		if json.Unmarshal(body, &bare) == nil {
			if bare.Message != "" {
				message = bare.Message
			} else if bare.Error != "" {
				message = bare.Error
			}
		}
	}
	if message == "" {
		message = http.StatusText(status)
	}
	if message == "" {
		message = "http error"
	}
	return &ApiError{
		Code:       httpStatusCode(status),
		Message:    message,
		HTTPStatus: status,
	}
}

// httpStatusCode returns a stable string code for common HTTP statuses
// so callers can switch on err.Code without parsing the message.
func httpStatusCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusConflict:
		return "conflict"
	case http.StatusTooManyRequests:
		return "rate_limited"
	}
	switch {
	case status >= 500:
		return "server_error"
	case status >= 400:
		return "http_error"
	}
	return "http_error"
}

// ListApplications returns a list of available applications.
func (c *Client) ListApplications(ctx context.Context, opts *ListApplicationsOptions) (*Response[[]Application], error) {
	query := url.Values{}

	if opts != nil {
		if opts.Page > 0 {
			query.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			query.Set("pageSize", strconv.Itoa(opts.PageSize))
		}
		if opts.Search != "" {
			query.Set("search", opts.Search)
		}
		if len(opts.Tags) > 0 {
			query.Set("tags", strings.Join(opts.Tags, ","))
		}
	}

	resp, err := c.doRequest(ctx, "GET", "/v1/applications", query, nil, nil)
	if err != nil {
		return nil, err
	}

	result, err := parseResponse[applicationsData](resp)
	if err != nil {
		return nil, err
	}

	return &Response[[]Application]{
		Success:    result.Success,
		Data:       result.Data.Applications,
		Error:      result.Error,
		Meta:       result.Meta,
		HTTPStatus: result.HTTPStatus,
	}, nil
}

// GetApplication returns details for a specific application.
func (c *Client) GetApplication(ctx context.Context, applicationID string) (*Response[Application], error) {
	resp, err := c.doRequest(ctx, "GET", "/v1/applications/"+url.PathEscape(applicationID), nil, nil, nil)
	if err != nil {
		return nil, err
	}

	return parseResponse[Application](resp)
}

// ListVersions returns all versions for an application.
//
// Pass WithChannel to access versions outside the default production
// channel (e.g. WithChannel(ChannelTest) to include disabled versions).
func (c *Client) ListVersions(ctx context.Context, applicationID string, opts ...RequestOption) (*Response[[]Version], error) {
	path := fmt.Sprintf("/v1/applications/%s/versions", url.PathEscape(applicationID))

	resp, err := c.doRequest(ctx, "GET", path, nil, nil, newRequestOptions(opts))
	if err != nil {
		return nil, err
	}

	result, err := parseResponse[versionsData](resp)
	if err != nil {
		return nil, err
	}

	return &Response[[]Version]{
		Success:    result.Success,
		Data:       result.Data.Versions,
		Error:      result.Error,
		Meta:       result.Meta,
		HTTPStatus: result.HTTPStatus,
	}, nil
}

// ListVersionFiles returns all files for a specific version.
func (c *Client) ListVersionFiles(ctx context.Context, applicationID, version string) (*Response[[]VersionFile], error) {
	path := fmt.Sprintf("/v1/applications/%s/versions/%s/files",
		url.PathEscape(applicationID), url.PathEscape(version))

	resp, err := c.doRequest(ctx, "GET", path, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	result, err := parseResponse[filesData](resp)
	if err != nil {
		return nil, err
	}

	return &Response[[]VersionFile]{
		Success:    result.Success,
		Data:       result.Data.Files,
		Error:      result.Error,
		Meta:       result.Meta,
		HTTPStatus: result.HTTPStatus,
	}, nil
}

// GetDownloadInfo returns download URL and metadata for a file.
//
// Pass WithChannel to obtain a download URL for a version that is only
// accessible on a non-default channel (e.g. WithChannel(ChannelTest) for
// disabled versions).
func (c *Client) GetDownloadInfo(ctx context.Context, applicationID, version, fileID string, opts ...RequestOption) (*Response[DownloadInfo], error) {
	query := url.Values{}
	query.Set("applicationId", applicationID)
	query.Set("version", version)
	if fileID != "" {
		query.Set("fileId", fileID)
	}

	resp, err := c.doRequest(ctx, "GET", "/v1/downloads/url", query, nil, newRequestOptions(opts))
	if err != nil {
		return nil, err
	}

	return parseResponse[DownloadInfo](resp)
}

// DownloadFile downloads a file and returns its contents along with metadata.
// If verifyChecksum is true, the checksum is verified against the expected value.
//
// Pass WithChannel to download from a non-default release channel.
func (c *Client) DownloadFile(ctx context.Context, applicationID, version, fileID string, verifyChecksum bool, opts ...RequestOption) ([]byte, *DownloadInfo, error) {
	// Get download info
	infoResp, err := c.GetDownloadInfo(ctx, applicationID, version, fileID, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get download info: %w", err)
	}

	if !infoResp.Success {
		if infoResp.Error != nil {
			return nil, nil, fmt.Errorf("API error: %s - %s", infoResp.Error.Code, infoResp.Error.Message)
		}
		return nil, nil, fmt.Errorf("failed to get download URL")
	}

	// Download the file
	req, err := http.NewRequestWithContext(ctx, "GET", infoResp.Data.URL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create download request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file content: %w", err)
	}

	// Verify checksum if requested
	if verifyChecksum && infoResp.Data.Checksum != "" {
		hash := sha256.Sum256(content)
		actualChecksum := hex.EncodeToString(hash[:])
		if actualChecksum != infoResp.Data.Checksum {
			return nil, nil, fmt.Errorf("checksum mismatch: expected %s, got %s",
				infoResp.Data.Checksum, actualChecksum)
		}
	}

	return content, &infoResp.Data, nil
}

// DownloadFileToWriter downloads a file and writes it to the provided writer.
// Returns the download metadata.
//
// Pass WithChannel to download from a non-default release channel.
func (c *Client) DownloadFileToWriter(ctx context.Context, applicationID, version, fileID string, w io.Writer, opts ...RequestOption) (*DownloadInfo, error) {
	// Get download info
	infoResp, err := c.GetDownloadInfo(ctx, applicationID, version, fileID, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get download info: %w", err)
	}

	if !infoResp.Success {
		if infoResp.Error != nil {
			return nil, fmt.Errorf("API error: %s - %s", infoResp.Error.Code, infoResp.Error.Message)
		}
		return nil, fmt.Errorf("failed to get download URL")
	}

	// Download the file
	req, err := http.NewRequestWithContext(ctx, "GET", infoResp.Data.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to write file content: %w", err)
	}

	return &infoResp.Data, nil
}

// CreateShareLink creates a shareable download link for a file.
func (c *Client) CreateShareLink(ctx context.Context, applicationID, version, fileID string, expiresMinutes int) (*Response[ShareLink], error) {
	payload := map[string]interface{}{
		"applicationId":  applicationID,
		"version":        version,
		"expiresMinutes": expiresMinutes,
	}
	if fileID != "" {
		payload["fileId"] = fileID
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(ctx, "POST", "/v1/downloads/share", nil, bytes.NewReader(body), nil)
	if err != nil {
		return nil, err
	}

	return parseResponse[ShareLink](resp)
}

// GetStats returns download statistics for an application.
func (c *Client) GetStats(ctx context.Context, applicationID string) (*Response[Stats], error) {
	path := fmt.Sprintf("/v1/applications/%s/stats", url.PathEscape(applicationID))

	resp, err := c.doRequest(ctx, "GET", path, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	return parseResponse[Stats](resp)
}
