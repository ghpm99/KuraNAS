package vacuum

import (
	"context"
	"reflect"
	"testing"
	"time"

	"nas-go/api/pkg/tuya"
)

func TestStateMappingCoversEveryDeviceValue(t *testing.T) {
	expectedStates := map[string]string{
		"idle":              "idle",
		"sweep":             "cleaning",
		"mop":               "cleaning",
		"totaling":          "cleaning",
		"pointing":          "cleaning",
		"areaing":           "cleaning",
		"curpointing":       "cleaning",
		"selectroom":        "cleaning",
		"pause":             "paused",
		"tocharge":          "returning",
		"chargring":         "charging",
		"fullcharge":        "charged",
		"dustcenterworking": "emptying",
		"dormant":           "sleeping",
		"remotectl":         "remote_control",
		"fault":             "fault",
	}
	for deviceValue, wantState := range expectedStates {
		t.Run(deviceValue, func(t *testing.T) {
			fixture := newServiceFixture(t)
			fixture.client.readFn = func(tuya.Device) (map[string]any, error) {
				return map[string]any{dpState: deviceValue}, nil
			}

			status, err := fixture.service.GetStatus(context.Background())
			if err != nil || status.State != wantState {
				t.Fatalf("got %q (%v), want %q", status.State, err, wantState)
			}
		})
	}
}

func TestUnmappedOrMissingStateIsUnknown(t *testing.T) {
	for name, dataPoints := range map[string]map[string]any{
		"unmapped": {dpState: "brand_new_state"},
		"missing":  {},
		"not text": {dpState: float64(3)},
	} {
		t.Run(name, func(t *testing.T) {
			fixture := newServiceFixture(t)
			fixture.client.readFn = func(tuya.Device) (map[string]any, error) { return dataPoints, nil }

			status, err := fixture.service.GetStatus(context.Background())
			if err != nil || status.State != "unknown" {
				t.Fatalf("got %q (%v)", status.State, err)
			}
		})
	}
}

func TestStatusFieldsAndFaultCodes(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.client.readFn = func(tuya.Device) (map[string]any, error) {
		return map[string]any{
			dpState:           "sweep",
			dpBatteryPercent:  float64(87),
			dpCleaningSeconds: float64(1200),
			dpCleaningArea:    float64(14),
			dpFaultBitmap:     float64(1 | 1<<2 | 1<<29),
		}, nil
	}

	status, err := fixture.service.GetStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := StatusDto{
		State:           "cleaning",
		BatteryPercent:  87,
		CleaningSeconds: 1200,
		CleaningArea:    14,
		FaultCodes:      []int{1, 3, 30},
		ReadAt:          fixture.clock.instant,
	}
	if !reflect.DeepEqual(status, want) {
		t.Fatalf("got %+v, want %+v", status, want)
	}
}

func TestMissingFaultBitmapGivesEmptyList(t *testing.T) {
	fixture := newServiceFixture(t)

	status, err := fixture.service.GetStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if status.FaultCodes == nil || len(status.FaultCodes) != 0 {
		t.Fatalf("expected an empty non-nil list, got %#v", status.FaultCodes)
	}
}

func TestStatusReadAtIsWhenTheSnapshotWasRead(t *testing.T) {
	fixture := newServiceFixture(t)
	first, _ := fixture.service.GetStatus(context.Background())
	fixture.clock.advance(time.Second)
	second, _ := fixture.service.GetStatus(context.Background())

	if !first.ReadAt.Equal(second.ReadAt) {
		t.Fatalf("reused snapshot must keep its read instant: %v vs %v", first.ReadAt, second.ReadAt)
	}
}

func TestStatusPropagatesDeviceFailure(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.repository.isFound = false

	if _, err := fixture.service.GetStatus(context.Background()); err == nil {
		t.Fatal("expected an error")
	}
	if _, err := fixture.service.GetPreferences(context.Background()); err == nil {
		t.Fatal("expected an error")
	}
	if _, err := fixture.service.GetConsumables(context.Background()); err == nil {
		t.Fatal("expected an error")
	}
	if _, err := fixture.service.GetTotals(context.Background()); err == nil {
		t.Fatal("expected an error")
	}
}

func TestTotals(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.client.readFn = func(tuya.Device) (map[string]any, error) {
		return map[string]any{
			dpTotalCleaningSeconds: float64(36000),
			dpTotalCleaningArea:    float64(420),
			dpTotalCleaningCount:   float64(17),
		}, nil
	}

	totals, err := fixture.service.GetTotals(context.Background())
	want := TotalsDto{CleaningSeconds: 36000, CleaningArea: 420, CleaningCount: 17}
	if err != nil || totals != want {
		t.Fatalf("got %+v (%v), want %+v", totals, err, want)
	}
}

func TestIntegerDataPointAcceptsIntegers(t *testing.T) {
	if reported, isPresent := integerDataPoint(map[string]any{"1": 7}, "1"); !isPresent || reported != 7 {
		t.Fatalf("got %d %v", reported, isPresent)
	}
}
