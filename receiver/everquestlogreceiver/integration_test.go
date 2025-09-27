// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package everquestlogreceiver

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/everquestlogreceiver/internal/metadata"
)

func TestEverQuestLogIntegration(t *testing.T) {
	// Create a temporary log file
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "everquest.log")
	
	// Write sample EverQuest log entries
	content := `[Wed May 28 23:58:03 2025] You slow down.
[Wed May 28 23:58:08 2025] Syrielle tells the group, 'DA SONG IS ON = DA SONG IS ON'
[Sat Sep 27 22:08:39 2025] A wyvern hits YOU for 26 points of damage.
[Sat Sep 27 22:08:41 2025] You slash a wyvern for 4 points of damage.
[Sat Sep 27 22:08:45 2025] You have become better at 1H Slashing! (15)
`
	
	err := os.WriteFile(logFile, []byte(content), 0600)
	require.NoError(t, err)
	
	// Create receiver factory and config
	factory := NewFactory()
	cfg := createDefaultConfig()
	cfg.InputConfig.Include = []string{logFile}
	cfg.InputConfig.StartAt = "beginning"
	cfg.InputConfig.PollInterval = 10 * time.Millisecond
	
	// Create consumer to capture logs
	sink := new(consumertest.LogsSink)
	
	// Create receiver
	receiver, err := factory.CreateLogs(context.Background(), receivertest.NewNopSettings(metadata.Type), cfg, sink)
	require.NoError(t, err, "Failed to create receiver")
	require.NotNil(t, receiver, "Receiver should not be nil")
	
	// Start receiver
	err = receiver.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err, "Failed to start receiver")
	
	// Wait for logs to be processed
	assert.Eventually(t, func() bool {
		return sink.LogRecordCount() >= 5
	}, 5*time.Second, 100*time.Millisecond, "Expected at least 5 log records")
	
	// Stop receiver
	err = receiver.Shutdown(context.Background())
	require.NoError(t, err, "Failed to shutdown receiver")
	
	// Verify we got the expected number of logs
	logs := sink.AllLogs()
	assert.GreaterOrEqual(t, len(logs), 1, "Should have at least one log batch")
	
	totalRecords := 0
	for _, logData := range logs {
		for i := 0; i < logData.ResourceLogs().Len(); i++ {
			resourceLogs := logData.ResourceLogs().At(i)
			for j := 0; j < resourceLogs.ScopeLogs().Len(); j++ {
				scopeLogs := resourceLogs.ScopeLogs().At(j)
				totalRecords += scopeLogs.LogRecords().Len()
			}
		}
	}
	
	assert.Equal(t, 5, totalRecords, "Should have exactly 5 log records")
	
	// Verify first log record contains expected data
	if len(logs) > 0 && logs[0].ResourceLogs().Len() > 0 {
		firstResourceLogs := logs[0].ResourceLogs().At(0)
		if firstResourceLogs.ScopeLogs().Len() > 0 {
			firstScopeLogs := firstResourceLogs.ScopeLogs().At(0)
			if firstScopeLogs.LogRecords().Len() > 0 {
				firstLogRecord := firstScopeLogs.LogRecords().At(0)
				
				// Check that the log body contains the expected message
				body := firstLogRecord.Body().AsString()
				assert.Contains(t, body, "You slow down", "Log body should contain the EverQuest message")
				
				// Check that timestamp was parsed (not zero)
				assert.NotEqual(t, pcommon.Timestamp(0), firstLogRecord.Timestamp(), "Timestamp should be parsed")
			}
		}
	}
}

func TestEverQuestLogConfigValidation(t *testing.T) {
	cfg := createDefaultConfig()
	
	// Test default configuration is valid
	assert.NotNil(t, cfg)
	assert.Equal(t, "Local", cfg.Location)
	assert.NotEmpty(t, cfg.BaseConfig.Operators)
	
	// Test configuration with custom timezone and required include
	cfg.Location = "America/New_York"
	cfg.InputConfig.Include = []string{"/path/to/everquest.log"}
	
	factory := NewFactory()
	receiver, err := factory.CreateLogs(context.Background(), receivertest.NewNopSettings(metadata.Type), cfg, consumertest.NewNop())
	assert.NoError(t, err)
	assert.NotNil(t, receiver)
}