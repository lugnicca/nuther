package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"testing"
)

func TestChecksumForFindsGoReleaserAsset(t *testing.T) {
	checksum := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	got, err := checksumFor(""+
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa  nuther_0.2.2_linux_amd64.tar.gz\n"+
		checksum+"  nuther_0.2.2_windows_amd64.zip\n", "nuther_0.2.2_windows_amd64.zip")
	if err != nil {
		t.Fatalf("checksumFor returned error: %v", err)
	}
	if got != checksum {
		t.Fatalf("checksumFor = %q, want %q", got, checksum)
	}
}

func TestChecksumForRejectsMissingAsset(t *testing.T) {
	_, err := checksumFor("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef  other.zip\n", "nuther.zip")
	if err == nil {
		t.Fatal("checksumFor should reject missing asset")
	}
}

func TestVerifySHA256(t *testing.T) {
	data := []byte("nuther")
	sum := sha256.Sum256(data)
	if err := verifySHA256(data, hex.EncodeToString(sum[:])); err != nil {
		t.Fatalf("verifySHA256 returned error: %v", err)
	}
	if err := verifySHA256(data, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"); err == nil {
		t.Fatal("verifySHA256 should reject checksum mismatch")
	}
}

func TestAssetNameFor(t *testing.T) {
	if got := assetNameFor("v0.2.2", "windows", "amd64"); got != "nuther_0.2.2_windows_amd64.zip" {
		t.Fatalf("windows asset = %q", got)
	}
	if got := assetNameFor("0.2.2", "linux", "arm64"); got != "nuther_0.2.2_linux_arm64.tar.gz" {
		t.Fatalf("linux asset = %q", got)
	}
}

func TestIsNewer(t *testing.T) {
	cases := []struct {
		latest  string
		current string
		want    bool
	}{
		{"0.2.2", "0.2.1", true},
		{"0.2.1", "0.2.1", false},
		{"0.2.0", "0.2.1", false},
		{"v1.0.0", "0.9.9", true},
	}
	for _, tc := range cases {
		if got := isNewer(tc.latest, tc.current); got != tc.want {
			t.Fatalf("isNewer(%q, %q) = %v, want %v", tc.latest, tc.current, got, tc.want)
		}
	}
}

func TestExtractBinaryFromZip(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("nuther.exe")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("binary")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	got, err := extractBinary("nuther_0.2.2_windows_amd64.zip", buf.Bytes())
	if err != nil {
		t.Fatalf("extractBinary returned error: %v", err)
	}
	if string(got) != "binary" {
		t.Fatalf("extractBinary = %q", string(got))
	}
}

func TestExtractBinaryFromTarGz(t *testing.T) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	payload := []byte("binary")
	if err := tw.WriteHeader(&tar.Header{Name: "nuther", Mode: 0755, Size: int64(len(payload))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}

	got, err := extractBinary("nuther_0.2.2_linux_amd64.tar.gz", buf.Bytes())
	if err != nil {
		t.Fatalf("extractBinary returned error: %v", err)
	}
	if string(got) != "binary" {
		t.Fatalf("extractBinary = %q", string(got))
	}
}

func TestExtractBinaryRejectsArchiveWithoutBinary(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("README.md")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.WriteString(w, "nope")
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := extractBinary("nuther_0.2.2_windows_amd64.zip", buf.Bytes()); err == nil {
		t.Fatal("extractBinary should reject archive without nuther binary")
	}
}
