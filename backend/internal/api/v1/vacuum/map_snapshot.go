package vacuum

import (
	"time"

	"nas-go/api/pkg/tuyacloud"
)

const mapLifetime = 30 * time.Second

type mapSnapshot struct {
	layout tuyacloud.MapLayout
	readAt time.Time
	isRead bool
}

func (snapshot mapSnapshot) isFreshAt(instant time.Time) bool {
	return snapshot.isRead && instant.Sub(snapshot.readAt) < mapLifetime
}
