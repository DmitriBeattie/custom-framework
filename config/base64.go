package config

import (
	"encoding/base64"
	"encoding/json"
)

func FromBase64(encoded string) (*Config, error) {
	var c Config

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(decoded, &c); err != nil {
		return nil, err
	}

	return &c, nil
}