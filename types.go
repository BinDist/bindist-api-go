package bindist

import "time"

// ApiError represents an error returned by the API.
//
// HTTPStatus is set by the client to the HTTP status code of the
// underlying response when the error was synthesized from a non-2xx
// response that did not conform to the standard error envelope — for
// example, errors raised by auth middleware, reverse proxies, load
// balancers, rate limiters, or gateway timeouts, which never reach the
// API application's own error renderer. In that case Code is derived
// from the HTTP status (e.g. "unauthorized", "rate_limited") and is
// coarser than a code returned by the server's structured error payload.
// HTTPStatus is zero for errors that came from the server envelope.
type ApiError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

// Pagination contains pagination information.
type Pagination struct {
	Page        int  `json:"page"`
	Limit       int  `json:"limit"`
	Total       int  `json:"total"`
	HasNext     bool `json:"hasNext"`
	HasPrevious bool `json:"hasPrevious"`
}

// Meta contains response metadata.
type Meta struct {
	RequestID  string      `json:"requestId"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Application represents an application.
type Application struct {
	ApplicationID string    `json:"applicationId"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	IsActive      bool      `json:"isActive"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	Tags          []string  `json:"tags,omitempty"`
}

// Version represents a version of an application.
type Version struct {
	VersionID     string    `json:"versionId"`
	ApplicationID string    `json:"applicationId"`
	Version       string    `json:"version"`
	ReleaseNotes  string    `json:"releaseNotes,omitempty"`
	IsActive      bool      `json:"isActive"`
	IsEnabled     bool      `json:"isEnabled"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	FileSize      int64     `json:"fileSize"`
	DownloadCount int       `json:"downloadCount"`
}

// VersionFile represents a file within a version.
type VersionFile struct {
	FileID      string `json:"fileId"`
	FileName    string `json:"fileName"`
	FileType    string `json:"fileType"`
	FileSize    int64  `json:"fileSize"`
	Checksum    string `json:"checksum"`
	Order       int    `json:"order"`
	Description string `json:"description,omitempty"`
}

// DownloadInfo contains download URL and file metadata.
type DownloadInfo struct {
	DownloadID string    `json:"downloadId"`
	URL        string    `json:"url"`
	ExpiresAt  time.Time `json:"expiresAt"`
	FileName   string    `json:"fileName"`
	FileSize   int64     `json:"fileSize"`
	Checksum   string    `json:"checksum"`
}

// ShareLink contains share link information.
type ShareLink struct {
	ShareURL  string    `json:"shareUrl"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// Stats contains application statistics.
type Stats struct {
	TotalDownloads int `json:"totalDownloads"`
}

// ListApplicationsOptions contains options for listing applications.
type ListApplicationsOptions struct {
	Page     int
	PageSize int
	Search   string
	Tags     []string
}

// Release channel constants. Channels allow access to versions outside the
// default (production) channel. For example, the "Test" channel exposes
// disabled versions for pre-release testing.
const (
	ChannelTest = "Test"
)

// RequestOption configures a single API request.
type RequestOption func(*requestOptions)

// requestOptions holds per-request configuration applied by RequestOptions.
type requestOptions struct {
	channel string
}

func newRequestOptions(opts []RequestOption) *requestOptions {
	o := &requestOptions{}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// WithChannel sets the release channel for the request. When set, the
// X-Channel header is sent and versions from that channel become accessible
// (e.g. ChannelTest to access disabled/pre-release versions).
func WithChannel(channel string) RequestOption {
	return func(o *requestOptions) {
		o.channel = channel
	}
}

// apiResponse is the generic API response wrapper.
type apiResponse[T any] struct {
	Success bool      `json:"success"`
	Data    T         `json:"data,omitempty"`
	Error   *ApiError `json:"error,omitempty"`
	Meta    *Meta     `json:"meta,omitempty"`
}

// Response is a generic response type.
//
// HTTPStatus is the HTTP status code of the underlying response.
type Response[T any] struct {
	Success    bool
	Data       T
	Error      *ApiError
	Meta       *Meta
	HTTPStatus int
}

// applicationsData wraps the applications array.
type applicationsData struct {
	Applications []Application `json:"applications"`
}

// versionsData wraps the versions array.
type versionsData struct {
	Versions []Version `json:"versions"`
}

// filesData wraps the files array.
type filesData struct {
	Files []VersionFile `json:"files"`
}
