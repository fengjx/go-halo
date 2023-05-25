package event

import (
	"testing"
	"time"
)

func TestEvent(t *testing.T) {
	var test Topic = "topic-test"
	for i := 0; i < 100; i++ {
		idx := i
		Subscribe(test, func(msg interface{}) {
			eventMsg := msg.(map[string]string)
			t.Logf("idx - %d - %s\n", idx, eventMsg["hello"])
			time.Sleep(time.Second)
		})
	}
	Publish(test, map[string]string{"hello": "world"})

	Quit()
}
