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
