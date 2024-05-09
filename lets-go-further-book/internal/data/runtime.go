package data

import (
	"fmt"
	"strconv"
)

type Runtime int32

// Deliberately use a value receiver for MarshalJSON instead of pointer receiver
// (e.g. `func (r *Runtime) MarshalJSON() ([]byte, error)`) to get more flexibility:
// This works on both Runtime values and pointers to Runtime values.
func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)
	quotedJSONValue := strconv.Quote(jsonValue)
	return []byte(quotedJSONValue), nil
}
