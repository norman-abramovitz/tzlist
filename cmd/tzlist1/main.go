package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode"
)

type TzAliasType map[string][]string

func NewTzAlias() TzAliasType {
	return make(map[string][]string)
}

func (tza TzAliasType) Add(key, value string) {
	slice, exists := tza[key]
	if !exists {
		tza[key] = []string{value}
		return
	}

	index, found := slices.BinarySearch(slice, value)

	if !found {
		tza[key] = slices.Insert(slice, index, value)
	}
}

var Debug bool = false
var TzAliases TzAliasType = NewTzAlias()

func main() {
	pflag.BoolVarP(&Debug, "debug", "d", false, "enable debug information")
	// fmt.Println(GetOsTimeZones())
	for _, zone := range GetOsTimeZones() {
		aliases, exist := TzAliases[zone]
		if exist {
			fmt.Printf("%s has %d aliases %+v\n", zone, len(aliases), aliases)
		} else {
			fmt.Println(zone)
		}
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

		// /usr/share/zoneinfo//Africa/Asmera -> points to Asmara
		// path: /usr/share/zoneinfo//Africa newPath /usr/share/zoneinfo//Africa/Asmera ResovedPath: /usr/share/zoneinfo/Africa/Asmara

		if info.IsDir() {
			zones = walkTzDir(newPath, zones)
		} else {
			parts := strings.Split(newPath, "//")
			if len(parts) != 2 {
				continue
			}
			if zoneInfo, err := time.LoadLocation(parts[1]); err == nil {
				if Debug {
					fmt.Printf("Debug: %#v\n", *zoneInfo)
				}

				if info.Type()&os.ModeSymlink != 0 {
					symTarget, err := os.Readlink(newPath)
					if err != nil {
						fmt.Printf("Could not read link target for %s: %v\n", newPath, err)
						continue
					}
					if Debug {
						fmt.Printf("%s -> points to %s\n", newPath, symTarget)
					}
					resolvedPath, err := filepath.EvalSymlinks(newPath)
					if err != nil {
						fmt.Printf("Could not eval symlink for %s: %v\n", newPath, err)
						continue
					}
					atz, found := strings.CutPrefix(resolvedPath, parts[0]+"/")
					if !found {
						fmt.Printf("Could not extract timezone alias from %s\n", resolvedPath)
						continue
					}
					if Debug {
						fmt.Printf("TIMEZONE %s alias: %s\n", atz, parts[1])
					}
					TzAliases.Add(atz, parts[1])
				} else {
					zones = append(zones, parts[1])
				}
			}
		}
	}
	return zones
}
