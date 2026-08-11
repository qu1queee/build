// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

const shipASCII = `
        |    |    |
       )_)  )_)  )_)
      )___))___))___)\
     )____)____)_____)\\
_____|____|____|____\\\
\                     /
`

var buildQuotes = []string{
	"Your image is building. The harbor awaits.",
	"Layers cache or layers don't — there is no try without push.",
	"Shipwright never sinks; it only retries.",
	"A smooth sea never made a skilled shipbuilder.",
	"May your digests be immutable and your CVEs be zero.",
	"Build once, run anywhere. Debug twice, everywhere.",
	"The best time to pin your base image was yesterday. The second best time is now.",
	"Every great container image starts as a humble Dockerfile.",
	"Patience is a virtue, especially when pulling from Docker Hub.",
	"Today's BuildRun is tomorrow's production deployment.",
	"Keep calm and kubectl get buildruns -w.",
	"From git clone to image push: the voyage continues.",
}

// randomBuildQuote returns a randomly selected build-related message.
func randomBuildQuote() string {
	return buildQuotes[rand.IntN(len(buildQuotes))]
}

// formatMotd renders the ASCII ship and a random quote for waiting pods.
func formatMotd() string {
	var b strings.Builder
	b.WriteString(strings.TrimRight(shipASCII, "\n"))
	b.WriteString("\n\n  ")
	b.WriteString(randomBuildQuote())
	b.WriteString("\n")
	return b.String()
}

// printMotd writes a message-of-the-day to stdout.
func printMotd() {
	fmt.Print(formatMotd())
}
