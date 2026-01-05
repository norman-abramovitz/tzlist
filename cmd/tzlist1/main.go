package main

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"
	"github.com/spf13/pflag"
)

var Debug bool = false

func main() {
	pflag.BoolVarP(&Debug, "debug", "d", true, "enable debug information")
	// fmt.Println(GetOsTimeZones())
	for _, zone := range GetOsTimeZones() {
		fmt.Println(zone)
	}
}

func GetOsTimeZones() []string {
	var zones []string
	var zoneDirs = []string{
		// Update path according to your OS
		"/usr/share/zoneinfo/",
		"/usr/share/lib/zoneinfo/",
		"/usr/lib/locale/TZ/",
	}

	for _, zd := range zoneDirs {
		zones = walkTzDir(zd, zones)
	}

	return zones
}

func walkTzDir(path string, zones []string) []string {
	dirInfos, err := os.ReadDir(path)
	if err != nil {
		return zones
	}

	isAlpha := func(s string) bool {
		for _, r := range s {
			if !unicode.IsLetter(r) {
				return false
			}
		}
		return true
	}

	for _, info := range dirInfos {
		if info.Name() != strings.ToUpper(info.Name()[:1])+info.Name()[1:] {
			if Debug {
				fmt.Println("Debug:", info.Name(), "!=", strings.ToUpper(info.Name()[:1])+info.Name()[1:])
			}
			continue
		}

		if !isAlpha(info.Name()[:1]) {
			continue
		}

		newPath := path + "/" + info.Name()

		if info.IsDir() {
			zones = walkTzDir(newPath, zones)
		} else {
			parts := strings.Split(newPath, "//")
			if len(parts) == 2 {
				if zoneInfo, err := time.LoadLocation(parts[1]); err == nil {
					if Debug {
					fmt.Printf("Debug: %#v\n", *zoneInfo)
				}
					zones = append(zones, parts[1])
				}
			}
		}
	}

	return zones
}
