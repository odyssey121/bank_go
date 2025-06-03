package util

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
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

// DeSerialize decodes a serialized slice with JSON records
func DeSerializeGinErr(slice *gin.H, r io.Reader) error {
	const op = "utils.DeSerialize"
	e := json.NewDecoder(r)
	err := e.Decode(slice)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
