package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/pflag"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"github.com/tzlist/rfc9636"
)

const (
	LevelTrace = slog.Level(-8)
	LevelFatal = slog.Level(12) 
)

// Custom Logger methods for Trace and Fatal
func Trace(msg string, args ...any) {
	slog.Log(context.Background(), LevelTrace, msg, args...)
}

func Fatal(msg string, args ...any) {
	slog.Log(context.Background(), LevelFatal, msg, args...)
	os.Exit(1) // Terminate the program after logging
}

// Information returned by time.Zone function
type TzZoneType struct {
	Name string
	Offset int
}

// Offsets [0] standard time
// Offsets [1] daylight savings time
// len[Offsets] > 1 Has daylight savings time

type TzInfoType struct {
	Aliases []string
	Offsets []TzZoneType
	Extend string
}

type TzInfoMap map[string]TzInfoType 

var TzInfos = make(TzInfoMap)

func (tzi TzInfoMap) AddZoneAlias( zone string, alias string) {
	zoneInfo, exists := tzi[zone]
	if !exists {
		zoneInfo = NewTzInfo()
	}
	index, found := slices.BinarySearch(zoneInfo.Aliases, alias)

	if !found {
		zoneInfo.Aliases = slices.Insert(zoneInfo.Aliases, index, alias)
	}
	tzi[zone] = zoneInfo
	return
	
}

func (tzi TzInfoMap) Add( zone string, data *rfc9636.Location) {
	zoneInfo, exists := tzi[zone]
	if !exists {
		zoneInfo = NewTzInfo()
	}

	year := time.Now().Year()
	loc, err := time.LoadLocation(zone)
	if err != nil {
		Fatal("LoadLocation failed", "error", err);

	}
	
	// Check offset on a winter date (Jan 1) and a summer date (Jul 1)
	winterTime := time.Date(year, time.January, 1, 0, 0, 0, 0, loc)
	summerTime := time.Date(year, time.July, 1, 0, 0, 0, 0, loc)

	xst, winterOffset := winterTime.Zone()
	xdt, summerOffset := summerTime.Zone()
	
	zoneInfo.Offsets = append(zoneInfo.Offsets, TzZoneType { xst, winterOffset })

	if winterOffset != summerOffset {
		zoneInfo.Offsets = append(zoneInfo.Offsets, TzZoneType { xdt, summerOffset })
	}

	zoneInfo.Extend = data.Extend()
	tzi[zone] = zoneInfo
}

func NewTzInfo() TzInfoType {
	return TzInfoType {
		Aliases: make([]string, 0),
		Offsets: make([]TzZoneType, 0, 2),
		Extend: "",
	}
}


type TzAliasType map[string][]string

func NewTzAlias() TzAliasType {
	return make(map[string][]string)
}

// Assuming right now that the data is coming in sorted order given the way zoneinfo directory is walked
// Also BinarySearch will return information sorted as well
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

var TzAliases TzAliasType = NewTzAlias()

func main() {
	pflag.FuncP("loglevel", "l", "Set loglevel to trace, debug, info, warning, error or fatal", func(value string) error {
		lv := strings.ToLower(value)
		if strings.HasPrefix("trace", lv) {
			slog.SetLogLoggerLevel(LevelTrace)
                } else if strings.HasPrefix("debug", lv) {
			slog.SetLogLoggerLevel(slog.LevelDebug)
                } else if strings.HasPrefix("info", lv) {
			slog.SetLogLoggerLevel(slog.LevelInfo)
                } else if strings.HasPrefix("warning", lv) {
			slog.SetLogLoggerLevel(slog.LevelWarn)
                } else if strings.HasPrefix("error", lv) {
			slog.SetLogLoggerLevel(slog.LevelError)
                } else if strings.HasPrefix("fatal", lv) {
			slog.SetLogLoggerLevel(LevelFatal)
                } else {
                        return errors.New("The loglevel parameter value must be a prefix of one of theses words, \"trace\", \"debug\", \"info\", \"warning\", \"error\" or \"fatal\".")
                }
		return nil
	})

	pflag.Parse()

	numAliases := 0
	for _, name := range GetOsTimeZones() {
		zone, exist := TzInfos[name]
		if exist {
			fmt.Printf( "%-30s numAliases %-3d numOffsets %-2d Extend %s\n", name, len(zone.Aliases), len(zone.Offsets), zone.Extend)
		} else {
			fmt.Printf( "Missing zone %s\n", zone)
		}
		numAliases += len(zone.Aliases)
	}
	fmt.Println(numAliases + len(TzInfos))
}

// UsesDST checks if a given location observes Daylight Saving Time by comparing offsets.

func UsesDST(timezone string) (bool, string, string, int, error) {
	year := time.Now().Year()
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return false, "", "", year, err
	}

	// Check offset on a winter date (Jan 1) and a summer date (Jul 1)
	winterTime := time.Date(year, time.January, 1, 0, 0, 0, 0, loc)
	summerTime := time.Date(year, time.July, 1, 0, 0, 0, 0, loc)

	xst, winterOffset := winterTime.Zone()
	xdt, summerOffset := summerTime.Zone()

	// If the offsets are different, the timezone uses DST rules.
	return winterOffset != summerOffset, xst, xdt, year, nil
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
		Trace( "zoneinfo directory is not available", "path", path );
		return zones
	}

	// Linux Convention
	//   The zoneinfo names are capitalized.  We can ignore directories and files that do not follow
	//   that convention for now. We might need an exception list in the future if we want to allow
	//   localtime and posixrules zoneinfo files to be processed as well

	for _, info := range dirInfos {
		if info.IsDir() && info.Name() != strings.ToUpper(info.Name()[:1])+info.Name()[1:] {
			Trace("Skipping directory because name is not capitalized ", "filename", info.Name())
			continue
		}

		newPath := path + "/" + info.Name()

		if info.IsDir() {
			zones = walkTzDir(newPath, zones)
		} else {
			parts := strings.Split(newPath, "//")
			if len(parts) != 2 {
				continue
			}
			if zoneInfo, err := rfc9636.LoadLocation(parts[1],[]string{parts[0]}); err == nil {
				slog.Debug("dump of zoneinfo", "timezone", parts[1])
				if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
					rfc9636.DumpLocation(zoneInfo)
				}

				if info.Type()&os.ModeSymlink != 0 {
					symTarget, err := os.Readlink(newPath)
					if err != nil {
						slog.Error("Could not read link target", "path", newPath, "error", err)
						continue
					}
					slog.Debug("source points to target", "source", newPath, "target", symTarget)
					resolvedPath, err := filepath.EvalSymlinks(newPath)
					if err != nil {
						slog.Error("Could not evaluate symlink", "symlink",newPath, "error", err)
						continue
					}
					atz, found := strings.CutPrefix(resolvedPath, parts[0]+"/")
					if !found {
						slog.Error("Could not extract timezone alias", "path", resolvedPath)
						continue
					}
					slog.Debug("Timezone has alias", "timezone", atz, "alias", parts[1])
					TzInfos.AddZoneAlias(atz, parts[1])

					TzAliases.Add(atz, parts[1])
				} else {
					TzInfos.Add(parts[1], zoneInfo)
					zones = append(zones, parts[1])
				}

			} else {
				Trace("File is not a timezone file", "file", newPath);
			}

		}
	}
	return zones
}
