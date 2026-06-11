package accesscontrol

import "time"

type AllowedIPModel struct {
	ID        int
	CIDR      string
	Label     string
	Enabled   bool
	CreatedAt time.Time
}
