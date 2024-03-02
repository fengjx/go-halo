package addr_test

import (
	"net"
	"testing"

	"github.com/fengjx/go-halo/addr"
)

func TestIsLocal(t *testing.T) {
	testData := []struct {
		addr   string
		expect bool
	}{
		{"localhost", true},
		{"localhost:8080", true},
		{"127.0.0.1", true},
		{"127.0.0.1:1001", true},
		{"80.1.1.1", false},
	}

	for _, d := range testData {
		res := addr.IsLocal(d.addr)
		if res != d.expect {
			t.Fatalf("expected %t got %t", d.expect, res)
		}
	}
}

func TestExtractor(t *testing.T) {
	testData := []struct {
		addr   string
		expect string
		parse  bool
	}{
		{"127.0.0.1", "127.0.0.1", false},
		{"10.0.0.1", "10.0.0.1", false},
		{"", "", true},
		{"0.0.0.0", "", true},
		{"[::]", "", true},
		{"::", "", true},
	}

	for _, d := range testData {
		address, err := addr.Extract(d.addr)
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}
		t.Log("address", d.addr, address)
		if d.parse {
			ip := net.ParseIP(address)
			if ip == nil {
				t.Error("Unexpected nil IP")
			}
		} else if address != d.expect {
			t.Errorf("Expected %s got %s", d.expect, address)
		}
	}
}

func TestExtractHostPort(t *testing.T) {
	listen, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer listen.Close()
	host, port, err := addr.ExtractHostPort(listen.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(host, port)
}

func TestInnerIP(t *testing.T) {
	t.Log(addr.InnerIP())
}
