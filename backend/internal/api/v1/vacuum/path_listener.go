package vacuum

import (
	"context"
	"fmt"
	"sync"
	"time"

	"nas-go/api/pkg/applog"
)

const (
	initialListenBackoff = 5 * time.Second
	maxListenBackoff     = 60 * time.Second
)

// PathListener keeps a long-lived connection to the robot so the cleaning path pushed by the device
// reaches the service. It reconnects with backoff and rereads the stored robot settings on every
// attempt.
type PathListener struct {
	service        *Service
	client         PushListenerClient
	initialBackoff time.Duration
	maxBackoff     time.Duration
	wake           chan struct{}

	mutex         sync.Mutex
	cancelAttempt context.CancelFunc
	stopListening context.CancelFunc
	finished      chan struct{}
}

// NewPathListener creates a PathListener that feeds service with the pushes received through client.
func NewPathListener(service *Service, client PushListenerClient) *PathListener {
	return &PathListener{
		service:        service,
		client:         client,
		initialBackoff: initialListenBackoff,
		maxBackoff:     maxListenBackoff,
		wake:           make(chan struct{}, 1),
	}
}

// Start launches the listening loop. It does nothing when the loop is already running.
func (l *PathListener) Start() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l.stopListening != nil {
		return
	}

	loopContext, stopListening := context.WithCancel(context.Background())
	l.stopListening = stopListening
	l.finished = make(chan struct{})
	go l.run(loopContext, l.finished)
}

// Stop ends the listening loop and waits until it has returned.
func (l *PathListener) Stop() {
	l.mutex.Lock()
	stopListening, finished := l.stopListening, l.finished
	l.stopListening, l.finished = nil, nil
	l.mutex.Unlock()

	if stopListening == nil {
		return
	}
	stopListening()
	<-finished
}

// Restart drops the current connection and reconnects right away, rereading the robot settings.
func (l *PathListener) Restart() {
	l.mutex.Lock()
	if l.cancelAttempt != nil {
		l.cancelAttempt()
	}
	l.mutex.Unlock()

	select {
	case l.wake <- struct{}{}:
	default:
	}
}

func (l *PathListener) run(loopContext context.Context, finished chan<- struct{}) {
	defer close(finished)

	backoff := l.initialBackoff
	for loopContext.Err() == nil {
		attemptDuration := l.listenOnce(loopContext)
		if attemptDuration >= l.maxBackoff {
			backoff = l.initialBackoff
		}
		if !l.waitBeforeRetry(loopContext, backoff) {
			return
		}
		backoff = min(backoff*2, l.maxBackoff)
	}
}

func (l *PathListener) listenOnce(loopContext context.Context) time.Duration {
	attemptContext, cancelAttempt := context.WithCancel(loopContext)
	defer cancelAttempt()
	l.mutex.Lock()
	l.cancelAttempt = cancelAttempt
	l.mutex.Unlock()

	device, canListen, err := l.service.deviceToListenTo()
	if err != nil {
		applog.Warn("vacuum: cannot read the robot settings to listen for the path", "error", err.Error())
		return 0
	}
	if !canListen {
		return 0
	}

	startedAt := time.Now()
	listenErr := l.client.Listen(attemptContext, device, l.service.acceptPush)
	if attemptContext.Err() == nil {
		applog.Warn("vacuum: robot connection for the path ended", "error", fmt.Sprint(listenErr))
	}
	return time.Since(startedAt)
}

func (l *PathListener) waitBeforeRetry(loopContext context.Context, backoff time.Duration) bool {
	timer := time.NewTimer(backoff)
	defer timer.Stop()

	select {
	case <-loopContext.Done():
		return false
	case <-l.wake:
		return true
	case <-timer.C:
		return true
	}
}
