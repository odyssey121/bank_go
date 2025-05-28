package util

import (
	"encoding/json"
	"fmt"
	"io"
)

// Serialize serializes a slice with JSON records
func Serialize(slice interface{}, w io.Writer) error {
	const op = "utils.Serialize"
	e := json.NewEncoder(w)
	e.SetIndent("", "\t")
	err := e.Encode(slice)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
