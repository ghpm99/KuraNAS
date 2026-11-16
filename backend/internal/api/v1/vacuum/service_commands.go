package vacuum

import (
	"context"
	"fmt"
)

func (s *Service) SendCommand(ctx context.Context, command string) error {
	write, isKnown := commandWrites[command]
	if !isKnown {
		return fmt.Errorf("%w: %q", ErrInvalidCommand, command)
	}
	return s.writeDataPoints(ctx, write.asPayload())
}

func (s *Service) SendRemoteControl(ctx context.Context, direction string) error {
	if !remoteControlDirections[direction] {
		return fmt.Errorf("%w: %q", ErrInvalidDirection, direction)
	}
	return s.writeDataPoints(ctx, dataPointWrite{dpRemoteControl, direction}.asPayload())
}
