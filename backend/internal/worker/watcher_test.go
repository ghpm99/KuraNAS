package worker

import "testing"

func TestSnapshotsChanged(t *testing.T) {
	base := map[string]fileSnapshot{
		"/a": {ModTimeUnix: 1, Size: 10, IsDir: false},
	}

	if snapshotsChanged(base, map[string]fileSnapshot{"/a": {ModTimeUnix: 1, Size: 10, IsDir: false}}) {
		t.Fatalf("expected equal snapshots to be unchanged")
	}

	if !snapshotsChanged(base, map[string]fileSnapshot{"/a": {ModTimeUnix: 2, Size: 10, IsDir: false}}) {
		t.Fatalf("expected change when metadata differs")
	}

	if !snapshotsChanged(base, map[string]fileSnapshot{"/b": {ModTimeUnix: 1, Size: 10, IsDir: false}}) {
		t.Fatalf("expected change when path differs")
	}

	if !snapshotsChanged(base, map[string]fileSnapshot{"/a": {ModTimeUnix: 1, Size: 10, IsDir: false}, "/b": {ModTimeUnix: 1, Size: 1, IsDir: false}}) {
		t.Fatalf("expected change when snapshot sizes differ")
	}
}
