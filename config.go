package main

import (
	"fmt"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	Source     string      `json:"source"`
	Output     string      `json:"output"`
	PostDir    string      `json:"PostDir"`
	Template   string      `json:"template"`
	Static     string      `json:"static"`
	Properties interface{} `json:"properties"`
}

func LoadDefaults(config *Config) {
	if config.Source == "" {
		config.Source = "content"
	}
	if config.Output == "" {
		config.Output = "output"
	}
	if config.PostDir == "" {
		config.PostDir = "post"
	}
	if config.Template == "" {
		config.Template = "template"
	}
	if config.Static == "" {
		config.Static = "static"
	}
}

func LoadConfig(configFile string) *Config {
	var config Config

	read := true
	_, err := os.Stat(configFile)
	if os.IsNotExist(err) {
		config = Config{}
		read = false
	}

	if read {
		content, err0 := ioutil.ReadFile(configFile)
		if err0 != nil {
			panic(err0)
		}

		err1 := yaml.Unmarshal(content, &config)
		if err1 != nil {
			panic(err1)
		}
		log.Println(config)
	}

	LoadDefaults(&config)

	_, serr := os.Stat(config.Source)
	if serr != nil {
		panic(serr)
	}

	me := os.MkdirAll(fmt.Sprintf("%s/%s", config.Output, config.PostDir), 0755)
	if me != nil {
		panic(me)
	}

	return &config
}
