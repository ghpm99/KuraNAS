package accesscontrol

import (
	"database/sql"
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"sync/atomic"

	"nas-go/api/pkg/database"
)

type Service struct {
	Repository RepositoryInterface

	// allowedPrefixes caches the enabled entries as parsed prefixes. The list
	// is read on every request by the middleware and changes rarely, so it is
	// rebuilt only at boot and after each CRUD mutation — never per request.
	allowedPrefixes atomic.Value // []netip.Prefix
}

func NewService(repository RepositoryInterface) ServiceInterface {
	service := &Service{Repository: repository}
	service.allowedPrefixes.Store([]netip.Prefix{})
	if err := service.Reload(); err != nil {
		// An unreadable whitelist must not take the server down: it just
		// means only loopback can get in until the table is reachable.
		fmt.Printf("accesscontrol: initial whitelist load failed: %v\n", err)
	}
	return service
}

func (s *Service) withTransaction(fn func(tx *sql.Tx) error) error {
	return database.ExecOptionalTx(s.Repository.GetDbContext(), fn)
}

// normalizeCIDR validates and canonicalizes user input: a bare IP becomes a
// /32 (IPv4) or /128 (IPv6) prefix, IPv4-mapped IPv6 addresses are unmapped,
// and the prefix is masked to its canonical form (192.168.1.7/24 → 192.168.1.0/24).
func normalizeCIDR(input string) (netip.Prefix, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return netip.Prefix{}, ErrEmptyAllowedIPInput
	}

	if strings.Contains(trimmed, "/") {
		prefix, err := netip.ParsePrefix(trimmed)
		if err != nil {
			return netip.Prefix{}, ErrInvalidCIDR
		}
		prefix, err = prefix.Addr().Unmap().Prefix(prefix.Bits())
		if err != nil {
			return netip.Prefix{}, ErrInvalidCIDR
		}
		return prefix, nil
	}

	addr, err := netip.ParseAddr(trimmed)
	if err != nil {
		return netip.Prefix{}, ErrInvalidCIDR
	}
	addr = addr.Unmap()
	return netip.PrefixFrom(addr, addr.BitLen()), nil
}

func (s *Service) GetAllowedIPs() ([]AllowedIPDto, error) {
	models, err := s.Repository.GetAll()
	if err != nil {
		return nil, fmt.Errorf("GetAllowedIPs: %w", err)
	}

	dtos := make([]AllowedIPDto, 0, len(models))
	for _, model := range models {
		dtos = append(dtos, model.ToDto())
	}
	return dtos, nil
}

func (s *Service) findDuplicate(cidr string, ignoreID int) (bool, error) {
	models, err := s.Repository.GetAll()
	if err != nil {
		return false, err
	}
	for _, model := range models {
		if model.CIDR == cidr && model.ID != ignoreID {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) CreateAllowedIP(dto CreateAllowedIPDto) (AllowedIPDto, error) {
	prefix, err := normalizeCIDR(dto.CIDR)
	if err != nil {
		return AllowedIPDto{}, err
	}

	duplicated, err := s.findDuplicate(prefix.String(), 0)
	if err != nil {
		return AllowedIPDto{}, fmt.Errorf("CreateAllowedIP: %w", err)
	}
	if duplicated {
		return AllowedIPDto{}, ErrDuplicateAllowedIP
	}

	model := AllowedIPModel{
		CIDR:    prefix.String(),
		Label:   strings.TrimSpace(dto.Label),
		Enabled: true,
	}

	var created AllowedIPModel
	err = s.withTransaction(func(tx *sql.Tx) error {
		var createErr error
		created, createErr = s.Repository.Create(tx, model)
		return createErr
	})
	if err != nil {
		return AllowedIPDto{}, fmt.Errorf("CreateAllowedIP: %w", err)
	}

	if reloadErr := s.Reload(); reloadErr != nil {
		return AllowedIPDto{}, fmt.Errorf("CreateAllowedIP reload: %w", reloadErr)
	}
	return created.ToDto(), nil
}

func (s *Service) UpdateAllowedIP(id int, dto UpdateAllowedIPDto) (AllowedIPDto, error) {
	if id <= 0 {
		return AllowedIPDto{}, ErrInvalidAllowedIPID
	}

	existing, err := s.Repository.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AllowedIPDto{}, ErrAllowedIPNotFound
		}
		return AllowedIPDto{}, fmt.Errorf("UpdateAllowedIP get by id: %w", err)
	}

	if dto.CIDR != nil {
		prefix, normErr := normalizeCIDR(*dto.CIDR)
		if normErr != nil {
			return AllowedIPDto{}, normErr
		}
		existing.CIDR = prefix.String()
	}
	if dto.Label != nil {
		existing.Label = strings.TrimSpace(*dto.Label)
	}
	if dto.Enabled != nil {
		existing.Enabled = *dto.Enabled
	}

	duplicated, err := s.findDuplicate(existing.CIDR, id)
	if err != nil {
		return AllowedIPDto{}, fmt.Errorf("UpdateAllowedIP: %w", err)
	}
	if duplicated {
		return AllowedIPDto{}, ErrDuplicateAllowedIP
	}

	var updated AllowedIPModel
	err = s.withTransaction(func(tx *sql.Tx) error {
		var updateErr error
		updated, updateErr = s.Repository.Update(tx, existing)
		return updateErr
	})
	if err != nil {
		return AllowedIPDto{}, fmt.Errorf("UpdateAllowedIP: %w", err)
	}

	if reloadErr := s.Reload(); reloadErr != nil {
		return AllowedIPDto{}, fmt.Errorf("UpdateAllowedIP reload: %w", reloadErr)
	}
	return updated.ToDto(), nil
}

func (s *Service) DeleteAllowedIP(id int) error {
	if id <= 0 {
		return ErrInvalidAllowedIPID
	}

	err := s.withTransaction(func(tx *sql.Tx) error {
		return s.Repository.Delete(tx, id)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrAllowedIPNotFound
		}
		return fmt.Errorf("DeleteAllowedIP: %w", err)
	}

	if reloadErr := s.Reload(); reloadErr != nil {
		return fmt.Errorf("DeleteAllowedIP reload: %w", reloadErr)
	}
	return nil
}

func (s *Service) IsAllowed(addr netip.Addr) bool {
	// Strip any IPv4-mapping and scope id: a zoned address (fe80::1%eth0)
	// never matches a prefix, so the zone must go before Contains.
	addr = addr.Unmap().WithZone("")

	prefixes, _ := s.allowedPrefixes.Load().([]netip.Prefix)
	for _, prefix := range prefixes {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

func (s *Service) Reload() error {
	models, err := s.Repository.GetAll()
	if err != nil {
		return fmt.Errorf("Reload: %w", err)
	}

	prefixes := make([]netip.Prefix, 0, len(models))
	for _, model := range models {
		if !model.Enabled {
			continue
		}
		prefix, parseErr := netip.ParsePrefix(model.CIDR)
		if parseErr != nil {
			// A corrupt row must not block the valid ones.
			fmt.Printf("accesscontrol: skipping invalid stored cidr %q: %v\n", model.CIDR, parseErr)
			continue
		}
		prefixes = append(prefixes, prefix)
	}

	s.allowedPrefixes.Store(prefixes)
	return nil
}
