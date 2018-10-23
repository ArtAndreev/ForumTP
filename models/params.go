package models

import (
	"time"
)

type ThreadQueryParams struct {
	Desc  bool
	Limit int
	Since time.Time
}

type UserQueryParams struct {
	Desc  bool
	Limit int
	Since string
}

type PostQueryArgs struct {
	Related []string
}
