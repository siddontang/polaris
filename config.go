package polaris

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	//polaris listen addr
	HttpAddr string `json:"http_addr"`

	Middlewares []struct {
		Name   string          `json:"name"`
		Config json.RawMessage `json:"config"`
	} `json:"middlewares"`
}

func ParserConfig(configFile string) (*Config, error) {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var c Config

	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return &c, err
}
