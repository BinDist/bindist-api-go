package bindist

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
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

// AdminClient is the BinDist admin API client.
type AdminClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewAdminClient creates a new BinDist admin API client.
func NewAdminClient(baseURL, apiKey string) *AdminClient {
	return &AdminClient{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetHTTPClient sets a custom HTTP client.
func (c *AdminClient) SetHTTPClient(client *http.Client) {
	c.httpClient = client
}

func (c *AdminClient) doRequest(ctx context.Context, method, path string, query url.Values, body io.Reader) (*http.Response, error) {
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

	return c.httpClient.Do(req)
}

// Customer represents a customer.
type Customer struct {
	CustomerID string    `json:"customerId"`
	Name       string    `json:"name"`
	APIKey     string    `json:"apiKey,omitempty"`
	IsActive   bool      `json:"isActive"`
	Notes      string    `json:"notes,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

// CreateCustomerResponse is the response from creating a customer.
type CreateCustomerResponse struct {
	CustomerID string `json:"customerId"`
	APIKey     string `json:"apiKey"`
	Name       string `json:"name"`
	CreatedAt  string `json:"createdAt"`
}

// CreateApplicationOptions contains options for creating an application.
type CreateApplicationOptions struct {
	ApplicationID string   `json:"applicationId"`
	Name          string   `json:"name"`
	CustomerIDs   []string `json:"customerIds"`
	Description   string   `json:"description,omitempty"`
	Tags          []string `json:"tags,omitempty"`
}

// UploadResponse is the response from uploading a file.
type UploadResponse struct {
	Message       string `json:"message"`
	VersionID     string `json:"versionId"`
	ApplicationID string `json:"applicationId"`
	Version       string `json:"version"`
	FileSize      int64  `json:"fileSize"`
	Checksum      string `json:"checksum"`
}

// LargeUploadURLResponse is the response from getting a large upload URL.
type LargeUploadURLResponse struct {
	UploadID  string `json:"uploadId"`
	UploadURL string `json:"uploadUrl"`
}

// UpdateVersionOptions contains options for updating a version.
type UpdateVersionOptions struct {
	IsEnabled    *bool   `json:"isEnabled,omitempty"`
	IsActive     *bool   `json:"isActive,omitempty"`
	ReleaseNotes *string `json:"releaseNotes,omitempty"`
}

// Activity represents an upload or download activity.
type Activity struct {
	Type          string    `json:"type"`
	ApplicationID string    `json:"applicationId"`
	Version       string    `json:"version"`
	CustomerID    string    `json:"customerId"`
	Timestamp     time.Time `json:"timestamp"`
}

// customersData wraps the customers array.
type customersData struct {
	Customers []Customer `json:"customers"`
}

// activitiesData wraps the activities array.
type activitiesData struct {
	Activities []Activity `json:"activities"`
}

// CreateCustomer creates a new customer with an API key.
func (c *AdminClient) CreateCustomer(ctx context.Context, name string, parentCustomerID string, notes string) (*Response[CreateCustomerResponse], error) {
	if parentCustomerID == "" {
		parentCustomerID = "admin"
	}

	payload := map[string]interface{}{
		"name": name,
	}
	if notes != "" {
		payload["notes"] = notes
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/v1/management/customers/%s/apikeys", url.PathEscape(parentCustomerID))
	resp, err := c.doRequest(ctx, "POST", path, nil, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	return parseResponse[CreateCustomerResponse](resp)
}

// CreateApplication creates a new application.
func (c *AdminClient) CreateApplication(ctx context.Context, opts CreateApplicationOptions) (*Response[Application], error) {
	body, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(ctx, "POST", "/v1/management/applications", nil, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	return parseResponse[Application](resp)
}

// UploadSmallFile uploads a small file (< 10MB) directly.
func (c *AdminClient) UploadSmallFile(ctx context.Context, applicationID, version, fileName string, content []byte, releaseNotes string) (*Response[UploadResponse], error) {
	payload := map[string]interface{}{
		"applicationId": applicationID,
		"version":       version,
		"fileName":      fileName,
		"fileContent":   base64.StdEncoding.EncodeToString(content),
		"fileType":      "MAIN",
	}
	if releaseNotes != "" {
		payload["releaseNotes"] = releaseNotes
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(ctx, "POST", "/v1/management/upload", nil, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	return parseResponse[UploadResponse](resp)
}

// GetLargeUploadURL gets a pre-signed URL for uploading a large file.
func (c *AdminClient) GetLargeUploadURL(ctx context.Context, applicationID, version, fileName string, fileSize int64, contentType string) (*Response[LargeUploadURLResponse], error) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	payload := map[string]interface{}{
		"applicationId": applicationID,
		"version":       version,
		"fileName":      fileName,
		"fileSize":      fileSize,
		"contentType":   contentType,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(ctx, "POST", "/v1/management/upload/large-url", nil, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	return parseResponse[LargeUploadURLResponse](resp)
}

// CompleteLargeUpload completes a large file upload.
func (c *AdminClient) CompleteLargeUpload(ctx context.Context, uploadID, applicationID, version, fileName string, fileSize int64, checksum, releaseNotes string) (*Response[UploadResponse], error) {
	payload := map[string]interface{}{
		"uploadId":      uploadID,
		"applicationId": applicationID,
		"version":       version,
		"fileName":      fileName,
		"fileSize":      fileSize,
		"checksum":      checksum,
	}
	if releaseNotes != "" {
		payload["releaseNotes"] = releaseNotes
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(ctx, "POST", "/v1/management/upload/large-complete", nil, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	return parseResponse[UploadResponse](resp)
}

// UploadLargeFile uploads a large file using the multi-step process.
func (c *AdminClient) UploadLargeFile(ctx context.Context, applicationID, version, fileName string, content []byte, releaseNotes string) (*Response[UploadResponse], error) {
	fileSize := int64(len(content))
	hash := sha256.Sum256(content)
	checksum := hex.EncodeToString(hash[:])

	// Step 1: Get upload URL
	urlResp, err := c.GetLargeUploadURL(ctx, applicationID, version, fileName, fileSize, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get upload URL: %w", err)
	}

	if !urlResp.Success {
		if urlResp.Error != nil {
			return nil, fmt.Errorf("API error: %s - %s", urlResp.Error.Code, urlResp.Error.Message)
		}
		return nil, fmt.Errorf("failed to get upload URL")
	}

	// Step 2: Upload to S3
	req, err := http.NewRequestWithContext(ctx, "PUT", urlResp.Data.UploadURL, bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("failed to create upload request: %w", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	s3Resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}
	defer s3Resp.Body.Close()

	if s3Resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("S3 upload failed with status %d", s3Resp.StatusCode)
	}

	// Step 3: Complete upload
	return c.CompleteLargeUpload(ctx, urlResp.Data.UploadID, applicationID, version, fileName, fileSize, checksum, releaseNotes)
}

// UpdateVersion updates version metadata.
func (c *AdminClient) UpdateVersion(ctx context.Context, applicationID, version string, opts UpdateVersionOptions) (*Response[Version], error) {
	body, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/v1/applications/%s/versions/%s",
		url.PathEscape(applicationID), url.PathEscape(version))

	resp, err := c.doRequest(ctx, "PATCH", path, nil, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	return parseResponse[Version](resp)
}

// UpdateCustomer updates customer metadata.
func (c *AdminClient) UpdateCustomer(ctx context.Context, customerID string, name *string, isActive *bool, notes *string) (*Response[Customer], error) {
	payload := map[string]interface{}{}
	if name != nil {
		payload["name"] = *name
	}
	if isActive != nil {
		payload["isActive"] = *isActive
	}
	if notes != nil {
		payload["notes"] = *notes
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/v1/management/customers/%s", url.PathEscape(customerID))
	resp, err := c.doRequest(ctx, "PATCH", path, nil, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	return parseResponse[Customer](resp)
}

// DeleteApplication soft-deletes an application.
func (c *AdminClient) DeleteApplication(ctx context.Context, applicationID string) (*Response[map[string]interface{}], error) {
	path := fmt.Sprintf("/v1/management/applications/%s", url.PathEscape(applicationID))

	resp, err := c.doRequest(ctx, "DELETE", path, nil, nil)
	if err != nil {
		return nil, err
	}

	return parseResponse[map[string]interface{}](resp)
}

// ListActivity lists upload and download activity.
func (c *AdminClient) ListActivity(ctx context.Context, activityType, applicationID string, page, pageSize int) (*Response[[]Activity], error) {
	query := url.Values{}
	if activityType != "" {
		query.Set("type", activityType)
	}
	if applicationID != "" {
		query.Set("applicationId", applicationID)
	}
	if page > 0 {
		query.Set("page", strconv.Itoa(page))
	}
	if pageSize > 0 {
		query.Set("pageSize", strconv.Itoa(pageSize))
	}

	resp, err := c.doRequest(ctx, "GET", "/v1/activity", query, nil)
	if err != nil {
		return nil, err
	}

	result, err := parseResponse[activitiesData](resp)
	if err != nil {
		return nil, err
	}

	return &Response[[]Activity]{
		Success:    result.Success,
		Data:       result.Data.Activities,
		Error:      result.Error,
		Meta:       result.Meta,
		HTTPStatus: result.HTTPStatus,
	}, nil
}

// ListCustomers lists all customers.
func (c *AdminClient) ListCustomers(ctx context.Context, page, pageSize int) (*Response[[]Customer], error) {
	query := url.Values{}
	if page > 0 {
		query.Set("page", strconv.Itoa(page))
	}
	if pageSize > 0 {
		query.Set("pageSize", strconv.Itoa(pageSize))
	}

	resp, err := c.doRequest(ctx, "GET", "/v1/management/customers", query, nil)
	if err != nil {
		return nil, err
	}

	result, err := parseResponse[customersData](resp)
	if err != nil {
		return nil, err
	}

	return &Response[[]Customer]{
		Success:    result.Success,
		Data:       result.Data.Customers,
		Error:      result.Error,
		Meta:       result.Meta,
		HTTPStatus: result.HTTPStatus,
	}, nil
}
