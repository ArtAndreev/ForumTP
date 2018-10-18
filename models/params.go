package models

import (
	"time"
)

type ThreadQueryParams struct {
	Desc  bool
	Limit int
	Since time.Time
}
