package config

import (
	"encoding/json"
	"fmt"
)

type Tag string

type Configurator struct {
	Description string                 `json:"description"`
	Kind        string                 `json:"kind"`
	Instance    map[Tag]ConfigInstance `json:"instances"`
}

type ConfigInstance interface {
	InstanceKind() string
}

type Config struct {
	Data map[string]*Configurator `json:"config"`
}

func (c *Config) UnmarshalJSON(data []byte) error {
	conf := struct {
		Data map[string]struct {
			Description string                  `json:"description"`
			Kind        string                  `json:"kind"`
			Instance    map[Tag]json.RawMessage `json:"instances"`
		} `json:"config"`
	}{}

	if err := json.Unmarshal(data, &conf); err != nil {
		return err
	}
	if c.Data == nil {
		c.Data = make(map[string]*Configurator, len(conf.Data))
	}

	for nm, payload := range conf.Data {
		_conf := &Configurator{
			Description: payload.Description,
			Kind:        payload.Kind,
			Instance:    make(map[Tag]ConfigInstance, len(payload.Instance)),
		}

		for tag, instance := range payload.Instance {
			var srv Service
			var db DatabaseConfig
			var n Nats
			var f File
			var tb TelegramBotConfig

			switch payload.Kind {
			case (&srv).InstanceKind():
				if err := json.Unmarshal(instance, &srv); err != nil {
					return err
				}

				_conf.Instance[tag] = &srv

			case (&db).InstanceKind():
				if err := json.Unmarshal(instance, &db); err != nil {
					return err
				}

				_conf.Instance[tag] = &db
			case (&n).InstanceKind():
				if err := json.Unmarshal(instance, &n); err != nil {
					return err
				}

				_conf.Instance[tag] = &n
			case (&f).InstanceKind():
				if err := json.Unmarshal(instance, &f); err != nil {
					return err
				}

				_conf.Instance[tag] = &f
			case (&tb).InstanceKind():
				if err := json.Unmarshal(instance, &tb); err != nil {
					return err
				}

				_conf.Instance[tag] = &tb
			default:
				return fmt.Errorf("Wrong instance kind %s", payload.Kind)
			}
		}

		c.Data[nm] = _conf
	}

	return nil
}

func (c *Config) GetConfigurator(configName string) *Configurator {
	return c.Data[configName]
}

func (cor *Configurator) GetInstance(tag Tag) ConfigInstance {
	if cor == nil {
		return nil
	}

	return cor.Instance[tag]
}
