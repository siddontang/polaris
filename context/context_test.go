package context

import (
	"testing"
)

func TestContext(t *testing.T) {
	Set("1", "hello")

	v := Get("1").(string)
	if v != "hello" {
		t.Fatal(v)
	}
}
