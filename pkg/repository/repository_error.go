package repository

import (
	"errors"

	"gorm.io/gorm"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrDuplicatedKey  = errors.New("duplicated key")
)

func transformGormError(err error) error {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return ErrRecordNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return ErrDuplicatedKey
	default:
		return err
	}
}
