package vacuum

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestCommandsWriteTheExactDataPoint(t *testing.T) {
	tests := []struct {
		command string
		want    map[string]any
	}{
		{"start", map[string]any{"101": true}},
		{"stop", map[string]any{"101": false}},
		{"pause", map[string]any{"102": true}},
		{"resume", map[string]any{"102": false}},
		{"dock", map[string]any{"103": true}},
		{"locate", map[string]any{"112": true}},
	}
	for _, test := range tests {
		t.Run(test.command, func(t *testing.T) {
			fixture := newServiceFixture(t)

			if err := fixture.service.SendCommand(context.Background(), test.command); err != nil {
				t.Fatal(err)
			}
			if len(fixture.client.writes) != 1 || !reflect.DeepEqual(fixture.client.writes[0].dataPoints, test.want) {
				t.Fatalf("got %+v, want %v", fixture.client.writes, test.want)
			}
			if fixture.client.writes[0].device.LocalKey != testLocalKey {
				t.Fatalf("write must use the stored credentials")
			}
		})
	}
}

func TestUnknownCommandIsRejectedWithoutWriting(t *testing.T) {
	fixture := newServiceFixture(t)

	err := fixture.service.SendCommand(context.Background(), "explode")
	if !errors.Is(err, ErrInvalidCommand) || len(fixture.client.writes) != 0 {
		t.Fatalf("got %v with %d writes", err, len(fixture.client.writes))
	}
}

func TestRemoteControlWritesDirectionToDataPoint111(t *testing.T) {
	for _, direction := range []string{"forward", "backward", "left", "right", "stop"} {
		t.Run(direction, func(t *testing.T) {
			fixture := newServiceFixture(t)

			if err := fixture.service.SendRemoteControl(context.Background(), direction); err != nil {
				t.Fatal(err)
			}
			want := map[string]any{"111": direction}
			if !reflect.DeepEqual(fixture.client.writes[0].dataPoints, want) {
				t.Fatalf("got %v, want %v", fixture.client.writes[0].dataPoints, want)
			}
		})
	}
}

func TestUnknownDirectionIsRejected(t *testing.T) {
	fixture := newServiceFixture(t)

	err := fixture.service.SendRemoteControl(context.Background(), "up")
	if !errors.Is(err, ErrInvalidDirection) || len(fixture.client.writes) != 0 {
		t.Fatalf("got %v with %d writes", err, len(fixture.client.writes))
	}
}
