package ui

import (
	"errors"
	"testing"
	"time"

	"nuther/internal/smart"
)

func TestDrivesLoadedMsg(t *testing.T) {
	drives := []smart.DriveInfo{
		{Device: "/dev/sda", Model: "Test Drive"},
	}

	msg := DrivesLoadedMsg{
		Drives: drives,
		Error:  nil,
	}

	if len(msg.Drives) != 1 {
		t.Errorf("Drives length = %d, want 1", len(msg.Drives))
	}

	if msg.Error != nil {
		t.Error("Error should be nil")
	}
}

func TestDrivesLoadedMsgWithError(t *testing.T) {
	err := errors.New("scan failed")
	msg := DrivesLoadedMsg{
		Drives: nil,
		Error:  err,
	}

	if msg.Drives != nil {
		t.Error("Drives should be nil on error")
	}

	if msg.Error == nil {
		t.Error("Error should not be nil")
	}

	if msg.Error.Error() != "scan failed" {
		t.Errorf("Error message = %q, want %q", msg.Error.Error(), "scan failed")
	}
}

func TestDriveRefreshedMsg(t *testing.T) {
	drive := smart.DriveInfo{
		Device: "/dev/sda",
		Model:  "Refreshed Drive",
	}

	msg := DriveRefreshedMsg{
		Index: 2,
		Drive: drive,
		Error: nil,
	}

	if msg.Index != 2 {
		t.Errorf("Index = %d, want 2", msg.Index)
	}

	if msg.Drive.Model != "Refreshed Drive" {
		t.Errorf("Drive.Model = %q, want %q", msg.Drive.Model, "Refreshed Drive")
	}

	if msg.Error != nil {
		t.Error("Error should be nil")
	}
}

func TestDriveRefreshedMsgWithError(t *testing.T) {
	err := errors.New("refresh failed")
	msg := DriveRefreshedMsg{
		Index: 0,
		Error: err,
	}

	if msg.Error == nil {
		t.Error("Error should not be nil")
	}
}

func TestScreenshotMsg(t *testing.T) {
	// Success case
	successMsg := ScreenshotMsg{
		Success: true,
		Error:   nil,
		Path:    "/home/user/screenshot.png",
	}

	if !successMsg.Success {
		t.Error("Success should be true")
	}

	if successMsg.Path != "/home/user/screenshot.png" {
		t.Errorf("Path = %q, want %q", successMsg.Path, "/home/user/screenshot.png")
	}

	// Error case
	err := errors.New("screenshot failed")
	errorMsg := ScreenshotMsg{
		Success: false,
		Error:   err,
		Path:    "",
	}

	if errorMsg.Success {
		t.Error("Success should be false")
	}

	if errorMsg.Error == nil {
		t.Error("Error should not be nil")
	}
}

func TestTickMsg(t *testing.T) {
	now := time.Now()
	msg := TickMsg(now)

	// Convert back to time.Time
	asTime := time.Time(msg)
	if !asTime.Equal(now) {
		t.Error("TickMsg should preserve time value")
	}
}
