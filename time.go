package mysqltime

import (
    "database/sql/driver"
    _ "errors"
    "fmt"
    "strings"
	"strconv"
    "time"
)

// Time represents a custom type for handling MySQL TIME data.
type Time struct {
    duration time.Duration
    valid    bool
}

// NewTime creates a new mysqltime.Time from a time.Duration.
func NewTime(d time.Duration) Time {
    return Time{duration: d, valid: true}
}

// SetDuration sets the Time based on a time.Duration.
func (t *Time) SetDuration(d time.Duration) error {
    t.duration = d
    t.valid = true
    return nil
}

// GetDuration returns the time.Duration and a boolean indicating if the value is valid.
func (t Time) GetDuration() (time.Duration, bool) {
    if !t.valid {
        return 0, false
    }
    return t.duration, true
}

// MarshalText implements the encoding.TextMarshaler interface.
func (t Time) MarshalText() ([]byte, error) {
    if !t.valid {
        return []byte{}, nil
    }
    return []byte(t.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (t *Time) UnmarshalText(text []byte) error {
    str := string(text)
    if str == "" {
        t.valid = false
        return nil
    }
    duration, err := parseMySQLTime(str)
    if err != nil {
        return err
    }
    t.duration = duration
    t.valid = true
    return nil
}

// Value implements the driver.Valuer interface for database serialization.
func (t Time) Value() (driver.Value, error) {
    if !t.valid {
        return nil, nil
    }
    return t.String(), nil
}

// Scan implements the sql.Scanner interface for database deserialization.
func (t *Time) Scan(value interface{}) error {
    if value == nil {
        t.duration = 0
        t.valid = false
        return nil
    }

    var strVal string
    switch v := value.(type) {
    case []byte:
        strVal = string(v) // Handle TIME values returned as []byte
    case string:
        strVal = v // Handle TIME values returned as string
    default:
        return fmt.Errorf("unsupported type for mysqltime.Time: %T", value)
    }

    duration, err := parseMySQLTime(strVal)
    if err != nil {
        return fmt.Errorf("failed to parse mysqltime.Time from value '%s': %w", strVal, err)
    }

    t.duration = duration
    t.valid = true
    return nil
}

func (t Time) String() string {
    if !t.valid {
        return ""
    }

    // Calculate the absolute values for hours, minutes, and seconds
    hours := int64(t.duration.Hours())
    minutes := int64(t.duration.Minutes()) % 60
    seconds := int64(t.duration.Seconds()) % 60

    // Format the string without a "+" sign
    if t.duration < 0 {
        return fmt.Sprintf("-%03d:%02d:%02d", -hours, -minutes, -seconds)
    }
    return fmt.Sprintf("%03d:%02d:%02d", hours, minutes, seconds)
}

// parseMySQLTime parses a MySQL TIME string into a time.Duration.
func parseMySQLTime(s string) (time.Duration, error) {
    // Handle cases with colons directly
    if strings.Contains(s, ":") {
        negative := false
        if strings.HasPrefix(s, "-") {
            negative = true
            s = s[1:]
        }

        var hours, minutes, seconds int64
        parts := strings.Split(s, ":")
        if len(parts) == 2 { // Cases like '11:12' should be interpreted as '11:12:00'
            parsedHours, err := strconv.ParseInt(parts[0], 10, 64)
            if err != nil {
                return 0, fmt.Errorf("invalid TIME format: %s", s)
            }
            parsedMinutes, err := strconv.ParseInt(parts[1], 10, 64)
            if err != nil {
                return 0, fmt.Errorf("invalid TIME format: %s", s)
            }
            hours, minutes, seconds = parsedHours, parsedMinutes, 0
        } else if len(parts) == 3 { // Cases like '11:12:13'
            parsedHours, err := strconv.ParseInt(parts[0], 10, 64)
            if err != nil {
                return 0, fmt.Errorf("invalid TIME format: %s", s)
            }
            parsedMinutes, err := strconv.ParseInt(parts[1], 10, 64)
            if err != nil {
                return 0, fmt.Errorf("invalid TIME format: %s", s)
            }
            parsedSeconds, err := strconv.ParseInt(parts[2], 10, 64)
            if err != nil {
                return 0, fmt.Errorf("invalid TIME format: %s", s)
            }
            hours, minutes, seconds = parsedHours, parsedMinutes, parsedSeconds
        } else {
            return 0, fmt.Errorf("invalid TIME format: %s", s)
        }

        duration := time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second
        if negative {
            duration = -duration
        }
        return duration, nil
    }

    // Handle cases without colons (abbreviated times)
    switch len(s) {
    case 6: // HHMMSS format
        hours, err := strconv.Atoi(s[:2])
        if err != nil {
            return 0, err
        }
        minutes, err := strconv.Atoi(s[2:4])
        if err != nil {
            return 0, err
        }
        seconds, err := strconv.Atoi(s[4:])
        if err != nil {
            return 0, err
        }
        return time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second, nil
    case 4: // MMSS format
        minutes, err := strconv.Atoi(s[:2])
        if err != nil {
            return 0, err
        }
        seconds, err := strconv.Atoi(s[2:])
        if err != nil {
            return 0, err
        }
        return time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second, nil
    case 2: // SS format
        seconds, err := strconv.Atoi(s)
        if err != nil {
            return 0, err
        }
        return time.Duration(seconds) * time.Second, nil
    default:
        return 0, fmt.Errorf("invalid TIME format: %s", s)
    }
}