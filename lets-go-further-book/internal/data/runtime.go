package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

type Runtime int32

// Deliberately use a value receiver for MarshalJSON instead of pointer receiver
// (e.g. `func (r *Runtime) MarshalJSON() ([]byte, error)`) to get more flexibility:
// This works on both Runtime values and pointers to Runtime values.
func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)
	quotedJSONValue := strconv.Quote(jsonValue)
	return []byte(quotedJSONValue), nil
}

// Use pointer because Unmarshal needs to modify the value of the receiver.
// Don't want to be updating a copy.
func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return err
	}

	parts := strings.Split(unquotedJSONValue, " ")
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	i, err := strconv.Atoi(parts[0])
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// Use * operator to dereference the pointer to the receiver, to set the value.
	*r = Runtime(i)
	return nil
}
