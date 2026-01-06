package bindist

import "time"

// ApiError represents an error returned by the API.
type ApiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
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

// apiResponse is the generic API response wrapper.
type apiResponse[T any] struct {
	Success bool      `json:"success"`
	Data    T         `json:"data,omitempty"`
	Error   *ApiError `json:"error,omitempty"`
	Meta    *Meta     `json:"meta,omitempty"`
}

// Response is a generic response type.
type Response[T any] struct {
	Success bool
	Data    T
	Error   *ApiError
	Meta    *Meta
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
