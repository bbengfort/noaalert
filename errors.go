package noaalert

import "errors"

var (
	ErrNoProperties = errors.New("parsed alert contains no properties")
	ErrNoHeadline   = errors.New("parsed alert conains no headline")
)
