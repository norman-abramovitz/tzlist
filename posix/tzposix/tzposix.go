package tzposix

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// HumanReadableTZ parses a POSIX TZ string and returns a human-readable description.
// It handles a common format like "EST5EDT,M3.2.0/02:00:00,M11.1.0/02:00:00"
func HumanReadableTZ(posixTZ string) (string, error) {
	// A basic regex to capture the main parts:
	// 1. Standard Time Abbr (STD)
	// 2. STD Offset
	// 3. Optional DST Abbr (DST)
	// 4. DST Offset (DST)
	// 5. Optional DST Offset (assumed +1 hour if absent)
	// 6. Optional DST Start/End Rules
	regex := `^(?<StdName>[[:alpha:]]{3,}|<[[:alnum:]+-]+>)` +
		`(?<StdOffset>[0-9:+-]+)` +
		`(?<DstName>[[:alpha:]]{3,}|<[[:alnum:]+-]+>)?` +
		`(?<DstOffset>[0-9:+-]+)?` +
		`,?(?<StartRule>(?:M|J)?[0-9\.]+/[0-9:+-]+|(?:M|J)?[0-9\.]+)?` +
		`,?(?<EndRule>(?:M|J)?[0-9\.]+/[0-9:+-]+|(?:M|J)?[0-9\.]+)?$`
	re := regexp.MustCompile(regex)

	matches := re.FindStringSubmatch(posixTZ)

	if matches == nil {
		return "", fmt.Errorf("invalid POSIX TZ string format: %s", posixTZ)
	}
	// for i, name := range re.SubexpNames() {
	// fmt.Printf("'%s'\t %d -> %s\n", name, i, matches[i])
	// }
	// for i := range matches {
	// fmt.Printf("''\t %d -> %s\n", i, matches[i])
	// }

	// fmt.Printf("DEBUG  %d  %+v\n", len(matches), matches)

	stdAbbr := matches[1]
	stdOffsetStr := matches[2]
	dstAbbr := matches[3]
	dstOffsetStr := matches[4]
	startRule := matches[5]
	endRule := matches[6]

	// Convert offset to human-friendly format (UTC+/-H:M)
	stdOffset, err := parseOffset(stdOffsetStr)
	if err != nil {
		return "", fmt.Errorf("invalid standard offset: %w", err)
	}
	stdDesc := fmt.Sprintf("Standard Time: %s (UTC%s)", stdAbbr, formatOffset(stdOffset))

	if dstAbbr == "" {
		return stdDesc + "\n(No Daylight Saving Time rules)", nil
	}

	// Calculate DST offset if not explicitly provided (POSIX default is 1 hour ahead)
	dstOffset := stdOffset - 3600 // DST is typically 1 hour *ahead* (west) of standard time, so offset is smaller in POSIX

	if dstOffsetStr != "" {
		parsedDstOffset, err := parseOffset(dstOffsetStr)
		if err != nil {
			return "", fmt.Errorf("invalid daylight offset: %w", err)
		}
		dstOffset = parsedDstOffset
	}
	dstDesc := fmt.Sprintf("Daylight Time: %s (UTC%s)", dstAbbr, formatOffset(dstOffset))

	rulesDesc := ""
	if startRule != "" && endRule != "" {
		rulesDesc = fmt.Sprintf("\nRules: Starts %s, Ends %s", parseRule(startRule), parseRule(endRule))
	}

	return fmt.Sprintf("%s\n%s%s", stdDesc, dstDesc, rulesDesc), nil
}

// parseOffset converts a POSIX offset string (e.g., "5", "-10:30") to seconds west of UTC
func parseOffset(offsetStr string) (int, error) {
	// POSIX offsets are West of Greenwich, opposite of ISO 8601
	// "EST5" means 5 hours West of UTC (UTC+5 if we follow standard notation)

	sign := 1
	if strings.HasPrefix(offsetStr, "+") {
		offsetStr = strings.TrimPrefix(offsetStr, "+")
	} else if strings.HasPrefix(offsetStr, "-") {
		offsetStr = strings.TrimPrefix(offsetStr, "-")
		sign = -1
	}

	parts := strings.Split(offsetStr, ":")
	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	minutes := 0
	if len(parts) > 1 {
		minutes, err = strconv.Atoi(parts[1])
		if err != nil {
			return 0, err
		}
	}
	return sign * (hours*3600 + minutes*60), nil
}

// formatOffset converts seconds offset to " +H:M" or " -H:M" string
func formatOffset(offsetSeconds int) string {
	sign := "+"
	if offsetSeconds > 0 {
		sign = "-" // POSIX is backwards, so >0 seconds is actually UTC-X
		offsetSeconds = -offsetSeconds
	}
	// Absolute value for calculation
	absOffset := offsetSeconds
	if absOffset < 0 {
		absOffset = -absOffset
	}

	hours := absOffset / 3600
	minutes := (absOffset % 3600) / 60
	return fmt.Sprintf(" %s%02d:%02d", sign, hours, minutes)
}

// parseRule converts a POSIX rule string (e.g., "M3.2.0/02:00:00") to a description
func parseRule(rule string) string {
	if strings.HasPrefix(rule, "M") {
		// Month.Week.Day format
		parts := strings.Split(strings.TrimPrefix(rule, "M"), ".")
		if len(parts) >= 3 {
			month := parts[0]
			week := parts[1]
			day := parts[2]
			timeStr := "02:00:00" // default
			if strings.Contains(day, "/") {
				timeParts := strings.Split(day, "/")
				day = timeParts[0]
				timeStr = timeParts[1]
			}

			// Mapping basic values to human terms
			months := []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}
			weekDesc := map[string]string{"1": "first", "2": "second", "3": "third", "4": "fourth", "5": "last"}
			dayDesc := map[string]string{"0": "Sunday", "1": "Monday", "2": "Tuesday", "3": "Wednesday", "4": "Thursday", "5": "Friday", "6": "Saturday"}

			return fmt.Sprintf("on the %s %s of %s at %s", weekDesc[week], dayDesc[day], months[atoi(month)-1], timeStr)
		}
	}
	// Handle Julian day or other formats as needed
	return fmt.Sprintf("Rule: %s", rule)
}

func atoi(s string) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return 0
}

/*
func main() {
	tzPOSIX := "EST5EDT4,M3.2.0/02:00:00,M11.1.0/02:00:00"
	description, err := HumanReadableTZ(tzPOSIX)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("POSIX TZ Variable: %s\nDescription:\n%s\n", tzPOSIX, description)

    fmt.Println(strings.Repeat("-", 20))

    tzStatic := "UTC0"
	descriptionStatic, errStatic := HumanReadableTZ(tzStatic)
	if errStatic != nil {
		fmt.Println("Error:", errStatic)
		return
	}
    fmt.Printf("POSIX TZ Variable: %s\nDescription:\n%s\n", tzStatic, descriptionStatic)
}
*/
