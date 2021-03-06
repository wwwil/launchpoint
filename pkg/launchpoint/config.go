package launchpoint

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func LoadConfigFromFile(configFilePath string) (*Config, error) {
	// Open, read, unmarshal, and validate the config file.
	configFile, err := os.Open(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open configuration file from: %s", configFilePath)
	}
	defer configFile.Close()
	configFileData, err := ioutil.ReadAll(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file from: %s", configFilePath)
	}
	configFile.Close()
	var config Config
	err = yaml.UnmarshalStrict(configFileData, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration file from: %s", configFilePath)
	}
	if !config.IsValid() {
		return nil, fmt.Errorf("invalid config loaded from: %s", configFilePath)
	}
	return &config, nil
}

// Config is the struct for configuration loaded in from YAML.
type Config struct {
	GPIOTriggers         []GPIOTrigger `yaml:"gpioTriggers"`
	ConsoleInputTriggers []ConsoleInputTrigger `yaml:"consoleTriggers"`
}

func (c Config) GetRequestsForGPIOPin(pin int) []Request {
	for _, trigger := range c.GPIOTriggers {
		if trigger.Pin == pin {
			return trigger.Requests
		}
	}
	return []Request{}
}

func (c Config) GetRequestsForConsoleInputValue(value string) []Request {
	for _, trigger := range c.ConsoleInputTriggers {
		if trigger.Value == value {
			return trigger.Requests
		}
	}
	return []Request{}
}

// TODO: Return more helpful info about why config is not valid.
func (c Config) IsValid() bool {
	for _, trigger := range c.GPIOTriggers {
		if !trigger.IsValid() {
			return false
		}
	}
	for _, trigger := range c.ConsoleInputTriggers {
		if !trigger.IsValid() {
			return false
		}
	}
	return true
}

type GPIOTrigger struct {
	Pin      int
	Requests []Request
}

func (t GPIOTrigger) IsValid() bool {
	if t.Pin > 40 || t.Pin < 0 {
		return false
	}
	for _, request := range t.Requests {
		if !request.IsValid() {
			return false
		}
	}
	return true
}

type ConsoleInputTrigger struct {
	Value    string
	Requests []Request
}

func (t ConsoleInputTrigger) IsValid() bool {
	for _, request := range t.Requests {
		if !request.IsValid() {
			return false
		}
	}
	return true
}

type Request struct {
	Address string
	Method  string
	Data    string
}

func (r Request) Make() error {
	req, err := http.NewRequest(r.Method, r.Address, bytes.NewBuffer([]byte(r.Data)))
	// TODO: Handle and set headers.
	//req.Header.Set("X-Custom-Header", "myvalue")
	//req.Header.Set("Content-Type", "application/json")
	//req, err := http.NewRequest("GET", "http://10.8.4.2:8888", bytes.NewBuffer([]byte{}))

	//resp, err := http.Get( "http://10.8.4.2:8888")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the status.
	if resp.StatusCode != 200 {
		return fmt.Errorf("status: %s", resp.Status)
	}

	// TODO: Add body checking.
	// Check the body.
	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	return err
	//}

	return nil
}

func (r Request) IsValid() bool {
	if !(strings.HasPrefix(r.Address, "http://") || strings.HasPrefix(r.Address, "https://")) {
		return false
	}
	if !(r.Method == "GET" || r.Method == "POST") {
		return false
	}
	return true
}
