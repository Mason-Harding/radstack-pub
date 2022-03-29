package config

import (
	"bufio"
	"errors"
	"os"
	"os/user"
	"strings"
)

var (
	usr, _                = user.Current()
	dir                   = usr.HomeDir
	defaultConfigFileName = dir + "/.radstack/config"
)

type Config struct {
	valCache           map[string]*string
	configFileVals     map[string]*string
	configFileIsLoaded bool
	configFileName     string
	onlyUseConfigFile  bool // if true, ignore env vars.  Only used for local environments and testing.
}

func NewConfig() *Config {
	return NewConfigWithFile(defaultConfigFileName, false)
}

func NewConfigWithFile(configFile string, onlyUseFile bool) *Config {
	c := new(Config)
	c.valCache = map[string]*string{}
	c.configFileVals = map[string]*string{}
	c.configFileName = configFile
	c.onlyUseConfigFile = onlyUseFile
	return c
}

// GetValue will look up the value first in its cache, then in env vars, then in the config file.  Will return a nil
// pointer if not found.  The config file supports multi line strings as long as they begin and end with a tick mark `
func (c *Config) GetValue(name string) *string {
	nameLower := strings.ToLower(name)
	if v := c.valCache[nameLower]; v != nil {
		return v
	}

	if v := valFromEnv(name); v != nil && !c.onlyUseConfigFile {
		c.valCache[nameLower] = v
		return v
	}

	if v := c.valFromConfigFile(name); v != nil {
		c.valCache[nameLower] = v
		return v
	}
	return nil
}

// GetValueOrDefault will get a value, or return a default.  It will however not cache the default.
func (c *Config) GetValueOrDefault(name string, defaultValue string) string {
	v := c.GetValue(name)
	if v != nil {
		return *v
	}
	return defaultValue
}

func valFromEnv(name string) *string {
	v, hasV := os.LookupEnv(name)
	if !hasV {
		return nil
	}
	return &v
}

func (c *Config) valFromConfigFile(name string) *string {
	if !c.configFileIsLoaded && c.configFileExists() {
		err := c.loadConfigFileVals()
		if err != nil {
			panic(err)
		}
	}

	return c.configFileVals[name]
}

func (c *Config) configFileExists() bool {
	stat, err := os.Stat(c.configFileName)
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
	f, err := os.Open(defaultConfigFileName)
	if err != nil {
		return err
	}
	defer f.Close()

	isInMultiLineString := false
	var multiLineKey *string
	var multiLineValueBuilder strings.Builder
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if !isInMultiLineString {
			tokens := strings.Split(scanner.Text(), "=")
			if len(tokens) != 2 {
				return errors.New("radstack config file is not properly formed.  Must be key=value pairs")
			}
			v := tokens[1]
			c.configFileVals[tokens[0]] = &v
			if v[0:1] == "`" {
				isInMultiLineString = true
				multiLineKey = &tokens[0]
				// write all but the first char, as that would be the tick mark
				multiLineValueBuilder.WriteString(tokens[1][1:len(tokens[1])])
				multiLineValueBuilder.WriteRune('\n')
			}
		} else {
			v := scanner.Text()

			if v[len(v)-1:] == "`" {
				// write all but the last char, as that would be the tick mark
				multiLineValueBuilder.WriteString(v[:len(v)-1])
				isInMultiLineString = false
				c.configFileVals[*multiLineKey] = addrOf(multiLineValueBuilder.String())
				multiLineKey = nil
				multiLineValueBuilder = strings.Builder{}
			} else {
				multiLineValueBuilder.WriteString(v)
				multiLineValueBuilder.WriteRune('\n')
			}
		}
	}

	c.configFileIsLoaded = true
	return nil
}

func addrOf(s string) *string {
	return &s
}
