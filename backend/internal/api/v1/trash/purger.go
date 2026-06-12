package trash

import (
	"log"
	"time"
)

// Purger enforces the retention policy in the background: once at startup and
// then on every tick, items older than the configured retention are removed
// for good. It is the "agendador" of the trash — no orchestrator job needed,
// purging is cheap and idempotent.
type Purger struct {
	service  ServiceInterface
	interval time.Duration
	stop     chan struct{}
	done     chan struct{}
}

// DefaultPurgeInterval keeps the worst-case overshoot of the retention window
// to half a day.
const DefaultPurgeInterval = 12 * time.Hour

func NewPurger(service ServiceInterface, interval time.Duration) *Purger {
	if interval <= 0 {
		interval = DefaultPurgeInterval
	}
	return &Purger{
		service:  service,
		interval: interval,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
}

func (p *Purger) Start() {
	go p.run()
}

func (p *Purger) Stop() {
	close(p.stop)
	<-p.done
}

func (p *Purger) run() {
	defer close(p.done)

	p.purge()

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.purge()
		case <-p.stop:
			return
		}
	}
}

func (p *Purger) purge() {
	purged, err := p.service.PurgeExpired()
	if err != nil {
		log.Printf("trash: scheduled purge failed: %v", err)
		return
	}
	if purged > 0 {
		log.Printf("trash: scheduled purge removed %d expired item(s)", purged)
	}
}
