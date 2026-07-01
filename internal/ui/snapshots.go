package ui

import (
	"fmt"
	"sort"

	"nuther/internal/smartwatch"

	tea "github.com/charmbracelet/bubbletea"
)

// LoadSnapshotsCmd loads the persisted SMART snapshot index from disk.
func LoadSnapshotsCmd(storePath string) tea.Cmd {
	return func() tea.Msg {
		index, err := smartwatch.NewStore(storePath).LoadIndex()
		return SnapshotsLoadedMsg{Index: index, Error: err}
	}
}

// OpenSelectedSnapshotCmd loads the selected archived snapshot from disk.
func OpenSelectedSnapshotCmd(storePath string, index smartwatch.Index, selectedSnapshot int) tea.Cmd {
	return func() tea.Msg {
		records := append([]smartwatch.SnapshotRecord(nil), index.Snapshots...)
		sort.SliceStable(records, func(i, j int) bool {
			return records[i].Timestamp.After(records[j].Timestamp)
		})
		if selectedSnapshot < 0 || selectedSnapshot >= len(records) {
			return SnapshotOpenedMsg{Error: fmt.Errorf("no snapshot selected")}
		}
		snapshot, err := smartwatch.NewStore(storePath).ReadSnapshot(records[selectedSnapshot].ID)
		return SnapshotOpenedMsg{Snapshot: snapshot, Error: err}
	}
}
