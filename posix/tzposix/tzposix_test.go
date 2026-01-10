package tzposix

import (
    _ "fmt"
    "strings"
    "testing"
)

func TestHumanReadableTZ(t *testing.T) {
    var tests = []struct {
        tz string
        expectSst string
        expectDst string
	expectRules string
	expectError error
    }{
	    {tz: "EET-2EEST,M4.5.5/0,M10.5.4/24",
	    expectSst:  "Standard Time: EET (UTC +02:00)",
	    expectDst:  "Daylight Time: EEST (UTC +03:00)",
	    expectRules: "Rules: Starts on the last Friday of April at 0, Ends on the last Thursday of October at 24",
	    expectError: nil,
	    },
	    {tz: "PST8PDT,M3.2.0,M11.1.0",
	    expectSst:  "Standard Time: PST (UTC -08:00)",
	    expectDst:  "Daylight Time: PDT (UTC -07:00)",
	    expectRules: "Rules: Starts on the second Sunday of March at 02:00:00, Ends on the first Sunday of November at 02:00:00",
	    expectError: nil,
	    },
    }

    for _, tt := range tests {

        testname := tt.tz
        t.Run(testname, func(t *testing.T) {
            ans,err := HumanReadableTZ(tt.tz)
	    if err != nil {
                t.Errorf("got %v, want nil", err)
		return
	    }
	    parts := strings.Split(ans, "\n")
	    if len(parts) != 3 {
                t.Errorf("got %d parts, want 3", len(parts))
		return
	    }
	    if parts[0] != tt.expectSst {
                t.Errorf("got %s parts, want %s", parts[0], tt.expectSst)
		return
	    }
	    if parts[1] != tt.expectDst {
                t.Errorf("got %s parts, want %s", parts[1], tt.expectDst)
		return
	    }
	    if parts[2] != tt.expectRules {
                t.Errorf("got %s parts, want %s", parts[2], tt.expectRules)
		return
	    }
        })
    }
}

func TestHumanReadableTZNoDst(t *testing.T) {
    var tests = []struct {
        tz string
        expectSst string
        expectDst string
	expectRules string
	expectError error
    }{
	    {tz: "GMT0",
	    expectSst:  "Standard Time: GMT (UTC +00:00)",
	    expectRules: "(No Daylight Saving Time rules)",
	    expectError: nil,
	    },
	    {tz: "EAT-3",
	    expectSst:  "Standard Time: EAT (UTC +03:00)",
	    expectRules: "(No Daylight Saving Time rules)",
	    expectError: nil,
	    },
    }

    for _, tt := range tests {

        testname := tt.tz
        t.Run(testname, func(t *testing.T) {
            ans,err := HumanReadableTZ(tt.tz)
	    if err != nil {
                t.Errorf("got %v, want nil", err)
		return
	    }
	    parts := strings.Split(ans, "\n")
	    if len(parts) != 2 {
                t.Errorf("got %d parts, want 2", len(parts))
		return
	    }
	    if parts[0] != tt.expectSst {
                t.Errorf("got %s parts, want %s", parts[0], tt.expectSst)
		return
	    }
	    if parts[1] != tt.expectRules {
                t.Errorf("got %s parts, want %s", parts[1], tt.expectRules)
		return
	    }
        })
    }
}
