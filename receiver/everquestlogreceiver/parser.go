// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package everquestlogreceiver

import (
	"regexp"
	"strings"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
)

// EverQuest timestamp format: [Day Mon DD HH:MM:SS YYYY]
const EverQuestTimeLayout = "Mon Jan 02 15:04:05 2006"

// EverQuest log categories
const (
	CategoryCombat       = "combat"
	CategoryCommunication = "communication"
	CategoryStatus       = "status"
	CategorySystem       = "system"
	CategoryMechanics    = "mechanics"
	CategoryUnknown      = "unknown"
)

var (
	// Regex to parse EverQuest log lines: [timestamp] message
	eqLogRegex = regexp.MustCompile(`^\[(\w+\s+\w+\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}\s+\d{4})\]\s+(.*)$`)
	
	// Combat log patterns
	combatPatterns = []*regexp.Regexp{
		regexp.MustCompile(`hits?\s+.+\s+for\s+\d+\s+points?\s+of\s+damage`), // damage
		regexp.MustCompile(`misses?\s+.+`),                                   // misses
		regexp.MustCompile(`you have slain`),                                 // kills
		regexp.MustCompile(`backstab|riposte|critical`),                      // special attacks
		regexp.MustCompile(`you (slash|pierce|bash|crush|kick|punch|bite)\s`), // attack types (with space after)
	}
	
	// Communication patterns
	commPatterns = []*regexp.Regexp{
		regexp.MustCompile(`tells? (you|the group|the guild|the raid)`),  // tells
		regexp.MustCompile(`says?`),                                      // says
		regexp.MustCompile(`shouts?`),                                    // shouts
		regexp.MustCompile(`auctions?`),                                  // auctions
		regexp.MustCompile(`ooc:`),                                       // out of character
	}
	
	// Status patterns  
	statusPatterns = []*regexp.Regexp{
		regexp.MustCompile(`you are (hungry|thirsty|tired)`),            // basic needs
		regexp.MustCompile(`you feel (better|worse)`),                   // health status
		regexp.MustCompile(`you (sit|stand|slow down|speed up)`),        // movement/position
		regexp.MustCompile(`you have become better at`),                 // skill ups
	}
	
	// System patterns
	systemPatterns = []*regexp.Regexp{
		regexp.MustCompile(`loading`),                                    // zone loading
		regexp.MustCompile(`welcome to everquest`),                      // login
		regexp.MustCompile(`thank you for playing everquest`),           // logout
		regexp.MustCompile(`you have entered`),                          // zone entry
		regexp.MustCompile(`system:`),                                    // system messages
	}
	
	// Game mechanics patterns
	mechanicsPatterns = []*regexp.Regexp{
		regexp.MustCompile(`begin casting|stop casting`),                // spellcasting
		regexp.MustCompile(`your .+ spell is ready`),                    // spell ready
		regexp.MustCompile(`song .+ begins to take hold`),               // bard songs
		regexp.MustCompile(`you activate|you deactivate`),               // abilities
		regexp.MustCompile(`auto attack`),                               // auto attack
	}
)

// ParseEverQuestLogLine parses a single EverQuest log line
func ParseEverQuestLogLine(line string, timezone *time.Location) (plog.LogRecord, error) {
	logRecord := plog.NewLogRecord()
	
	// Parse the log line with regex
	matches := eqLogRegex.FindStringSubmatch(line)
	if len(matches) != 3 {
		// If it doesn't match EverQuest format, treat as raw message
		logRecord.Body().SetStr(line)
		logRecord.SetTimestamp(pcommon.NewTimestampFromTime(time.Now().UTC()))
		logRecord.Attributes().PutStr("everquest.category", CategoryUnknown)
		logRecord.Attributes().PutStr("everquest.timezone", timezone.String())
		return logRecord, nil
	}
	
	timestampStr := matches[1]
	message := matches[2]
	
	// Parse timestamp and convert to UTC
	localTime, err := time.ParseInLocation(EverQuestTimeLayout, timestampStr, timezone)
	if err != nil {
		// If timestamp parsing fails, use current time
		localTime = time.Now()
	}
	utcTime := localTime.UTC()
	
	// Set log record fields
	logRecord.SetTimestamp(pcommon.NewTimestampFromTime(utcTime))
	logRecord.Body().SetStr(message)
	
	// Categorize the log message
	category := categorizeLogMessage(message)
	logRecord.Attributes().PutStr("everquest.category", category)
	logRecord.Attributes().PutStr("everquest.original_timestamp", timestampStr)
	logRecord.Attributes().PutStr("everquest.timezone", timezone.String())
	
	// Add specific attributes based on category
	addCategorySpecificAttributes(logRecord, category, message)
	
	return logRecord, nil
}

// categorizeLogMessage determines the category of an EverQuest log message
func categorizeLogMessage(message string) string {
	msgLower := strings.ToLower(message)
	
	// Check status patterns first (more specific patterns should be checked first)
	for _, pattern := range statusPatterns {
		if pattern.MatchString(msgLower) {
			return CategoryStatus
		}
	}
	
	// Check communication patterns
	for _, pattern := range commPatterns {
		if pattern.MatchString(msgLower) {
			return CategoryCommunication
		}
	}
	
	// Check system patterns
	for _, pattern := range systemPatterns {
		if pattern.MatchString(msgLower) {
			return CategorySystem
		}
	}
	
	// Check mechanics patterns
	for _, pattern := range mechanicsPatterns {
		if pattern.MatchString(msgLower) {
			return CategoryMechanics
		}
	}
	
	// Check combat patterns last (they have more general patterns)
	for _, pattern := range combatPatterns {
		if pattern.MatchString(msgLower) {
			return CategoryCombat
		}
	}
	
	return CategoryUnknown
}

// addCategorySpecificAttributes adds category-specific attributes to the log record
func addCategorySpecificAttributes(logRecord plog.LogRecord, category, message string) {
	switch category {
	case CategoryCombat:
		// Extract combat-specific information
		if regexp.MustCompile(`(\d+)\s+points?\s+of\s+damage`).MatchString(message) {
			damageMatch := regexp.MustCompile(`(\d+)\s+points?\s+of\s+damage`).FindStringSubmatch(message)
			if len(damageMatch) > 1 {
				logRecord.Attributes().PutStr("everquest.combat.damage", damageMatch[1])
			}
		}
		
	case CategoryCommunication:
		// Extract communication details
		if strings.Contains(message, "tells you") {
			logRecord.Attributes().PutStr("everquest.communication.type", "tell")
		} else if strings.Contains(message, "tells the group") {
			logRecord.Attributes().PutStr("everquest.communication.type", "group")
		} else if strings.Contains(message, "tells the guild") {
			logRecord.Attributes().PutStr("everquest.communication.type", "guild")
		}
		
	case CategoryStatus:
		// Extract status information
		if strings.Contains(message, "You have become better at") {
			logRecord.Attributes().PutStr("everquest.status.type", "skill_improvement")
		} else if strings.Contains(message, "hungry") || strings.Contains(message, "thirsty") {
			logRecord.Attributes().PutStr("everquest.status.type", "basic_needs")
		}
	}
}