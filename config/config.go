package config

import (
	"log"
	"os"

	"github.com/ghodss/yaml"
)

const (
	MUTE_GROUP = "MuteGroup"
	TAP_DELAY  = "TapDelay"
)

type Input struct {
	Name    string
	Channel byte
}

type Output struct {
	Name string
	Ip   string
	Port uint
}

type Mapping struct {
	Name    string
	Target  byte
	CC      byte
	ValueOn byte
}

type Config struct {
	Input    Input
	Output   Output
	Mappings []Mapping
}

func ReadConfig(path string) Config {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalln("Config file is not readable: " + path)
	}
	config := Config{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalln("Could not parse config file: " + path)
	}
	return config
}
