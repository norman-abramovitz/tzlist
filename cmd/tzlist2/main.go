package main

import (
	"fmt"
	_ "io/ioutil"
	_ "log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"github.com/spf13/pflag"
)
var Debug bool = false

func main() {
	pflag.BoolVarP(&Debug, "debug", "d", true, "enable debug information")
	timezones := GetSystemTimeZones()
	for _, tz := range timezones {
		fmt.Println(tz)
	}
}

// GetSystemTimeZones attempts to find and list all IANA time zones
// available on the current operating system.
func GetSystemTimeZones() []string {
	var zones []string
	// Common paths for the zoneinfo directory
	zoneDirs := []string{
		"/usr/share/zoneinfo/",
		"/usr/lib/zoneinfo/",
		"/usr/share/lib/zoneinfo/",
		// Add Windows paths if needed, though they use a different naming convention
	}

	for _, zoneDir := range zoneDirs {
		if _, err := os.Stat(zoneDir); os.IsNotExist(err) {
			continue // Skip if directory doesn't exist
		}

		// Walk the directory and validate with time.LoadLocation
		filepath.Walk(zoneDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			// Exclude non-timezone files (like 'VERSION' on macOS)
			if strings.Contains(path, "zoneinfo/posix/") ||
			   strings.Contains(path, "zoneinfo/localtime/") ||
			   strings.Contains(path, "zoneinfo/posixrules/") ||
			   strings.Contains(path, "zoneinfo/right/") {
				return nil
			}

			// Get the IANA name (e.g., "America/New_York") relative to the base dir
			tzName := strings.TrimPrefix(path, zoneDir)

			// Validate that Go can actually load this location name
			if _, err := time.LoadLocation(tzName); err == nil {
				zones = append(zones, tzName)
			}
			return nil
		})
	}
	return zones
}
