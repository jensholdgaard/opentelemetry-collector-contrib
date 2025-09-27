# EverQuest Log Receiver

This receiver is designed to parse EverQuest game log files and convert them to structured telemetry data. It acts as a local agent that monitors EverQuest log files with proper time zone handling and log categorization.

## Features

- **EverQuest Log Parser**: Parses log entries with the format `[Day Mon DD HH:MM:SS YYYY] Message`
- **UTC Time Conversion**: Converts local system time to UTC for consistent log visualization across different time zones
- **Log Categories**: Supports different types of log entries including:
  - Combat logs (damage, attacks, misses)
  - Communication logs (tells, guild messages, group messages)
  - Status messages (hunger, thirst, fatigue)
  - System messages (logging on/off, spell effects)
  - Game mechanics (auto attack, song effects)
- **Structured Output**: Converts log entries to structured telemetry data with proper attributes
- **File Monitoring**: Built on the reliable stanza file input framework with support for:
  - Log rotation handling
  - Configurable polling intervals
  - Starting position control (beginning/end)

## Configuration

| Field       | Default      | Description                                                                                                                                                                                                                                                                                                  |
|-------------|--------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `include`   | required     | A list of file glob patterns that match the file paths to be read                                                                                                                                                                                                                                          |
| `exclude`   | []           | A list of file glob patterns to exclude from reading                                                                                                                                                                                                                                                        |
| `location`  | `Local`      | The timezone location to use when parsing timestamps. Can be `Local` for system timezone or IANA timezone names like `America/New_York`                                                                                                                                                                   |
| `start_at`  | `end`        | Where to start reading files. Options are `beginning` or `end`                                                                                                                                                                                                                                              |
| `poll_interval` | `200ms`  | The duration between filesystem polls                                                                                                                                                                                                                                                                       |

For additional configuration options (like operators, retry_on_failure, etc.), see the [stanza documentation](../../pkg/stanza/docs/README.md).

## Example Configurations

### Basic Configuration

```yaml
receivers:
  everquestlog:
    include:
      - /path/to/everquest/Logs/eqlog_*.txt
    location: America/New_York
    start_at: beginning
    poll_interval: 100ms
```

### Advanced Configuration with File Exclusion

```yaml
receivers:
  everquestlog:
    include:
      - /path/to/everquest/Logs/eqlog_*.txt
    exclude:
      - /path/to/everquest/Logs/eqlog_*_old.txt
    location: UTC
    start_at: end
    poll_interval: 500ms
    retry_on_failure:
      enabled: true
      initial_interval: 1s
      max_interval: 30s
      max_elapsed_time: 5m
```

## Log Format

The receiver handles EverQuest logs with the following format:
```
[Day Mon DD HH:MM:SS YYYY] Message
```

Examples:
```
[Wed May 28 23:58:03 2025] You slow down.
[Wed May 28 23:58:08 2025] Syrielle tells the group, 'DA SONG IS ON = DA SONG IS ON'
[Sat Sep 27 22:08:39 2025] A wyvern hits YOU for 26 points of damage.
[Sat Sep 27 22:08:41 2025] You slash a wyvern for 4 points of damage.
[Sat Sep 27 22:08:45 2025] You have become better at 1H Slashing! (15)
[Sat Sep 27 22:09:01 2025] Thorgrim tells you, 'heading to gfay'
[Sun Sep 28 10:15:22 2025] You are hungry.
[Sun Sep 28 10:15:22 2025] You are thirsty.
[Sun Sep 28 10:16:45 2025] You begin casting Heal.
[Sun Sep 28 10:16:48 2025] You feel better.
```

## Output Structure

The receiver automatically categorizes logs and adds structured attributes:

### Standard Attributes

- `everquest.category`: One of `combat`, `communication`, `status`, `system`, `mechanics`, or `unknown`
- `everquest.original_timestamp`: The original timestamp string from the log
- `everquest.timezone`: The timezone used for parsing

### Category-Specific Attributes

#### Combat Logs
- `everquest.combat.damage`: Damage amount (when applicable)

#### Communication Logs
- `everquest.communication.type`: `tell`, `group`, `guild`, etc.

#### Status Logs
- `everquest.status.type`: `skill_improvement`, `basic_needs`, etc.

## Log Categories

### Combat
- Damage dealing and receiving
- Misses and attacks
- Kills and special attacks

### Communication
- Player tells (private messages)
- Group, guild, and raid chat
- Public channels (say, shout, auction, OOC)

### Status
- Character state changes (hunger, thirst, fatigue)
- Skill improvements
- Movement and position changes
- Health status updates

### System
- Zone loading and transitions
- Login/logout messages
- Server announcements

### Mechanics
- Spellcasting
- Spell readiness
- Bard songs
- Ability activation
- Auto attack

## Time Zone Handling

The receiver automatically converts EverQuest log timestamps (which are in local time) to UTC for consistent processing across different time zones. This is particularly important for:

- Multi-zone deployments
- Log aggregation from different geographical locations
- Consistent time-based queries and visualizations

The `location` configuration parameter allows you to specify the timezone of the EverQuest client that generated the logs.

## Performance Considerations

- Use specific file patterns in `include` to avoid monitoring unnecessary files
- Set appropriate `poll_interval` based on your log volume and latency requirements
- Consider using `exclude` patterns to filter out rotated or archived log files
- The receiver is optimized for real-time log monitoring and handles log rotation automatically

## Troubleshooting

### Common Issues

1. **No logs being processed**: Check that the `include` patterns match your EverQuest log files
2. **Incorrect timestamps**: Verify the `location` setting matches your EverQuest client's timezone
3. **Missing log entries**: Ensure the log files are readable by the collector process
4. **High CPU usage**: Increase `poll_interval` to reduce filesystem polling frequency

### Debug Configuration

Enable debug logging to troubleshoot parsing issues:

```yaml
service:
  telemetry:
    logs:
      level: debug
```

This will show detailed information about log parsing and categorization.