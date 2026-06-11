package accesscontrol

import (
	"database/sql"
	"errors"
	"net/netip"
	"testing"

	"nas-go/api/pkg/database"
)

type fakeRepository struct {
	rows   []AllowedIPModel
	nextID int
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{nextID: 1}
}

func (r *fakeRepository) GetDbContext() *database.DbContext { return nil }

func (r *fakeRepository) GetAll() ([]AllowedIPModel, error) {
	out := make([]AllowedIPModel, len(r.rows))
	copy(out, r.rows)
	return out, nil
}

func (r *fakeRepository) GetByID(id int) (AllowedIPModel, error) {
	for _, row := range r.rows {
		if row.ID == id {
			return row, nil
		}
	}
	return AllowedIPModel{}, sql.ErrNoRows
}

func (r *fakeRepository) Create(tx *sql.Tx, model AllowedIPModel) (AllowedIPModel, error) {
	model.ID = r.nextID
	r.nextID++
	r.rows = append(r.rows, model)
	return model, nil
}

func (r *fakeRepository) Update(tx *sql.Tx, model AllowedIPModel) (AllowedIPModel, error) {
	for index, row := range r.rows {
		if row.ID == model.ID {
			r.rows[index] = model
			return model, nil
		}
	}
	return AllowedIPModel{}, sql.ErrNoRows
}

func (r *fakeRepository) Delete(tx *sql.Tx, id int) error {
	for index, row := range r.rows {
		if row.ID == id {
			r.rows = append(r.rows[:index], r.rows[index+1:]...)
			return nil
		}
	}
	return sql.ErrNoRows
}

func mustAddr(t *testing.T, value string) netip.Addr {
	t.Helper()
	addr, err := netip.ParseAddr(value)
	if err != nil {
		t.Fatalf("parse addr %q: %v", value, err)
	}
	return addr
}

func TestCreateAllowedIPNormalizesInput(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"192.168.1.10", "192.168.1.10/32"},
		{" 192.168.1.10 ", "192.168.1.10/32"},
		{"::ffff:192.168.1.10", "192.168.1.10/32"},
		{"192.168.1.7/24", "192.168.1.0/24"},
		{"fd00::1", "fd00::1/128"},
	}

	for _, testCase := range cases {
		service := NewService(newFakeRepository())
		created, err := service.CreateAllowedIP(CreateAllowedIPDto{CIDR: testCase.input, Label: " device "})
		if err != nil {
			t.Fatalf("CreateAllowedIP(%q): %v", testCase.input, err)
		}
		if created.CIDR != testCase.expected {
			t.Fatalf("expected %q normalized to %q, got %q", testCase.input, testCase.expected, created.CIDR)
		}
		if created.Label != "device" {
			t.Fatalf("expected trimmed label, got %q", created.Label)
		}
		if !created.Enabled {
			t.Fatalf("new entries must be enabled by default")
		}
	}
}

func TestCreateAllowedIPRejectsInvalidInput(t *testing.T) {
	service := NewService(newFakeRepository())

	if _, err := service.CreateAllowedIP(CreateAllowedIPDto{CIDR: "not-an-ip"}); !errors.Is(err, ErrInvalidCIDR) {
		t.Fatalf("expected ErrInvalidCIDR, got %v", err)
	}
	if _, err := service.CreateAllowedIP(CreateAllowedIPDto{CIDR: "  "}); !errors.Is(err, ErrEmptyAllowedIPInput) {
		t.Fatalf("expected ErrEmptyAllowedIPInput, got %v", err)
	}
}

func TestCreateAllowedIPRejectsDuplicates(t *testing.T) {
	service := NewService(newFakeRepository())

	if _, err := service.CreateAllowedIP(CreateAllowedIPDto{CIDR: "192.168.1.10"}); err != nil {
		t.Fatalf("first create: %v", err)
	}
	// Same address written differently still collides after normalization.
	if _, err := service.CreateAllowedIP(CreateAllowedIPDto{CIDR: "::ffff:192.168.1.10"}); !errors.Is(err, ErrDuplicateAllowedIP) {
		t.Fatalf("expected ErrDuplicateAllowedIP, got %v", err)
	}
}

func TestIsAllowedReflectsCRUDImmediately(t *testing.T) {
	service := NewService(newFakeRepository())
	addr := mustAddr(t, "192.168.1.50")

	if service.IsAllowed(addr) {
		t.Fatalf("empty whitelist must not allow anything")
	}

	created, err := service.CreateAllowedIP(CreateAllowedIPDto{CIDR: "192.168.1.0/24", Label: "lan"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !service.IsAllowed(addr) {
		t.Fatalf("address inside the registered /24 must be allowed")
	}
	// IPv4-mapped IPv6 form of the same client must also match.
	if !service.IsAllowed(mustAddr(t, "::ffff:192.168.1.50")) {
		t.Fatalf("IPv4-mapped address must match the IPv4 entry")
	}
	if service.IsAllowed(mustAddr(t, "192.168.2.50")) {
		t.Fatalf("address outside the registered range must be blocked")
	}

	disabled := false
	if _, err := service.UpdateAllowedIP(created.ID, UpdateAllowedIPDto{Enabled: &disabled}); err != nil {
		t.Fatalf("disable: %v", err)
	}
	if service.IsAllowed(addr) {
		t.Fatalf("disabled entry must not allow access")
	}

	enabled := true
	if _, err := service.UpdateAllowedIP(created.ID, UpdateAllowedIPDto{Enabled: &enabled}); err != nil {
		t.Fatalf("enable: %v", err)
	}
	if !service.IsAllowed(addr) {
		t.Fatalf("re-enabled entry must allow access again")
	}

	if err := service.DeleteAllowedIP(created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if service.IsAllowed(addr) {
		t.Fatalf("deleted entry must not allow access")
	}
}

func TestUpdateAllowedIPValidations(t *testing.T) {
	service := NewService(newFakeRepository())

	if _, err := service.UpdateAllowedIP(0, UpdateAllowedIPDto{}); !errors.Is(err, ErrInvalidAllowedIPID) {
		t.Fatalf("expected ErrInvalidAllowedIPID, got %v", err)
	}
	if _, err := service.UpdateAllowedIP(99, UpdateAllowedIPDto{}); !errors.Is(err, ErrAllowedIPNotFound) {
		t.Fatalf("expected ErrAllowedIPNotFound, got %v", err)
	}

	created, err := service.CreateAllowedIP(CreateAllowedIPDto{CIDR: "10.0.0.1"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	badCIDR := "bogus"
	if _, err := service.UpdateAllowedIP(created.ID, UpdateAllowedIPDto{CIDR: &badCIDR}); !errors.Is(err, ErrInvalidCIDR) {
		t.Fatalf("expected ErrInvalidCIDR, got %v", err)
	}

	other, err := service.CreateAllowedIP(CreateAllowedIPDto{CIDR: "10.0.0.2"})
	if err != nil {
		t.Fatalf("create other: %v", err)
	}
	colliding := "10.0.0.1"
	if _, err := service.UpdateAllowedIP(other.ID, UpdateAllowedIPDto{CIDR: &colliding}); !errors.Is(err, ErrDuplicateAllowedIP) {
		t.Fatalf("expected ErrDuplicateAllowedIP, got %v", err)
	}

	if err := service.DeleteAllowedIP(0); !errors.Is(err, ErrInvalidAllowedIPID) {
		t.Fatalf("expected ErrInvalidAllowedIPID on delete, got %v", err)
	}
	if err := service.DeleteAllowedIP(99); !errors.Is(err, ErrAllowedIPNotFound) {
		t.Fatalf("expected ErrAllowedIPNotFound on delete, got %v", err)
	}
}

func TestGetAllowedIPsListsEverything(t *testing.T) {
	service := NewService(newFakeRepository())
	if _, err := service.CreateAllowedIP(CreateAllowedIPDto{CIDR: "10.0.0.1", Label: "a"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := service.CreateAllowedIP(CreateAllowedIPDto{CIDR: "10.0.0.2", Label: "b"}); err != nil {
		t.Fatalf("create: %v", err)
	}

	dtos, err := service.GetAllowedIPs()
	if err != nil {
		t.Fatalf("GetAllowedIPs: %v", err)
	}
	if len(dtos) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(dtos))
	}
}
