package github

// CreateReleaseRequest represents the payload to create a new GitHub release.
type CreateReleaseRequest struct {
	TagName         string `json:"tag_name"`
	TargetCommitish string `json:"target_commitish,omitempty"`
	Name            string `json:"name,omitempty"`
	Body            string `json:"body,omitempty"`
	Draft           bool   `json:"draft"`
	Prerelease      bool   `json:"prerelease"`
}

// ReleaseResponse represents the response from GitHub after creating a release.
type ReleaseResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	URL  string `json:"html_url"`
}

// AssetResponse represents a single asset attached to a GitHub release.
type AssetResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

//  func main() {
//      token := os.Getenv("GITHUB_TOKEN")
//      client := githubapi.NewClient(token, "octocat", "myrepo")
//
//      releaseReq := CreateReleaseRequest{
//          TagName:    "v1.0.0",
//          Name:       "Version 1.0.0",
//          Body:       "Initial production release",
//          Draft:      false,
//          Prerelease: false,
//      }
//
//      var release ReleaseResponse
//      err := client.PostJSON("/repos/{owner}/{repo}/releases", releaseReq, &release)
//      if err != nil {
//          log.Fatalf("Failed to create release: %v", err)
//      }
//
//      fmt.Printf("Created release: %s (ID: %d)\n", release.URL, release.ID)
//
//      files := []struct {
//          Path        string
//          ContentType string
//          Label       string
//      }{
//          {"dist/myapp-linux-amd64.zip", "application/zip", "Linux 64-bit"},
//          {"dist/myapp-windows-amd64.zip", "application/zip", "Windows 64-bit"},
//      }
//
//      for _, f := range files {
//          file, err := os.Open(f.Path)
//          if err != nil {
//              log.Fatalf("Failed to open file %s: %v", f.Path, err)
//          }
//          defer file.Close()
//
//          meta := githubapi.UploadMeta{
//              ReleaseID: release.ID,
//              Name:      filepath.Base(f.Path),
//              Label:     f.Label,
//          }
//
//          var asset AssetResponse
//          err = client.UploadBinary(meta, f.ContentType, file, &asset)
//          if err != nil {
//              log.Fatalf("Failed to upload asset %s: %v", meta.Name, err)
//          }
//
//          fmt.Printf("Uploaded asset: %s (%s)\n", asset.Name, asset.URL)
//      }
//  }
//
