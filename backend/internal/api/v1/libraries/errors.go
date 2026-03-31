package libraries

import "errors"

var (
	ErrInvalidCategory  = errors.New("invalid library category")
	ErrPathNotSubfolder = errors.New("path must be a subfolder of the entry point")
	ErrPathNotExists    = errors.New("the specified path does not exist")
)
