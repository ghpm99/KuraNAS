package accesscontrol

import "time"

type AllowedIPDto struct {
	ID        int       `json:"id"`
	CIDR      string    `json:"cidr"`
	Label     string    `json:"label,omitempty"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateAllowedIPDto struct {
	CIDR  string `json:"cidr" binding:"required"`
	Label string `json:"label"`
}

type UpdateAllowedIPDto struct {
	CIDR    *string `json:"cidr,omitempty"`
	Label   *string `json:"label,omitempty"`
	Enabled *bool   `json:"enabled,omitempty"`
}

func (m *AllowedIPModel) ToDto() AllowedIPDto {
	return AllowedIPDto{
		ID:        m.ID,
		CIDR:      m.CIDR,
		Label:     m.Label,
		Enabled:   m.Enabled,
		CreatedAt: m.CreatedAt,
	}
}
