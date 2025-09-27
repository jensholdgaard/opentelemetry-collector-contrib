// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package everquestlogreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/everquestlogreceiver"

import (
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/receiver"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/consumerretry"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/adapter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/input/file"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/parser/regex"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/entry"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/everquestlogreceiver/internal/metadata"
)

// NewFactory creates a factory for everquestlog receiver
func NewFactory() receiver.Factory {
	return adapter.NewFactory(ReceiverType{}, metadata.LogsStability)
}

// ReceiverType implements stanza.LogReceiverType
// to create an EverQuest log tailing receiver
type ReceiverType struct{}

// Type is the receiver type
func (ReceiverType) Type() component.Type {
	return metadata.Type
}

// CreateDefaultConfig creates a config with type and version
func (ReceiverType) CreateDefaultConfig() component.Config {
	return createDefaultConfig()
}

func createDefaultConfig() *EverQuestLogConfig {
	return &EverQuestLogConfig{
		BaseConfig: adapter.BaseConfig{
			Operators:      createDefaultOperators(),
			RetryOnFailure: consumerretry.NewDefaultConfig(),
		},
		InputConfig: *file.NewConfig(),
		Location:    "Local",
	}
}

// createDefaultOperators creates the default parsing operators for EverQuest logs
func createDefaultOperators() []operator.Config {
	regexConfig := regex.NewConfig()
	regexConfig.Regex = `^\[(?P<timestamp>\w+\s+\w+\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}\s+\d{4})\]\s+(?P<message>.*)$`
	
	// Parse timestamp field
	timestampField := entry.NewAttributeField("timestamp")
	timeConfig := helper.NewTimeParser()
	timeConfig.Layout = "%a %b %d %H:%M:%S %Y"  // strptime format for "Wed May 28 23:58:03 2025"
	timeConfig.ParseFrom = &timestampField
	timeConfig.Location = "Local" // Will be overridden by config
	_ = timeConfig.Validate() // Validate like the other receivers do
	regexConfig.TimeParser = &timeConfig
	
	return []operator.Config{
		{
			Builder: regexConfig,
		},
	}
}

// BaseConfig gets the base config from config, for now
func (ReceiverType) BaseConfig(cfg component.Config) adapter.BaseConfig {
	return cfg.(*EverQuestLogConfig).BaseConfig
}

// EverQuestLogConfig defines configuration for the everquestlog receiver
type EverQuestLogConfig struct {
	InputConfig        file.Config `mapstructure:",squash"`
	adapter.BaseConfig `mapstructure:",squash"`
	Location           string      `mapstructure:"location,omitempty"`

	// prevent unkeyed literal initialization
	_ struct{}
}

// InputConfig unmarshals the input operator
func (ReceiverType) InputConfig(cfg component.Config) operator.Config {
	return operator.NewConfig(&cfg.(*EverQuestLogConfig).InputConfig)
}