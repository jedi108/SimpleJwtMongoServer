package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

type Config struct {
	FileName string
	FileData []byte
}

func New() *Config {
	return &Config{}
}

func (c *Config) FileNameFromArgs(args []string) error {
	if len(args) <= 1 {
		return errors.New("args didt have config file name")
	}
	c.FileName = args[1]
	return nil
}

func (c *Config) LoadFromFile() error {
	fileData, err := ioutil.ReadFile(c.FileName)
	if err != nil {
		return err
	}
	c.FileData = fileData
	return err
}

func (cfg *Config) ApplyStruct(structure interface{}) error {
	return json.Unmarshal(cfg.FileData, &structure)
}
