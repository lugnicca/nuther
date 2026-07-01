package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/minio/selfupdate"
)

const DefaultRepo = "lugnicca/nuther"

var ErrNoUpdate = errors.New("no update available")

type Client struct {
	Repo       string
	HTTPClient *http.Client
}

type ReleaseInfo struct {
	CurrentVersion string
	LatestVersion  string
	AssetName      string
	URL            string
	ReleaseURL     string
	Checksum       string
}

func (c Client) Check(ctx context.Context, currentVersion string) (ReleaseInfo, error) {
	repo := c.Repo
	if repo == "" {
		repo = DefaultRepo
	}

	latestTag, err := c.latestReleaseTag(ctx, repo)
	if err != nil {
		return ReleaseInfo{}, err
	}
	latest := normalizeVersion(latestTag)
	current := normalizeVersion(currentVersion)
	releaseURL := fmt.Sprintf("https://github.com/%s/releases/tag/%s", repo, ensureVPrefix(latest))
	if current != "dev" && !isNewer(latest, current) {
		return ReleaseInfo{
			CurrentVersion: currentVersion,
			LatestVersion:  ensureVPrefix(latest),
			ReleaseURL:     releaseURL,
		}, ErrNoUpdate
	}

	assetName := assetNameFor(latest, runtime.GOOS, runtime.GOARCH)
	baseDownloadURL := fmt.Sprintf("https://github.com/%s/releases/download/%s", repo, ensureVPrefix(latest))
	checksums, err := c.downloadString(ctx, baseDownloadURL+"/checksums.txt")
	if err != nil {
		return ReleaseInfo{}, fmt.Errorf("download checksums: %w", err)
	}
	checksum, err := checksumFor(checksums, assetName)
	if err != nil {
		return ReleaseInfo{}, err
	}

	return ReleaseInfo{
		CurrentVersion: currentVersion,
		LatestVersion:  ensureVPrefix(latest),
		AssetName:      assetName,
		URL:            baseDownloadURL + "/" + assetName,
		ReleaseURL:     releaseURL,
		Checksum:       checksum,
	}, nil
}

func (c Client) Apply(ctx context.Context, info ReleaseInfo) error {
	archiveBytes, err := c.downloadBytes(ctx, info.URL)
	if err != nil {
		return fmt.Errorf("download update asset: %w", err)
	}
	if err := verifySHA256(archiveBytes, info.Checksum); err != nil {
		return err
	}

	binary, err := extractBinary(info.AssetName, archiveBytes)
	if err != nil {
		return err
	}
	if len(binary) == 0 {
		return fmt.Errorf("asset %s contains an empty nuther binary", info.AssetName)
	}

	if err := selfupdate.Apply(bytes.NewReader(binary), selfupdate.Options{}); err != nil {
		if rollbackErr := selfupdate.RollbackError(err); rollbackErr != nil {
			return fmt.Errorf("apply update: %w; rollback failed: %v", err, rollbackErr)
		}
		return fmt.Errorf("apply update: %w", err)
	}
	return nil
}

func (c Client) latestReleaseTag(ctx context.Context, repo string) (string, error) {
	url := fmt.Sprintf("https://github.com/%s/releases/latest", repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "nuther-updater")

	client := *c.httpClient()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusMovedPermanently && resp.StatusCode != http.StatusSeeOther && resp.StatusCode != http.StatusTemporaryRedirect && resp.StatusCode != http.StatusPermanentRedirect {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf("GitHub latest release returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	location := resp.Header.Get("Location")
	if location == "" {
		return "", errors.New("GitHub latest release redirect did not include Location header")
	}
	idx := strings.LastIndex(location, "/tag/")
	if idx == -1 {
		return "", fmt.Errorf("GitHub latest release redirect did not point to a tag: %s", location)
	}
	tag := strings.TrimSpace(location[idx+len("/tag/"):])
	if tag == "" {
		return "", fmt.Errorf("GitHub latest release redirect had empty tag: %s", location)
	}
	return tag, nil
}

func (c Client) downloadString(ctx context.Context, url string) (string, error) {
	b, err := c.downloadBytes(ctx, url)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c Client) downloadBytes(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "nuther-updater")
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s returned %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func (c Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: 30 * time.Second}
}

func assetNameFor(version, goos, goarch string) string {
	version = strings.TrimPrefix(version, "v")
	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}
	return fmt.Sprintf("nuther_%s_%s_%s%s", version, goos, goarch, ext)
}

func ensureVPrefix(version string) string {
	version = strings.TrimSpace(version)
	if strings.HasPrefix(version, "v") {
		return version
	}
	return "v" + version
}

func checksumFor(checksums, assetName string) (string, error) {
	for _, line := range strings.Split(checksums, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		filename := strings.TrimPrefix(fields[len(fields)-1], "*")
		if filename == assetName {
			checksum := strings.ToLower(fields[0])
			if len(checksum) != sha256.Size*2 {
				return "", fmt.Errorf("checksum for %s is not SHA256", assetName)
			}
			if _, err := hex.DecodeString(checksum); err != nil {
				return "", fmt.Errorf("checksum for %s is invalid hex: %w", assetName, err)
			}
			return checksum, nil
		}
	}
	return "", fmt.Errorf("checksums.txt has no entry for %s", assetName)
}

func verifySHA256(data []byte, expectedHex string) error {
	expected, err := hex.DecodeString(expectedHex)
	if err != nil {
		return fmt.Errorf("invalid expected checksum: %w", err)
	}
	actual := sha256.Sum256(data)
	if !bytes.Equal(actual[:], expected) {
		return fmt.Errorf("checksum mismatch: got %s, want %s", hex.EncodeToString(actual[:]), expectedHex)
	}
	return nil
}

func extractBinary(assetName string, archiveBytes []byte) ([]byte, error) {
	if strings.HasSuffix(assetName, ".zip") {
		return extractBinaryFromZip(archiveBytes)
	}
	if strings.HasSuffix(assetName, ".tar.gz") {
		return extractBinaryFromTarGz(archiveBytes)
	}
	return nil, fmt.Errorf("unsupported update asset format: %s", assetName)
}

func extractBinaryFromZip(archiveBytes []byte) ([]byte, error) {
	r, err := zip.NewReader(bytes.NewReader(archiveBytes), int64(len(archiveBytes)))
	if err != nil {
		return nil, err
	}
	for _, f := range r.File {
		if isNutherBinary(f.Name) {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}
	return nil, errors.New("update zip does not contain nuther binary")
}

func extractBinaryFromTarGz(archiveBytes []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(archiveBytes))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if header.Typeflag == tar.TypeReg && isNutherBinary(header.Name) {
			return io.ReadAll(tr)
		}
	}
	return nil, errors.New("update tar.gz does not contain nuther binary")
}

func isNutherBinary(name string) bool {
	base := strings.TrimPrefix(name, "./")
	parts := strings.Split(base, "/")
	base = parts[len(parts)-1]
	return base == "nuther" || base == "nuther.exe"
}

func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	if v == "" {
		return "dev"
	}
	return v
}

func isNewer(latest, current string) bool {
	latestParts, latestOK := parseVersion(latest)
	currentParts, currentOK := parseVersion(current)
	if !latestOK || !currentOK {
		return latest != current
	}
	for i := range latestParts {
		if latestParts[i] != currentParts[i] {
			return latestParts[i] > currentParts[i]
		}
	}
	return false
}

func parseVersion(v string) ([3]int, bool) {
	var out [3]int
	core := strings.SplitN(strings.TrimPrefix(v, "v"), "-", 2)[0]
	parts := strings.Split(core, ".")
	if len(parts) == 0 || len(parts) > 3 {
		return out, false
	}
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil {
			return out, false
		}
		out[i] = n
	}
	return out, true
}

func SupportedAssetNames(version string) []string {
	oses := []string{"darwin", "linux", "windows"}
	arches := []string{"amd64", "arm64"}
	var names []string
	for _, goos := range oses {
		for _, goarch := range arches {
			names = append(names, assetNameFor(version, goos, goarch))
		}
	}
	sort.Strings(names)
	return names
}
