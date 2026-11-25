package vacuum

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"testing"
)

func encodedPathReport(infoType int, pathID int64, startPos int, positions [][2]int) string {
	positionsJSON := "["
	for index, position := range positions {
		if index > 0 {
			positionsJSON += ","
		}
		positionsJSON += fmt.Sprintf("[%d,%d]", position[0], position[1])
	}
	positionsJSON += "]"
	report := fmt.Sprintf(`{"infoType":%d,"data":{"pathID":%d,"pointCounts":%d,"posArray":%s,"startPos":%d,"totalPoints":%d}}`,
		infoType, pathID, len(positions), positionsJSON, startPos, startPos+len(positions))
	return base64.StdEncoding.EncodeToString([]byte(report))
}

func TestPathTrackerStartsWithoutPath(t *testing.T) {
	snapshot := (&pathTracker{}).snapshot()

	if snapshot.PathID != nil || snapshot.Position != nil || snapshot.Points == nil || len(snapshot.Points) != 0 {
		t.Fatalf("unexpected empty snapshot: %+v", snapshot)
	}
}

func TestPathTrackerDecodesTheReportShapeSeenOnTheRobot(t *testing.T) {
	tracker := &pathTracker{}
	report := base64.StdEncoding.EncodeToString([]byte(
		`{"infoType":21011,"data":{"pathID":1795608351,"pointCounts":3,"posArray":[[-88,23],[-68,67],[-52,111]],"startPos":1,"totalPoints":4}}`))

	tracker.applyReport(report)

	snapshot := tracker.snapshot()
	if snapshot.PathID == nil || *snapshot.PathID != 1795608351 {
		t.Fatalf("unexpected path id: %v", snapshot.PathID)
	}
	if !reflect.DeepEqual(snapshot.Points, [][2]int{{-88, 23}, {-68, 67}, {-52, 111}}) {
		t.Fatalf("unexpected points: %v", snapshot.Points)
	}
	if snapshot.Position == nil || *snapshot.Position != (PointDto{X: -52, Y: 111}) {
		t.Fatalf("unexpected position: %+v", snapshot.Position)
	}
}

func TestPathTrackerPlacesPointsByIndexEvenWhenTheyArriveOutOfOrder(t *testing.T) {
	tracker := &pathTracker{}
	tracker.applyReport(encodedPathReport(21011, 1, 2, [][2]int{{3, 3}, {4, 4}}))
	tracker.applyReport(encodedPathReport(21011, 1, 0, [][2]int{{1, 1}, {2, 2}}))

	if !reflect.DeepEqual(tracker.snapshot().Points, [][2]int{{1, 1}, {2, 2}, {3, 3}, {4, 4}}) {
		t.Fatalf("unexpected points: %v", tracker.snapshot().Points)
	}
}

func TestPathTrackerOverwritesRepeatedPointsInsteadOfAppending(t *testing.T) {
	tracker := &pathTracker{}
	tracker.applyReport(encodedPathReport(21011, 1, 0, [][2]int{{1, 1}, {2, 2}, {3, 3}}))
	tracker.applyReport(encodedPathReport(21011, 1, 1, [][2]int{{2, 2}, {3, 3}, {4, 4}}))
	tracker.applyReport(encodedPathReport(21011, 1, 1, [][2]int{{2, 2}, {3, 3}, {4, 4}}))

	if !reflect.DeepEqual(tracker.snapshot().Points, [][2]int{{1, 1}, {2, 2}, {3, 3}, {4, 4}}) {
		t.Fatalf("unexpected points: %v", tracker.snapshot().Points)
	}
}

func TestPathTrackerSkipsMissingIndexesUntilTheyArrive(t *testing.T) {
	tracker := &pathTracker{}
	tracker.applyReport(encodedPathReport(21011, 1, 0, [][2]int{{1, 1}}))
	tracker.applyReport(encodedPathReport(21011, 1, 3, [][2]int{{4, 4}}))

	if !reflect.DeepEqual(tracker.snapshot().Points, [][2]int{{1, 1}, {4, 4}}) {
		t.Fatalf("unexpected points: %v", tracker.snapshot().Points)
	}
	tracker.applyReport(encodedPathReport(21011, 1, 1, [][2]int{{2, 2}, {3, 3}}))
	if !reflect.DeepEqual(tracker.snapshot().Points, [][2]int{{1, 1}, {2, 2}, {3, 3}, {4, 4}}) {
		t.Fatalf("unexpected points after the gap was filled: %v", tracker.snapshot().Points)
	}
}

func TestPathTrackerStartsOverWhenThePathIDChanges(t *testing.T) {
	tracker := &pathTracker{}
	tracker.applyReport(encodedPathReport(21011, 1, 0, [][2]int{{1, 1}, {2, 2}}))
	tracker.applyReport(encodedPathReport(21011, 2, 0, [][2]int{{9, 9}}))

	snapshot := tracker.snapshot()
	if *snapshot.PathID != 2 || !reflect.DeepEqual(snapshot.Points, [][2]int{{9, 9}}) {
		t.Fatalf("unexpected snapshot: %+v", snapshot)
	}
}

func TestPathTrackerIgnoresOtherReportsAndGarbage(t *testing.T) {
	tracker := &pathTracker{}
	tracker.applyReport(encodedPathReport(21011, 1, 0, [][2]int{{1, 1}}))

	tracker.applyReport(encodedPathReport(20002, 7, 0, [][2]int{{8, 8}}))
	tracker.applyReport("%%% not base64")
	tracker.applyReport(base64.StdEncoding.EncodeToString([]byte("not json")))

	snapshot := tracker.snapshot()
	if *snapshot.PathID != 1 || !reflect.DeepEqual(snapshot.Points, [][2]int{{1, 1}}) {
		t.Fatalf("unexpected snapshot: %+v", snapshot)
	}
}

func TestPathTrackerRejectsIndexesOutsideTheSaneRange(t *testing.T) {
	tracker := &pathTracker{}
	tracker.applyReport(encodedPathReport(21011, 1, -1, [][2]int{{1, 1}}))
	tracker.applyReport(encodedPathReport(21011, 1, maxPathPoints, [][2]int{{2, 2}}))
	malformed := base64.StdEncoding.EncodeToString([]byte(`{"infoType":21011,"data":{"pathID":1,"posArray":[[1],[5,6]],"startPos":0}}`))
	tracker.applyReport(malformed)

	if !reflect.DeepEqual(tracker.snapshot().Points, [][2]int{{5, 6}}) {
		t.Fatalf("unexpected points: %v", tracker.snapshot().Points)
	}
}

func TestPathTrackerResetForgetsThePath(t *testing.T) {
	tracker := &pathTracker{}
	tracker.applyReport(encodedPathReport(21011, 1, 0, [][2]int{{1, 1}}))

	tracker.reset()

	snapshot := tracker.snapshot()
	if snapshot.PathID != nil || len(snapshot.Points) != 0 || snapshot.Position != nil {
		t.Fatalf("unexpected snapshot: %+v", snapshot)
	}
}

func TestServiceFeedsThePathFromPushedDataPoints(t *testing.T) {
	fixture := newServiceFixture(t)

	fixture.service.acceptPush(map[string]any{dpPath: encodedPathReport(21011, 5, 0, [][2]int{{10, 20}})})
	fixture.service.acceptPush(map[string]any{dpBatteryPercent: 80.0})
	fixture.service.acceptPush(map[string]any{dpPath: 12.0})

	path := fixture.service.GetPath()
	if *path.PathID != 5 || !reflect.DeepEqual(path.Points, [][2]int{{10, 20}}) {
		t.Fatalf("unexpected path: %+v", path)
	}
}

func TestChangingTheRobotClearsThePathAndNotifiesTheHandler(t *testing.T) {
	fixture := newServiceFixture(t)
	notifications := 0
	fixture.service.SetRobotChangedHandler(func() { notifications++ })
	fixture.service.acceptPush(map[string]any{dpPath: encodedPathReport(21011, 5, 0, [][2]int{{10, 20}})})

	_, err := fixture.service.UpdateSettings(UpdateSettingsDto{Address: testAddress, DeviceID: testDeviceID})
	if err != nil {
		t.Fatal(err)
	}
	if notifications != 0 || len(fixture.service.GetPath().Points) != 1 {
		t.Fatal("saving the same robot must not restart anything")
	}

	if _, err := fixture.service.UpdateSettings(UpdateSettingsDto{Address: "10.9.9.9", DeviceID: testDeviceID}); err != nil {
		t.Fatal(err)
	}
	if notifications != 1 || len(fixture.service.GetPath().Points) != 0 {
		t.Fatalf("a changed robot must reset the path and notify, got %d notifications", notifications)
	}
}
