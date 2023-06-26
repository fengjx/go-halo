package json

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetPathVal(t *testing.T) {
	jsonStr := `{"code":1,"message":"success","result":{"totalMsgCount":0},"time":"2023-03-18 23:30:54"}`
	assert.Equal(t, "2023-03-18 23:30:54", GetPathVal(jsonStr, "time").ToString())
}
