package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

const (
	configFile = "config.yaml"
)

type config struct {
	ConsumerKey    string `yaml:"ConsumerKey"`
	ConsumerSecret string `yaml:"ConsumerSecret"`
	AccessToken    string `yaml:"AccessToken"`
	AccessSecret   string `yaml:"AccessSecret"`
	Period         string `yaml:"Period"`
}

func getConfig() *config {

	c := &config{}

	// Read File
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println("[INFO] No yml config, pulling from environment")
		c = &config{
			ConsumerKey:    os.Getenv("CONSUMER_KEY"),
			ConsumerSecret: os.Getenv("CONSUMER_SECRET"),
			AccessToken:    os.Getenv("ACCESS_TOKEN"),
			AccessSecret:   os.Getenv("ACCESS_SECRET"),
			Period:         os.Getenv("PERIOD"),
		}
	} else {
		// Unmarshall config
		err = yaml.Unmarshal(yamlFile, c)
		if err != nil {
			log.Fatalf("[ERROR] %v", err)
		}
	}

	return c
}
