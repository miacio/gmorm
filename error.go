package gmorm

import (
	"errors"
	"fmt"
)

func ErrorF(format string, params ...any) error {
	return errors.New(fmt.Sprintf(format, params...))
}
