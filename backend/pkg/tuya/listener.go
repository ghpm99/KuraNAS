package tuya

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

const (
	defaultHeartbeatInterval = 10 * time.Second
	missedHeartbeatsAllowed  = 3
)

// Listen keeps one connection to the device open and calls onDataPoints with the data points of every
// status push the device sends, keyed by id as a string. It sends a heartbeat at the client heartbeat
// interval and gives up when the device stays silent for several intervals. It returns
// ErrInvalidDevice, ErrUnreachable or ErrKeyRejected (wrapped) when the connection fails, or the
// context error when ctx ends; the caller decides when to listen again.
func (client *Client) Listen(ctx context.Context, device Device, onDataPoints func(dataPoints map[string]any)) error {
	if err := validateDevice(device); err != nil {
		return err
	}
	dialer := net.Dialer{Timeout: client.timeout}
	connection, err := dialer.DialContext(ctx, "tcp", normalizeAddress(device.Address))
	if err != nil {
		return client.listenFailure(ctx, fmt.Errorf("%w: %v", ErrUnreachable, err))
	}
	defer connection.Close()
	stopWatching := context.AfterFunc(ctx, func() { connection.Close() })
	defer stopWatching()

	var heartbeatWorker sync.WaitGroup
	heartbeatWorker.Add(1)
	stopHeartbeats := make(chan struct{})
	go func() {
		defer heartbeatWorker.Done()
		client.sendHeartbeats(connection, stopHeartbeats)
	}()
	defer func() {
		close(stopHeartbeats)
		connection.Close()
		heartbeatWorker.Wait()
	}()

	return client.listenFailure(ctx, client.deliverPushes(connection, device, onDataPoints))
}

func (client *Client) listenFailure(ctx context.Context, failure error) error {
	if ctx.Err() != nil {
		return fmt.Errorf("tuya: %w", ctx.Err())
	}
	return failure
}

func (client *Client) sendHeartbeats(connection net.Conn, stop <-chan struct{}) {
	ticker := time.NewTicker(client.heartbeatInterval)
	defer ticker.Stop()

	sequence := uint32(firstSequence)
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			sequence++
			connection.SetWriteDeadline(client.now().Add(client.timeout))
			if _, err := connection.Write(encodeFrame(sequence, commandHeartbeat, nil)); err != nil {
				connection.Close()
				return
			}
		}
	}
}

func (client *Client) deliverPushes(connection net.Conn, device Device, onDataPoints func(dataPoints map[string]any)) error {
	silenceLimit := client.heartbeatInterval * missedHeartbeatsAllowed
	for {
		if err := connection.SetReadDeadline(client.now().Add(silenceLimit)); err != nil {
			return fmt.Errorf("%w: %v", ErrUnreachable, err)
		}
		received, err := readFrame(connection)
		if err != nil {
			return classifyListenReadFailure(err)
		}
		if received.command != commandStatusPush || len(received.ciphertext()) == 0 {
			continue
		}
		plaintext, err := received.decrypt(device.LocalKey)
		if err != nil {
			return fmt.Errorf("%w: push cannot be decrypted", ErrKeyRejected)
		}
		dataPoints, err := parseDataPoints(plaintext)
		if err != nil {
			continue
		}
		onDataPoints(dataPoints)
	}
}

func classifyListenReadFailure(err error) error {
	if errors.Is(err, errInvalidFrame) {
		return err
	}
	return fmt.Errorf("%w: %v", ErrUnreachable, err)
}
