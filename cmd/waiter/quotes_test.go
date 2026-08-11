// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"strings"
	"testing"
)

func TestRandomBuildQuote(t *testing.T) {
	seen := make(map[string]struct{}, len(buildQuotes))
	for range len(buildQuotes) * 3 {
		quote := randomBuildQuote()
		if quote == "" {
			t.Fatal("expected non-empty quote")
		}
		seen[quote] = struct{}{}
	}
	if len(seen) != len(buildQuotes) {
		t.Fatalf("expected all %d quotes to appear eventually, got %d", len(buildQuotes), len(seen))
	}
}

func TestFormatMotd(t *testing.T) {
	motd := formatMotd()
	containsQuote := false
	for _, quote := range buildQuotes {
		if strings.Contains(motd, quote) {
			containsQuote = true
			break
		}
	}
	if !containsQuote {
		t.Fatalf("expected motd to contain a build quote, got: %q", motd)
	}
	if !strings.Contains(motd, "|") {
		t.Fatalf("expected motd to contain the ASCII ship, got: %q", motd)
	}
}
