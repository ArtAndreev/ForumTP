package queries

import (
	"fmt"
)

type UniqueFieldValueAlreadyExistsError struct {
	Model string
	Field string
}

type RecordNotFoundError struct {
	Model  string
	Params string
}

type NoRowsAffectedError struct {
	Model string
	Field string
}

type NullFieldError struct {
	Model string
	Field string
}

func (s UniqueFieldValueAlreadyExistsError) Error() string {
	return fmt.Sprintf("%s error: record with this %s already exists", s.Model, s.Field)
}

func (s RecordNotFoundError) Error() string {
	return fmt.Sprintf(`%s error: record with "%s" not found`, s.Model, s.Params)
}

func (s NoRowsAffectedError) Error() string {
	return fmt.Sprintf(`%s error: no rows affected with parameter "%s"`, s.Model, s.Field)
}

func (s NullFieldError) Error() string {
	return fmt.Sprintf(`%s error: %s is NULL`, s.Model, s.Field)
}
