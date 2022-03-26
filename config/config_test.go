package config

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNoValue(t *testing.T) {
	assert.Nil(t, NewConfig().GetValue("nope"))
}

func TestEnvVal(t *testing.T) {
	k := "KEY"
	v := "Value"
	os.Setenv(k, v)
	assert.Equal(t, v, *NewConfig().GetValue(k))
}

func TestFileVal(t *testing.T) {
	k := "KEY2"
	v := "Value2"
	f, err := os.CreateTemp("", "testConfigFile")
	f.Chmod(0600)
	if err != nil {
		panic(err)
	}
	configFileName = f.Name()

	f.WriteString(fmt.Sprintf("%s=%s", k, v))
	assert.Equal(t, v, *NewConfig().GetValue(k))
	os.Remove(f.Name())
}

func TestEnvBeforeFileVal(t *testing.T) {
	k := "KEY"
	envV := "Value"
	fileV := "Value2"
	f, err := os.CreateTemp("", "testConfigFile")
	f.Chmod(0600)
	if err != nil {
		panic(err)
	}
	configFileName = f.Name()
	f.WriteString(fmt.Sprintf("%s=%s", k, fileV))

	os.Setenv(k, envV)

	assert.Equal(t, envV, *NewConfig().GetValue(k))
	os.Remove(f.Name())
}
