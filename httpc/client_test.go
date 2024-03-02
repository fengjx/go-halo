package httpc_test

import (
	"testing"

	"github.com/fengjx/go-halo/httpc"
)

func TestPostJson(t *testing.T) {
	cli := httpc.New(&httpc.Config{
		DefaultHeaders: map[string]string{
			"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36 Edg/114.0.1823.67",
		},
	})
	resp, err := cli.Post("https://httpbin.org/post", map[string]string{
		"name": "test_name",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(resp.Body()))
}
