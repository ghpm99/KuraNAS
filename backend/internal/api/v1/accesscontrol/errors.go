package accesscontrol

import "errors"

var (
	ErrInvalidCIDR         = errors.New("invalid ip or cidr")
	ErrAllowedIPNotFound   = errors.New("allowed ip not found")
	ErrInvalidAllowedIPID  = errors.New("invalid allowed ip id")
	ErrDuplicateAllowedIP  = errors.New("ip or cidr already registered")
	ErrEmptyAllowedIPInput = errors.New("ip or cidr is required")
)
