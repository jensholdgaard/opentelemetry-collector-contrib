// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package everquestlogreceiver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEverQuestLogLine(t *testing.T) {
	// Use a fixed timezone for testing
	est, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	
	tests := []struct {
		name               string
		line               string
		expectedCategory   string
		expectedTimezone   *time.Location
		expectedAttributes map[string]string
	}{
		{
			name:             "Combat damage line",
			line:             "[Sat Sep 27 22:08:39 2025] A wyvern hits YOU for 26 points of damage.",
			expectedCategory: CategoryCombat,
			expectedTimezone: est,
			expectedAttributes: map[string]string{
				"everquest.combat.damage": "26",
			},
		},
		{
			name:             "Combat attack line",
			line:             "[Sat Sep 27 22:08:41 2025] You slash a wyvern for 4 points of damage.",
			expectedCategory: CategoryCombat,
			expectedTimezone: est,
			expectedAttributes: map[string]string{
				"everquest.combat.damage": "4",
			},
		},
		{
			name:             "Communication group tell",
			line:             "[Wed May 28 23:58:08 2025] Syrielle tells the group, 'DA SONG IS ON = DA SONG IS ON'",
			expectedCategory: CategoryCommunication,
			expectedTimezone: est,
			expectedAttributes: map[string]string{
				"everquest.communication.type": "group",
			},
		},
		{
			name:             "Communication personal tell",
			line:             "[Sat Sep 27 22:09:01 2025] Thorgrim tells you, 'heading to gfay'",
			expectedCategory: CategoryCommunication,
			expectedTimezone: est,
			expectedAttributes: map[string]string{
				"everquest.communication.type": "tell",
			},
		},
		{
			name:             "Status hungry/thirsty",
			line:             "[Sun Sep 28 10:15:22 2025] You are hungry.",
			expectedCategory: CategoryStatus,
			expectedTimezone: est,
			expectedAttributes: map[string]string{
				"everquest.status.type": "basic_needs",
			},
		},
		{
			name:             "Status skill improvement",
			line:             "[Sat Sep 27 22:08:45 2025] You have become better at 1H Slashing! (15)",
			expectedCategory: CategoryStatus,
			expectedTimezone: est,
			expectedAttributes: map[string]string{
				"everquest.status.type": "skill_improvement",
			},
		},
		{
			name:             "Mechanics spellcasting",
			line:             "[Sun Sep 28 10:16:45 2025] You begin casting Heal.",
			expectedCategory: CategoryMechanics,
			expectedTimezone: est,
			expectedAttributes: map[string]string{},
		},
		{
			name:             "Status movement",
			line:             "[Wed May 28 23:58:03 2025] You slow down.",
			expectedCategory: CategoryStatus,
			expectedTimezone: est,
			expectedAttributes: map[string]string{},
		},
		{
			name:             "Invalid format treated as unknown",
			line:             "This is not a valid EverQuest log line",
			expectedCategory: CategoryUnknown,
			expectedTimezone: est,
			expectedAttributes: map[string]string{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logRecord, err := ParseEverQuestLogLine(tt.line, tt.expectedTimezone)
			require.NoError(t, err)
			
			// Check category
			category, exists := logRecord.Attributes().Get("everquest.category")
			require.True(t, exists, "everquest.category attribute should exist")
			assert.Equal(t, tt.expectedCategory, category.AsString())
			
			// Check timezone attribute
			timezone, exists := logRecord.Attributes().Get("everquest.timezone")
			require.True(t, exists, "everquest.timezone attribute should exist")
			assert.Equal(t, tt.expectedTimezone.String(), timezone.AsString())
			
			// Check expected attributes
			for key, expectedValue := range tt.expectedAttributes {
				value, exists := logRecord.Attributes().Get(key)
				require.True(t, exists, "Expected attribute %s should exist", key)
				assert.Equal(t, expectedValue, value.AsString())
			}
			
			// Check that valid EverQuest lines have timestamps parsed correctly
			if tt.expectedCategory != CategoryUnknown {
				originalTimestamp, exists := logRecord.Attributes().Get("everquest.original_timestamp")
				require.True(t, exists, "everquest.original_timestamp should exist for valid lines")
				assert.NotEmpty(t, originalTimestamp.AsString())
				
				// Verify timestamp is in UTC
				timestamp := logRecord.Timestamp()
				utcTime := timestamp.AsTime()
				assert.Equal(t, time.UTC, utcTime.Location())
			}
		})
	}
}

func TestCategorizeLogMessage(t *testing.T) {
	tests := []struct {
		message  string
		expected string
	}{
		// Combat messages
		{"A wyvern hits YOU for 26 points of damage.", CategoryCombat},
		{"You slash a wyvern for 4 points of damage.", CategoryCombat},
		{"You have slain a skeleton!", CategoryCombat},
		{"The orc misses you.", CategoryCombat},
		{"You backstab the goblin for 45 points of damage.", CategoryCombat},
		
		// Communication messages
		{"Thorgrim tells you, 'hello there'", CategoryCommunication},
		{"Syrielle tells the group, 'ready?'", CategoryCommunication},
		{"Someone says, 'Anyone selling?'", CategoryCommunication},
		{"Player shouts, 'LFG!'", CategoryCommunication},
		{"OOC: Anyone know where the bank is?", CategoryCommunication},
		
		// Status messages
		{"You are hungry.", CategoryStatus},
		{"You are thirsty.", CategoryStatus},
		{"You feel better.", CategoryStatus},
		{"You sit down.", CategoryStatus},
		{"You have become better at 1H Slashing! (15)", CategoryStatus},
		{"You slow down.", CategoryStatus},
		
		// System messages
		{"LOADING, PLEASE WAIT...", CategorySystem},
		{"Welcome to EverQuest!", CategorySystem},
		{"You have entered The Qeynos Hills.", CategorySystem},
		{"SYSTEM: Server going down in 5 minutes", CategorySystem},
		
		// Mechanics messages
		{"You begin casting Heal.", CategoryMechanics},
		{"Your Greater Heal spell is ready.", CategoryMechanics},
		{"The song of vigor begins to take hold.", CategoryMechanics},
		{"You activate your discipline.", CategoryMechanics},
		{"Engaging auto attack.", CategoryMechanics},
		
		// Unknown messages
		{"Some random message that doesn't fit patterns", CategoryUnknown},
		{"", CategoryUnknown},
	}
	
	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			result := categorizeLogMessage(tt.message)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTimezoneConversion(t *testing.T) {
	// Test timezone conversion from local to UTC
	est, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	
	pst, err := time.LoadLocation("America/Los_Angeles")
	require.NoError(t, err)
	
	// Test with same timestamp in different timezones
	line := "[Sat Sep 27 22:08:39 2025] A wyvern hits YOU for 26 points of damage."
	
	// Parse with EST
	logRecordEST, err := ParseEverQuestLogLine(line, est)
	require.NoError(t, err)
	
	// Parse with PST
	logRecordPST, err := ParseEverQuestLogLine(line, pst)
	require.NoError(t, err)
	
	// Both should be converted to UTC, but the actual UTC times should be different
	// because the local times represent different absolute moments
	timestampEST := logRecordEST.Timestamp().AsTime()
	timestampPST := logRecordPST.Timestamp().AsTime()
	
	assert.Equal(t, time.UTC, timestampEST.Location())
	assert.Equal(t, time.UTC, timestampPST.Location())
	
	// PST timestamp should be 3 hours later than EST timestamp (for the same local time)
	diff := timestampPST.Sub(timestampEST)
	assert.Equal(t, 3*time.Hour, diff)
}