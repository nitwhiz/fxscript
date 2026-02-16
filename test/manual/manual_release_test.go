//go:build !debug

package test

import (
	"testing"
)

func TestManual(t *testing.T) {
	t.Skip("missing debug tag")
}
