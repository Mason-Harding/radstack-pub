package config

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

var (
	configFileName = "~/.radstack/config"
)

type Config struct {
	valCache           map[string]*string
	configFileVals     map[string]*string
	configFileIsLoaded bool
}

func NewConfig() *Config {
	c := new(Config)
	c.valCache = map[string]*string{}
	c.configFileVals = map[string]*string{}
	return c
}

func (c *Config) GetValue(name string) *string {
	nameLower := strings.ToLower(name)
	if v := c.valCache[nameLower]; v != nil {
		return v
	}

	if v := valFromEnv(name); v != nil {
		c.valCache[nameLower] = v
		return v
	}

	if v := c.valFromConfigFile(name); v != nil {
		c.valCache[nameLower] = v
		return v
	}
	return nil
}

func valFromEnv(name string) *string {
	v, hasV := os.LookupEnv(name)
	if !hasV {
		return nil
	}
	return &v
}

func (c *Config) valFromConfigFile(name string) *string {
	if !c.configFileIsLoaded && configFileExists() {
		err := c.loadConfigFileVals()
		if err != nil {
			panic(err)
		}
	}

	return c.configFileVals[name]
}

func configFileExists() bool {
	stat, err := os.Stat(configFileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false
		} else {
			panic(err)
		}
	}
	if !(stat.Mode().String() == "-rw-------" || stat.Mode().String() == "-r--------") {
		panic("radstack config file must be set to perm of 0400 or 0600")
	}

	return true
}

func (c *Config) loadConfigFileVals() error {
	if c.configFileIsLoaded {
		return errors.New("radstack config file is already loaded")
	}
	f, err := os.Open(configFileName)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		tokens := strings.Split(scanner.Text(), "=")
		if len(tokens) != 2 {
			return errors.New("radstack config file is not properly formed.  Must be key=value pairs")
		}
		c.configFileVals[tokens[0]] = &tokens[1]
	}

	c.configFileIsLoaded = true
	return nil
}
