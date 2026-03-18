package utils

import (
	"database/sql"
	"testing"
	"time"
)

func TestOptionalGetSetAndEmpty(t *testing.T) {
	var o Optional[int]
	if !o.IsEmpty() {
		t.Fatalf("expected empty optional")
	}
	if _, err := o.Get(); err == nil {
		t.Fatalf("expected error for empty optional")
	}

	o.Set(10)
	if o.IsEmpty() {
		t.Fatalf("expected non-empty optional after Set")
	}
	v, err := o.Get()
	if err != nil || v != 10 {
		t.Fatalf("unexpected optional value v=%d err=%v", v, err)
	}
}

func TestOptionalParseFromAndToNullTime(t *testing.T) {
	var o Optional[time.Time]

	if err := o.ParseFromNullTime(sql.NullTime{Valid: false}); err != nil {
		t.Fatalf("expected no error for invalid null time, got %v", err)
	}
	if o.HasValue {
		t.Fatalf("expected HasValue false when null time is invalid")
	}

	now := time.Now()
	o.Value = now
	if err := o.ParseFromNullTime(sql.NullTime{Time: now, Valid: true}); err != nil {
		t.Fatalf("expected parse success, got %v", err)
	}
	if !o.HasValue {
		t.Fatalf("expected HasValue true")
	}

	nt, err := o.ParseToNullTime()
	if err != nil || !nt.Valid {
		t.Fatalf("expected valid null time, nt=%+v err=%v", nt, err)
	}

	o.HasValue = false
	nt, err = o.ParseToNullTime()
	if err != nil || nt.Valid {
		t.Fatalf("expected invalid null time when optional empty, nt=%+v err=%v", nt, err)
	}
}

func TestOptionalTimeTypeValidation(t *testing.T) {
	var wrong Optional[int]
	if err := wrong.ParseFromNullTime(sql.NullTime{Time: time.Now(), Valid: true}); err == nil {
		t.Fatalf("expected type error on ParseFromNullTime for non-time optional")
	}

	wrong.Set(1)
	if _, err := wrong.ParseToNullTime(); err == nil {
		t.Fatalf("expected type error on ParseToNullTime for non-time optional")
	}
}
