// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:generate mdatagen metadata.yaml

// Package everquestlogreceiver implements a receiver for EverQuest game log files.
// It parses log entries with the format [Day Mon DD HH:MM:SS YYYY] Message,
// converts local timestamps to UTC, and categorizes different types of log entries
// including combat, communication, status, system, and game mechanics messages.
package everquestlogreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/everquestlogreceiver"