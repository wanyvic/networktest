package main

import (
	"encoding/json"
	"io/ioutil"
)

type Pool struct {
	Pool string `json:"pool"`
	Mode int    `json:"mode"`
}
type Config struct {
	AddrList         []Pool `json:"addr_list"`
	DialTimeout      int    `json:"dial_timeout"`
	ReadWriteTimeout int    `json:"read_write_timeout"`
	PingRate         int    `json:"ping_rate"`
}

func LoadConfigFile(filename string) (config *Config, _ error) {
	config = &Config{}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
