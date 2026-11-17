package autoshutdown

import (
	"errors"
	"math"
	"sync"
	"time"
)

const (
	stateIdle    = "idle"
	stateWarning = "warning"
)

// ErrNotInWarningPeriod is returned when a cancel is requested while no
// shutdown warning is counting down.
var ErrNotInWarningPeriod = errors.New("autoshutdown: no shutdown warning in progress")

// Countdown is the shared, goroutine-safe state of the shutdown warning: the
// scheduler starts and finishes it, the HTTP handlers read and cancel it.
type Countdown struct {
	mutex      sync.Mutex
	nowFn      func() time.Time
	isWarning  bool
	shutdownAt time.Time
}

func NewCountdown() *Countdown {
	return &Countdown{nowFn: time.Now}
}

func (c *Countdown) begin(shutdownAt time.Time) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.isWarning = true
	c.shutdownAt = shutdownAt
}

func (c *Countdown) finishIfElapsed(now time.Time) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if !c.isWarning || now.Before(c.shutdownAt) {
		return false
	}
	c.isWarning = false
	return true
}

func (c *Countdown) remaining(now time.Time) time.Duration {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return max(c.shutdownAt.Sub(now), 0)
}

// Status reports the current state of the warning period.
func (c *Countdown) Status() StatusDto {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if !c.isWarning {
		return StatusDto{State: stateIdle}
	}
	secondsRemaining := max(int(math.Ceil(c.shutdownAt.Sub(c.nowFn()).Seconds())), 0)
	shutdownAt := c.shutdownAt
	return StatusDto{State: stateWarning, ShutdownAt: &shutdownAt, SecondsRemaining: &secondsRemaining}
}

// CancelToday aborts the running warning so today's shutdown does not happen,
// returning the new status. It fails with ErrNotInWarningPeriod when idle.
func (c *Countdown) CancelToday() (StatusDto, error) {
	c.mutex.Lock()
	if !c.isWarning {
		c.mutex.Unlock()
		return StatusDto{}, ErrNotInWarningPeriod
	}
	c.isWarning = false
	c.mutex.Unlock()
	return c.Status(), nil
}
