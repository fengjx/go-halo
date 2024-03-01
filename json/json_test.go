package json_test

import (
	"os"
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

func TestEncoder(t *testing.T) {
	data := map[string]any{
		"msg": "ok",
		"data": []map[string]any{
			{
				"id":  2,
				"foo": "bar2",
			},
			{
				"id":  3,
				"foo": "bar3",
			},
		},
	}
	err := json.NewEncoder(os.Stdout).Encode(data)
	if err != nil {
		t.Fatal(err)
	}
}
