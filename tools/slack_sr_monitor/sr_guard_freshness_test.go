package main

import (
	"testing"
	"time"
)

func TestIsBlockFresh(t *testing.T) {
	if !isBlockFresh(time.Now()) {
		t.Error("expected current block time to be fresh")
	}
	if isBlockFresh(time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC)) {
		t.Error("expected year-2018 block time to be stale")
	}
}
