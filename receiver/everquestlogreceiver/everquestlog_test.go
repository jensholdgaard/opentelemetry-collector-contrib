// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package everquestlogreceiver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/everquestlogreceiver/internal/metadata"
)

func TestDefaultConfig(t *testing.T) {
	cfg := createDefaultConfig()
	assert.NotNil(t, cfg, "Failed to create default config")
	assert.NoError(t, componenttest.CheckConfigStruct(cfg))
}

func TestType(t *testing.T) {
	factory := NewFactory()
	ft := factory.Type()
	assert.EqualValues(t, metadata.Type, ft)
}

func TestCreateLogs(t *testing.T) {
	factory := NewFactory()
	cfg := createDefaultConfig()
	// Configure a log file that doesn't exist to avoid actual file operations in tests
	cfg.InputConfig.Include = []string{"/nonexistent/path/test.log"}

	receiver, err := factory.CreateLogs(context.Background(), receivertest.NewNopSettings(metadata.Type), cfg, consumertest.NewNop())
	assert.NoError(t, err, "Factory.CreateLogs failed")
	assert.NotNil(t, receiver, "Factory.CreateLogs returned nil")
}