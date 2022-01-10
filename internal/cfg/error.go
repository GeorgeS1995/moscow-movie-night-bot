package cfg

import "fmt"

type ConfigError struct {
	Err error
}

func (ce *ConfigError) Error() string {
	return fmt.Sprintf("Configuration error: %v", ce.Err.Error())
}
