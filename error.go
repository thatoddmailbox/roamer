package roamer

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// UndecodedConfigError is reported when there are entries in a config file that could not be understood.
type UndecodedConfigError struct {
	Filename  string
	Undecoded []toml.Key
}

// Error returns a string representation of the UndecodedConfigError.
func (e UndecodedConfigError) Error() string {
	keyString := ""
	for i, key := range e.Undecoded {
		if i != 0 {
			keyString += ", "
		}

		keyString += key.String()
	}

	return fmt.Sprintf(
		"roamer: in %s, the following config key(s) were not recognized: %s",
		e.Filename,
		keyString,
	)
}

// InvalidInputError is reported when invalid input is provided to roamer.
type InvalidInputError struct {
	Input string
}

// Error returns a string representation of the InvalidInputError.
func (e InvalidInputError) Error() string {
	return fmt.Sprintf(
		"roamer: invalid input '%s'",
		e.Input,
	)
}

// OffsetBoundError is reported when an offset is provided to roamer that goes out of bounds.
type OffsetBoundError struct {
	Input string
}

// Error returns a string representation of the OffsetBoundError.
func (e OffsetBoundError) Error() string {
	return fmt.Sprintf(
		"roamer: offset '%s' is out of bounds",
		e.Input,
	)
}
