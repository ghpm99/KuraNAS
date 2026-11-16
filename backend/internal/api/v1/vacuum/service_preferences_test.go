package vacuum

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"nas-go/api/pkg/tuya"
)

func textOf(text string) *string { return &text }
func numberOf(number int) *int   { return &number }
func flagOf(flag bool) *bool     { return &flag }

func TestPreferencesReportEveryDataPoint(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.client.readFn = func(tuya.Device) (map[string]any, error) {
		return map[string]any{
			dpSuction: "strong", dpWaterLevel: "mid", dpVolume: float64(6),
			dpDoNotDisturb: true, dpCarpetBoost: false, dpResumeAfterCharge: true,
			dpAutoEmpty: true, dpAutoEmptyFrequency: "2",
		}, nil
	}

	preferences, err := fixture.service.GetPreferences(context.Background())
	want := PreferencesDto{
		Suction: textOf("strong"), WaterLevel: textOf("mid"), Volume: numberOf(6),
		IsDoNotDisturb: flagOf(true), IsCarpetBoost: flagOf(false), IsResumeAfterCharge: flagOf(true),
		IsAutoEmpty: flagOf(true), AutoEmptyFrequency: textOf("2"),
	}
	if err != nil || !reflect.DeepEqual(preferences, want) {
		t.Fatalf("got %+v (%v)", preferences, err)
	}
}

func TestPreferencesMissingDataPointsAreNull(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.client.readFn = func(tuya.Device) (map[string]any, error) {
		return map[string]any{dpSuction: "auto"}, nil
	}

	preferences, err := fixture.service.GetPreferences(context.Background())
	want := PreferencesDto{Suction: textOf("auto")}
	if err != nil || !reflect.DeepEqual(preferences, want) {
		t.Fatalf("got %+v (%v)", preferences, err)
	}
}

func TestPatchWritesOnlyTheSentDataPointsInOneWrite(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.client.readFn = func(tuya.Device) (map[string]any, error) {
		return map[string]any{dpSuction: "max", dpVolume: float64(3)}, nil
	}

	patch := UpdatePreferencesDto{Suction: textOf("max"), Volume: numberOf(3), IsCarpetBoost: flagOf(true)}
	preferences, err := fixture.service.UpdatePreferences(context.Background(), patch)
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]any{"109": "max", "114": 3, "137": true}
	if len(fixture.client.writes) != 1 || !reflect.DeepEqual(fixture.client.writes[0].dataPoints, want) {
		t.Fatalf("got %+v, want one write of %v", fixture.client.writes, want)
	}
	if preferences.Suction == nil || *preferences.Suction != "max" || preferences.WaterLevel != nil {
		t.Fatalf("response must be the confirmed reading: %+v", preferences)
	}
}

func TestPatchMapsEveryFieldToItsDataPoint(t *testing.T) {
	fixture := newServiceFixture(t)

	patch := UpdatePreferencesDto{
		Suction: textOf("quiet"), WaterLevel: textOf("high"), Volume: numberOf(0),
		IsDoNotDisturb: flagOf(true), IsCarpetBoost: flagOf(false), IsResumeAfterCharge: flagOf(true),
		IsAutoEmpty: flagOf(false), AutoEmptyFrequency: textOf("3"),
	}
	if _, err := fixture.service.UpdatePreferences(context.Background(), patch); err != nil {
		t.Fatal(err)
	}

	want := map[string]any{
		"109": "quiet", "110": "high", "114": 0, "113": true,
		"137": false, "151": true, "135": false, "136": "3",
	}
	if !reflect.DeepEqual(fixture.client.writes[0].dataPoints, want) {
		t.Fatalf("got %v, want %v", fixture.client.writes[0].dataPoints, want)
	}
}

func TestPatchInvalidValuesAreRejectedWithoutWriting(t *testing.T) {
	tests := map[string]UpdatePreferencesDto{
		"suction":        {Suction: textOf("mop")},
		"water level":    {WaterLevel: textOf("ocean")},
		"volume high":    {Volume: numberOf(11)},
		"volume low":     {Volume: numberOf(-1)},
		"frequency":      {AutoEmptyFrequency: textOf("4")},
		"frequency text": {AutoEmptyFrequency: textOf("often")},
	}
	for name, patch := range tests {
		t.Run(name, func(t *testing.T) {
			fixture := newServiceFixture(t)

			_, err := fixture.service.UpdatePreferences(context.Background(), patch)
			if !errors.Is(err, ErrInvalidPreference) || len(fixture.client.writes) != 0 {
				t.Fatalf("got %v with %d writes", err, len(fixture.client.writes))
			}
		})
	}
}

func TestPatchWithoutFieldsIsRejected(t *testing.T) {
	fixture := newServiceFixture(t)

	_, err := fixture.service.UpdatePreferences(context.Background(), UpdatePreferencesDto{})
	if !errors.Is(err, ErrInvalidRequest) || len(fixture.client.writes) != 0 {
		t.Fatalf("got %v with %d writes", err, len(fixture.client.writes))
	}
}

func TestPatchWriteFailureIsReturned(t *testing.T) {
	fixture := newServiceFixture(t)
	fixture.repository.isFound = false

	_, err := fixture.service.UpdatePreferences(context.Background(), UpdatePreferencesDto{Volume: numberOf(1)})
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("got %v", err)
	}
}
