package vacuum

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"nas-go/api/pkg/tuya"
)

func TestResetWritesTheExactValueToDataPoint115(t *testing.T) {
	tests := map[string]string{
		"filter":     "resetfilter",
		"main_brush": "resetmainbrush",
		"side_brush": "resetsidebrush",
	}
	for part, wantValue := range tests {
		t.Run(part, func(t *testing.T) {
			fixture := newServiceFixture(t)

			if err := fixture.service.ResetConsumable(context.Background(), part); err != nil {
				t.Fatal(err)
			}
			want := map[string]any{"115": wantValue}
			if !reflect.DeepEqual(fixture.client.writes[0].dataPoints, want) {
				t.Fatalf("got %v, want %v", fixture.client.writes[0].dataPoints, want)
			}
		})
	}
}

func TestResetUnknownPartIsRejected(t *testing.T) {
	fixture := newServiceFixture(t)

	err := fixture.service.ResetConsumable(context.Background(), "wheel")
	if !errors.Is(err, ErrInvalidPart) || len(fixture.client.writes) != 0 {
		t.Fatalf("got %v with %d writes", err, len(fixture.client.writes))
	}
}

func TestConsumablesListsOnlyReportedPartsInFixedOrder(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.client.readFn = func(tuya.Device) (map[string]any, error) {
		return map[string]any{dpSideBrushSeconds: float64(100), dpFilterSeconds: float64(300), dpMainBrushSeconds: float64(803013)}, nil
	}

	consumables, err := fixture.service.GetConsumables(context.Background())
	want := ConsumablesDto{Items: []ConsumableDto{
		{Part: "filter", Seconds: 300, MaxSeconds: 540000},
		{Part: "main_brush", Seconds: 803013, MaxSeconds: 1080000},
		{Part: "side_brush", Seconds: 100, MaxSeconds: 720000},
	}}
	if err != nil || !reflect.DeepEqual(consumables, want) {
		t.Fatalf("got %+v (%v), want %+v", consumables, err, want)
	}
}

func TestConsumablesWithNothingReportedIsEmptyList(t *testing.T) {
	fixture := newServiceFixture(t)

	consumables, err := fixture.service.GetConsumables(context.Background())
	if err != nil || consumables.Items == nil || len(consumables.Items) != 0 {
		t.Fatalf("got %#v (%v)", consumables, err)
	}
}
