// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package ctxlog

import (
	"context"
	"flag"
	"io"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var flagOptions = &zap.Options{}

type levelFlag struct{}

func (l levelFlag) Set(input string) error { return nil }

func (l levelFlag) String() string { return "" }

// CustomZapFlagSet ...
func CustomZapFlagSet() *flag.FlagSet {
	f := flag.NewFlagSet("zap", flag.ExitOnError)
	flagOptions.BindFlags(f)

	var foo levelFlag
	f.Var(foo, "zap-level", "Deprecated: Please use --zap-log-level instead; set log level")

	return f
}

// NewLogger returns a new Logger instance
// by using the operator-sdk log/zap logging
// implementation TODO
func NewLogger(name string) logr.Logger {
	l := zap.New(zap.UseFlagOptions(flagOptions))
	logf.SetLogger(l)

	l = l.WithName(name)

	return l
}

// NewLoggerTo returns a new Logger which logs
// to a given destination.
func NewLoggerTo(destWriter io.Writer, name string) logr.Logger {
	l := zap.New(zap.UseFlagOptions(flagOptions), zap.WriteTo(destWriter))

	logf.SetLogger(l)

	l = l.WithName(name)

	return l
}

// Error returns an ERROR level log from an specified context
func Error(ctx context.Context, err error, msg string, v ...interface{}) {
	l := ExtractLogger(ctx)
	l.Error(err, msg, v...)
}

// Debug returns an DEBUG level log from an specified context
func Debug(ctx context.Context, msg string, v ...interface{}) {
	l := ExtractLogger(ctx)
	l.V(1).Info(msg, v...)
}

// Info returns an INFO level log from an specified context
func Info(ctx context.Context, msg string, v ...interface{}) {
	l := ExtractLogger(ctx)
	l.Info(msg, v...)
}
