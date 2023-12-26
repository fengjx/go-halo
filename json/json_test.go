package json_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fengjx/go-halo/json"
)

func TestGetPathVal(t *testing.T) {
	jsonStr := `{"code":1,"message":"success","result":{"totalMsgCount":0},"time":"2023-03-18 23:30:54"}`
	assert.Equal(t, "2023-03-18 23:30:54", json.GetNodeFromString(jsonStr, "time").ToString())
}

func TestDelay(t *testing.T) {
	data := map[string]any{
		"foo": "bar",
	}
	t.Logf("delay json: %s", json.ToJsonDelay(data))
}
