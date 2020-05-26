package lib

import "testing"

func TestGenerateCert(t *testing.T) {
	ca := DefaultCa()

	for _, n := range []string{"cn.xwsea.com", "127.0.0.1", "localhost"} {
		_, e := GenerateCert(n, ca)
		if e != nil {
			t.Errorf("make cert failed: %s", e)
		}

	}
}
