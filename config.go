package cruiseprocessor

import (
	"errors"
	"go.opentelemetry.io/collector/component"
)

var _ component.Config = (*Config)(nil)
var (
	errFiledInvalidByDuration           = errors.New("error validate \"cruise processor\" config: duration illegal")
	errFiledInvalidBySamplingPercentage = errors.New("error validate \"cruise processor\" config: sample percentage must ge zero")
)

type Config struct {
	Duration int64    `mapstructure:"duration"`
	Sample   Sampling `mapstructure:"sample"`
}

// Sampling sample config
// default, we sample fully for errors traces and normal is true
// according to the sampling percentage by tail sample for normal traces
type Sampling struct {
	Percentage float64 `mapstructure:"percentage"`
	Salt       string
	Normal     bool `mapstructure:"normal"`
	All        bool `mapstructure:"all"`
}

func (c *Config) Validate() error {
	if c.Duration < 0 {
		return errFiledInvalidByDuration
	}
	// check sample config
	if err := checkSamplingConfig(c.Sample); err != nil {
		return err
	}

	// TODO: validation for new functionality fields may be added later

	return nil
}

func checkSamplingConfig(s Sampling) error {
	if s.Percentage < 0 {
		return errFiledInvalidBySamplingPercentage
	}

	return nil
}
