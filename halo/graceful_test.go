package halo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetInterval(t *testing.T) {
	SetInterval(func() {
		t.Log("Interval")
	}, time.Second)
	err := Wait(time.Second * 3)
	if err != nil {
		assert.NoError(t, err)
	}
}
