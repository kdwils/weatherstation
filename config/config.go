package config

import "github.com/spf13/viper"

type Config struct {
	Tempest Tempest `json:"tempest" yaml:"tempest"`
}

type Tempest struct {
	Host     string `json:"host" yaml:"host"`
	Scheme   string `json:"scheme" yaml:"scheme"`
	Path     string `json:"path" yaml:"path"`
	Token    string `json:"token" yaml:"token"`
	DeviceID int    `json:"deviceID" yaml:"deviceID"`
}

type ConfigUnmarshaler interface {
	ReadInConfig() error
	Unmarshal(any, ...viper.DecoderConfigOption) error
	ConfigFileUsed() string
}

// New reads a new configuration
func New(cu ConfigUnmarshaler) (Config, error) {
	var c Config

	if cu.ConfigFileUsed() != "" {
		err := cu.ReadInConfig()
		if err != nil {
			return c, err
		}
	}

	err := cu.Unmarshal(&c)
	return c, err
}
