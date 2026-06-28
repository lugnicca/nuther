package smartwatch

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestServerSnapshotsRoutes(t *testing.T) {
	store := NewStore(t.TempDir())
	record, err := store.SaveSnapshot(time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC), ReasonManual, testDrive("/dev/sda", "SN123"))
	if err != nil {
		t.Fatalf("SaveSnapshot() error = %v", err)
	}

	handler := NewServer(store, NewEventHub("")).Handler()

	indexResp := httptest.NewRecorder()
	handler.ServeHTTP(indexResp, httptest.NewRequest(http.MethodGet, "/snapshots", nil))
	if indexResp.Code != http.StatusOK {
		t.Fatalf("/snapshots status = %d, want 200", indexResp.Code)
	}
	var index Index
	if err := json.Unmarshal(indexResp.Body.Bytes(), &index); err != nil {
		t.Fatalf("decode index: %v", err)
	}
	if len(index.Snapshots) != 1 {
		t.Fatalf("index snapshots = %d, want 1", len(index.Snapshots))
	}

	snapshotResp := httptest.NewRecorder()
	handler.ServeHTTP(snapshotResp, httptest.NewRequest(http.MethodGet, "/snapshots/"+record.ID, nil))
	if snapshotResp.Code != http.StatusOK {
		t.Fatalf("/snapshots/{id} status = %d, want 200", snapshotResp.Code)
	}
	var snapshot Snapshot
	if err := json.Unmarshal(snapshotResp.Body.Bytes(), &snapshot); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if snapshot.ID != record.ID {
		t.Fatalf("snapshot id = %q, want %q", snapshot.ID, record.ID)
	}
}

func TestEventHubPostsWebhook(t *testing.T) {
	received := make(chan Event, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		var event Event
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			t.Errorf("decode event: %v", err)
		}
		received <- event
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	hub := NewEventHub(server.URL)
	event := Event{
		Type:       "smart_snapshot",
		Timestamp:  time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC),
		DeviceKey:  "serial-sn123",
		SnapshotID: "snapshot-1",
		Reason:     ReasonManual,
	}
	hub.Publish(t.Context(), event)

	got := <-received
	if got.SnapshotID != event.SnapshotID {
		t.Fatalf("webhook snapshot id = %q, want %q", got.SnapshotID, event.SnapshotID)
	}
}
