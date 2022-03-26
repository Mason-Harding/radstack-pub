package db

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDbSession(t *testing.T) {
	assert.NotNil(t, DocumentSession)
}
