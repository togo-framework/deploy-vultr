package vultr

import (
	"strings"
	"testing"
)

func TestCloudInit(t *testing.T) {
	if ci := cloudInit("img:9"); !strings.Contains(ci, "img:9") || !strings.Contains(ci, "docker") {
		t.Fatal("bad cloud-init")
	}
}
