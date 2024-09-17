package encrypt

import "testing"

func TestMd5SumString(t *testing.T) {
	t.Logf(Md5SumString("admin"))
}
