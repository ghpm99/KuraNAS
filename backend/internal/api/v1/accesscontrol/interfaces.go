package accesscontrol

import (
	"database/sql"
	"net/netip"

	"nas-go/api/pkg/database"
)

type RepositoryInterface interface {
	GetDbContext() *database.DbContext
	GetAll() ([]AllowedIPModel, error)
	GetByID(id int) (AllowedIPModel, error)
	Create(tx *sql.Tx, model AllowedIPModel) (AllowedIPModel, error)
	Update(tx *sql.Tx, model AllowedIPModel) (AllowedIPModel, error)
	Delete(tx *sql.Tx, id int) error
}

type ServiceInterface interface {
	GetAllowedIPs() ([]AllowedIPDto, error)
	CreateAllowedIP(dto CreateAllowedIPDto) (AllowedIPDto, error)
	UpdateAllowedIP(id int, dto UpdateAllowedIPDto) (AllowedIPDto, error)
	DeleteAllowedIP(id int) error
	// IsAllowed reports whether addr matches an enabled whitelist entry.
	// Loopback is handled by the middleware, not here.
	IsAllowed(addr netip.Addr) bool
	// Reload rebuilds the in-memory prefix cache from the database. It is
	// called once at boot and after every mutation.
	Reload() error
}
