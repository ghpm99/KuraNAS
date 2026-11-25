package tuya

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"
)

const (
	maxFramesPerExchange = 8
	firstSequence        = 1
)

// Client talks the local Tuya protocol version 3.3 over TCP, one connection per call.
type Client struct {
	timeout           time.Duration
	heartbeatInterval time.Duration
	now               func() time.Time
}

// NewClient creates a Client whose timeout applies to the dial and to each read and write.
func NewClient(timeout time.Duration) *Client {
	return &Client{timeout: timeout, heartbeatInterval: defaultHeartbeatInterval, now: time.Now}
}

// ReadDataPoints returns the device data points keyed by their id as a string (for example "101").
// Numbers are decoded as float64. It returns ErrInvalidDevice, ErrUnreachable or ErrKeyRejected
// (wrapped) on failure, or the context error when ctx ends first.
func (client *Client) ReadDataPoints(ctx context.Context, device Device) (map[string]any, error) {
	if err := validateDevice(device); err != nil {
		return nil, err
	}
	request, err := buildReadFrame(device, firstSequence, client.now())
	if err != nil {
		return nil, err
	}
	reply, err := client.exchange(ctx, device, request, commandRead, false)
	if err != nil {
		return nil, err
	}
	plaintext, err := reply.decrypt(device.LocalKey)
	if err != nil {
		return nil, fmt.Errorf("%w: reply cannot be decrypted", ErrKeyRejected)
	}
	return parseDataPoints(plaintext)
}

// WriteDataPoints sets the given data points (keyed by id as a string) on the device and waits for
// its confirmation. It returns ErrInvalidDevice, ErrUnreachable or ErrKeyRejected (wrapped) on
// failure, or the context error when ctx ends first.
func (client *Client) WriteDataPoints(ctx context.Context, device Device, dataPoints map[string]any) error {
	if err := validateDevice(device); err != nil {
		return err
	}
	request, err := buildWriteFrame(device, firstSequence, client.now(), dataPoints)
	if err != nil {
		return err
	}
	confirmation, err := client.exchange(ctx, device, request, commandWrite, true)
	if err != nil {
		return err
	}
	return confirmReadable(confirmation, device.LocalKey)
}

func confirmReadable(confirmation frame, localKey string) error {
	if len(confirmation.ciphertext()) == 0 {
		return nil
	}
	if _, err := confirmation.decrypt(localKey); err != nil {
		return fmt.Errorf("%w: confirmation cannot be decrypted", ErrKeyRejected)
	}
	return nil
}

func (client *Client) exchange(ctx context.Context, device Device, request []byte, expectedCommand uint32, acceptsEmptyReply bool) (frame, error) {
	reply, err := client.converse(ctx, device, request, expectedCommand, acceptsEmptyReply)
	if err != nil && ctx.Err() != nil {
		return frame{}, fmt.Errorf("tuya: %w", ctx.Err())
	}
	return reply, err
}

func (client *Client) converse(ctx context.Context, device Device, request []byte, expectedCommand uint32, acceptsEmptyReply bool) (frame, error) {
	dialer := net.Dialer{Timeout: client.timeout}
	connection, err := dialer.DialContext(ctx, "tcp", normalizeAddress(device.Address))
	if err != nil {
		return frame{}, fmt.Errorf("%w: %v", ErrUnreachable, err)
	}
	defer connection.Close()
	stopWatching := context.AfterFunc(ctx, func() { connection.Close() })
	defer stopWatching()

	if err := connection.SetWriteDeadline(client.now().Add(client.timeout)); err != nil {
		return frame{}, fmt.Errorf("%w: %v", ErrUnreachable, err)
	}
	if _, err := connection.Write(request); err != nil {
		return frame{}, fmt.Errorf("%w: %v", ErrUnreachable, err)
	}
	return client.awaitReply(connection, expectedCommand, acceptsEmptyReply)
}

func (client *Client) awaitReply(connection net.Conn, expectedCommand uint32, acceptsEmptyReply bool) (frame, error) {
	hasReceivedFrame := false
	for attempt := 0; attempt < maxFramesPerExchange; attempt++ {
		if err := connection.SetReadDeadline(client.now().Add(client.timeout)); err != nil {
			return frame{}, fmt.Errorf("%w: %v", ErrUnreachable, err)
		}
		received, err := readFrame(connection)
		if err != nil {
			return frame{}, classifyReadFailure(err, hasReceivedFrame)
		}
		hasReceivedFrame = true
		if isExpectedReply(received, expectedCommand, acceptsEmptyReply) {
			return received, nil
		}
	}
	return frame{}, fmt.Errorf("%w: expected reply never arrived", ErrKeyRejected)
}

func isExpectedReply(received frame, expectedCommand uint32, acceptsEmptyReply bool) bool {
	if received.command != expectedCommand {
		return false
	}
	return acceptsEmptyReply || len(received.ciphertext()) > 0
}

func classifyReadFailure(err error, hasReceivedFrame bool) error {
	if errors.Is(err, errInvalidFrame) {
		return err
	}
	if hasReceivedFrame {
		return fmt.Errorf("%w: expected reply never arrived", ErrKeyRejected)
	}
	return fmt.Errorf("%w: %v", ErrUnreachable, err)
}
