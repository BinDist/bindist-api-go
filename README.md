# BinDist Go API Client

Go client library for the BinDist API.

## Requirements

- Go 1.21+

## Installation

```bash
go get github.com/BinDist/bindist-api-go
```

## Usage

### Customer Client

Use `Client` for end-user operations like listing applications and downloading files.

```go
package main

import (
    "context"
    "fmt"
    "os"

    bindist "github.com/BinDist/bindist-api-go"
)

func main() {
    ctx := context.Background()
    client := bindist.NewClient("https://api.bindist.com", "your-api-key")

    // List applications
    apps, err := client.ListApplications(ctx, nil)
    if err != nil {
        panic(err)
    }
    if apps.Success {
        for _, app := range apps.Data {
            fmt.Printf("%s (%s)\n", app.Name, app.ApplicationID)
        }
    }

    // List applications with filters
    apps, err = client.ListApplications(ctx, &bindist.ListApplicationsOptions{
        Search:   "myapp",
        Tags:     []string{"windows", "desktop"},
        Page:     1,
        PageSize: 10,
    })

    // Get application details
    app, err := client.GetApplication(ctx, "myapp")
    if err != nil {
        panic(err)
    }
    if app.Success {
        fmt.Printf("Name: %s\n", app.Data.Name)
        fmt.Printf("Description: %s\n", app.Data.Description)
    }

    // List versions
    versions, err := client.ListVersions(ctx, "myapp")
    if err != nil {
        panic(err)
    }
    if versions.Success {
        for _, ver := range versions.Data {
            fmt.Printf("%s - %d bytes\n", ver.Version, ver.FileSize)
        }
    }

    // List files in a version
    files, err := client.ListVersionFiles(ctx, "myapp", "1.0.0")
    if err != nil {
        panic(err)
    }
    if files.Success {
        for _, f := range files.Data {
            fmt.Printf("%s (%s) - %d bytes\n", f.FileName, f.FileType, f.FileSize)
        }
    }

    // Get download URL
    download, err := client.GetDownloadInfo(ctx, "myapp", "1.0.0", "")
    if err != nil {
        panic(err)
    }
    if download.Success {
        fmt.Printf("URL: %s\n", download.Data.URL)
        fmt.Printf("Expires: %s\n", download.Data.ExpiresAt)
    }

    // Download file with checksum verification
    content, metadata, err := client.DownloadFile(ctx, "myapp", "1.0.0", "", true)
    if err != nil {
        panic(err)
    }
    err = os.WriteFile(metadata.FileName, content, 0644)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Downloaded %s (%d bytes)\n", metadata.FileName, metadata.FileSize)

    // Download to file directly
    file, err := os.Create("output.exe")
    if err != nil {
        panic(err)
    }
    defer file.Close()
    metadata, err = client.DownloadFileToWriter(ctx, "myapp", "1.0.0", "", file)
    if err != nil {
        panic(err)
    }

    // Create a share link
    share, err := client.CreateShareLink(ctx, "myapp", "1.0.0", "", 60)
    if err != nil {
        panic(err)
    }
    if share.Success {
        fmt.Printf("Share URL: %s\n", share.Data.ShareURL)
    }

    // Get download statistics
    stats, err := client.GetStats(ctx, "myapp")
    if err != nil {
        panic(err)
    }
    if stats.Success {
        fmt.Printf("Total downloads: %d\n", stats.Data.TotalDownloads)
    }
}
```

### Admin Client

Use `AdminClient` for administrative operations like creating applications and uploading files.

```go
package main

import (
    "context"
    "fmt"
    "os"

    bindist "github.com/BinDist/bindist-api-go"
)

func main() {
    ctx := context.Background()
    admin := bindist.NewAdminClient("https://api.bindist.com", "admin-api-key")

    // Create a customer
    customer, err := admin.CreateCustomer(ctx, "Acme Corp", "", "Enterprise customer")
    if err != nil {
        panic(err)
    }
    if customer.Success {
        fmt.Printf("Customer ID: %s\n", customer.Data.CustomerID)
        fmt.Printf("API Key: %s\n", customer.Data.APIKey)
    }

    // Create an application
    app, err := admin.CreateApplication(ctx, bindist.CreateApplicationOptions{
        ApplicationID: "myapp",
        Name:          "My Application",
        CustomerIDs:   []string{"customer-1", "customer-2"},
        Description:   "A great application",
        Tags:          []string{"windows", "desktop"},
    })
    if err != nil {
        panic(err)
    }

    // Upload a small file (< 10MB)
    content, err := os.ReadFile("app.exe")
    if err != nil {
        panic(err)
    }

    result, err := admin.UploadSmallFile(ctx, "myapp", "1.0.0", "app.exe", content, "Initial release")
    if err != nil {
        panic(err)
    }

    // Upload a large file (>= 10MB)
    largeContent, err := os.ReadFile("large-app.exe")
    if err != nil {
        panic(err)
    }

    result, err = admin.UploadLargeFile(ctx, "myapp", "2.0.0", "large-app.exe", largeContent, "Major update")
    if err != nil {
        panic(err)
    }

    // Update version metadata
    isEnabled := true
    releaseNotes := "Updated release notes"
    _, err = admin.UpdateVersion(ctx, "myapp", "1.0.0", bindist.UpdateVersionOptions{
        IsEnabled:    &isEnabled,
        ReleaseNotes: &releaseNotes,
    })
    if err != nil {
        panic(err)
    }

    // Update customer
    newName := "New Name"
    isActive := true
    _, err = admin.UpdateCustomer(ctx, "customer-1", &newName, &isActive, nil)
    if err != nil {
        panic(err)
    }

    // Delete an application (soft delete)
    _, err = admin.DeleteApplication(ctx, "myapp")
    if err != nil {
        panic(err)
    }

    // List activity (uploads and downloads)
    activity, err := admin.ListActivity(ctx, "download", "", 1, 20)
    if err != nil {
        panic(err)
    }
    if activity.Success {
        for _, item := range activity.Data {
            fmt.Printf("%s: %s v%s\n", item.Type, item.ApplicationID, item.Version)
        }
    }

    // List customers
    customers, err := admin.ListCustomers(ctx, 1, 20)
    if err != nil {
        panic(err)
    }
    if customers.Success {
        for _, c := range customers.Data {
            fmt.Printf("%s (%s)\n", c.Name, c.CustomerID)
        }
    }
}
```

### Context Support

All API methods accept a `context.Context` as the first parameter, enabling:

- Request cancellation
- Timeouts
- Deadlines

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

apps, err := client.ListApplications(ctx, nil)
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        log.Println("Request timed out")
    }
}

// With cancellation
ctx, cancel := context.WithCancel(context.Background())
go func() {
    // Cancel after some condition
    cancel()
}()
```

## API Response

All API methods return a `Response[T]` struct:

```go
type Response[T any] struct {
    Success    bool
    Data       T
    Error      *ApiError
    Meta       *Meta
    HTTPStatus int
}

type ApiError struct {
    Code       string
    Message    string
    HTTPStatus int // set for errors synthesized from non-2xx responses
}

type Meta struct {
    RequestID  string
    Pagination *Pagination
}
```

### Error Handling

```go
response, err := client.ListApplications(ctx, nil)
if err != nil {
    // Network or parsing error
    log.Fatalf("Request failed: %v", err)
}

if response.Success {
    // Process response.Data
    apps := response.Data
} else {
    // API error
    fmt.Printf("Error: %s\n", response.Error.Message)
    fmt.Printf("Code: %s\n", response.Error.Code)
}
```

#### Errors from outside the API envelope

Not every error reaches the API's structured error renderer. Responses
from auth middleware, reverse proxies, load balancers, rate limiters, and
gateway timeouts arrive with a plain body (or no body at all) and cannot
be wrapped in the standard `{"success":false,"error":{...}}` envelope.

When the client sees a non-2xx response, it normalizes it into the usual
`Response[T]` shape so consumers only need one code path:

- `response.Success` is forced to `false`.
- `response.HTTPStatus` is set to the underlying HTTP status code.
- `response.Error` is either the server-provided error (with
  `HTTPStatus` filled in) or a synthesized `ApiError` whose `Code` is
  derived from the HTTP status (`unauthorized`, `forbidden`, `not_found`,
  `rate_limited`, `server_error`, `http_error`, ...) and whose `Message`
  is extracted from a bare `{"message":...}`/`{"error":...}` body or
  falls back to `http.StatusText`.

A synthesized error means the failure originated *outside* the API
application itself, so its `Code` is coarser than a code returned by the
server's own error renderer. Switch on `HTTPStatus` if you need to react
to the transport-level status, and on `Error.Code` if you need to react
to specific semantic categories:

```go
response, err := client.ListApplications(ctx, nil)
if err != nil {
    log.Fatalf("Request failed: %v", err)
}
if !response.Success {
    switch response.Error.Code {
    case "unauthorized":
        // Bad/missing API key — often from auth middleware, not the app.
    case "rate_limited":
        // Back off and retry.
    default:
        log.Printf("API error %d: %s", response.HTTPStatus, response.Error.Message)
    }
}
```

### Release Channels

Versions can be published on different release channels. By default, requests
return only versions on the production channel. Pass `WithChannel` to
`ListVersions`, `GetDownloadInfo`, `DownloadFile`, or `DownloadFileToWriter`
to access versions on a non-default channel — for example, the built-in
`ChannelTest` constant exposes disabled/pre-release versions.

```go
// List versions including disabled ones on the Test channel
versions, err := client.ListVersions(ctx, "myapp", bindist.WithChannel(bindist.ChannelTest))

// Get a download URL for a version that is only available on the Test channel
download, err := client.GetDownloadInfo(ctx, "myapp", "1.2.3-rc1", "",
    bindist.WithChannel(bindist.ChannelTest))

// Download directly from the Test channel
content, metadata, err := client.DownloadFile(ctx, "myapp", "1.2.3-rc1", "", true,
    bindist.WithChannel(bindist.ChannelTest))
```

Internally, `WithChannel` sets the `X-Channel` HTTP header on the request.

### Download with Checksum Verification

The `DownloadFile` method can optionally verify SHA256 checksums:

```go
// With verification (recommended)
content, metadata, err := client.DownloadFile(ctx, "myapp", "1.0.0", "", true)
if err != nil {
    // Could be a checksum mismatch error
    log.Fatalf("Download failed: %v", err)
}

// Without verification
content, metadata, err := client.DownloadFile(ctx, "myapp", "1.0.0", "", false)
```

## License

MIT License - see [LICENSE](LICENSE) for details.
