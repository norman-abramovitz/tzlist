package tzposix

import (
	_ "fmt"
	"strings"
	"testing"
)

/*
TODO Add these tests
CET6CEST530,M4.5.0/02:00:00,M10.5.0/03:00:00
HAST10HADT,M4.2.0/03:0:0,M10.2.0/03:0:00
AST9ADT,M3.2.0,M11.1.0
AST9ADT,M3.2.0/03:0:0,M11.1.0/03:0:0
EST5EDT,M3.2.0/02:00:00,M11.1.0/02:00:00
America/Sao_Paulo   GRNLNDST3GRNLNDDT,M10.3.0/00:00:00,M2.4.0/00:00:00
EST5EDT,M3.2.0/02:00:00,M11.1.0
EST5EDT,M3.2.0,M11.1.0/02:00:00
CST6CDT,M3.2.0/2:00:00,M11.1.0/2:00:00
MST7MDT,M3.2.0/2:00:00,M11.1.0/2:00:00
PST8PDT,M3.2.0/2:00:00,M11.1.0/2:00:00
*/

func TestHumanReadableTZ(t *testing.T) {
	var tests = []struct {
		tz          string
		expectSst   string
		expectDst   string
		expectRules string
		expectError error
	}{
		{tz: "EET-2EEST,M4.5.5/0,M10.5.4/24",
			expectSst:   "Standard Time: EET (UTC +02:00)",
			expectDst:   "Daylight Time: EEST (UTC +03:00)",
			expectRules: "Rules: Starts on the last Friday of April at 0, Ends on the last Thursday of October at 24",
			expectError: nil,
		},
		{tz: "PST8PDT,M3.2.0,M11.1.0",
			expectSst:   "Standard Time: PST (UTC -08:00)",
			expectDst:   "Daylight Time: PDT (UTC -07:00)",
			expectRules: "Rules: Starts on the second Sunday of March at 02:00:00, Ends on the first Sunday of November at 02:00:00",
			expectError: nil,
		},
		{tz: "NST-3:30NDT2:30,M3.2.0/2:30:2,M11.1.0/11:25:40",
			expectSst:   "Standard Time: NST (UTC +03:30)",
			expectDst:   "Daylight Time: NDT (UTC -02:30)",
			expectRules: "Rules: Starts on the second Sunday of March at 2:30:2, Ends on the first Sunday of November at 11:25:40",
			expectError: nil,
		},
		{tz: "CST6CDT,M3.2.0/2:00:00,M11.1.0/2:00:00",
			expectSst:   "Standard Time: CST (UTC -06:00)",
			expectDst:   "Daylight Time: CDT (UTC -05:00)",
			expectRules: "Rules: Starts on the second Sunday of March at 2:00:00, Ends on the first Sunday of November at 2:00:00",
			expectError: nil,
		},
		{tz: "GRNLNDST3GRNLNDDT,M10.3.0/00:00:00,M2.4.0/00:00:00",
			expectSst:   "Standard Time: GRNLNDST (UTC -03:00)",
			expectDst:   "Daylight Time: GRNLNDDT (UTC -02:00)",
			expectRules: "Rules: Starts on the third Sunday of October at 00:00:00, Ends on the fourth Sunday of February at 00:00:00",
			expectError: nil,
		},
		{tz: "<+00>0<+01>,0/0,J365/25",
			expectSst:   "Standard Time: <+00> (UTC +00:00)",
			expectDst:   "Daylight Time: <+01> (UTC +01:00)",
			expectRules: "Rules: Starts Rule: 0/0, Ends Rule: J365/25",
			expectError: nil,
		},
	}

	for _, tt := range tests {

		testname := tt.tz
		t.Run(testname, func(t *testing.T) {
			ans, err := HumanReadableTZ(tt.tz)
			if err != nil {
				t.Errorf("got %v, want nil", err)
				return
			}
			parts := strings.Split(ans, "\n")
			if len(parts) != 3 {
				t.Errorf("got %d parts, want 3\nactual %s", len(parts), ans)
				return
			}
			if parts[0] != tt.expectSst {
				t.Errorf("got %s want %s\nactual %s", parts[0], tt.expectSst, ans)
				return
			}
			if parts[1] != tt.expectDst {
				t.Errorf("got %s want %s\nactual %s", parts[1], tt.expectDst, ans)
				return
			}
			if parts[2] != tt.expectRules {
				t.Errorf("got %s want %s\nactual %s", parts[2], tt.expectRules, ans)
				return
			}
		})
	}
}

func TestHumanReadableTZNoDst(t *testing.T) {
	var tests = []struct {
		tz          string
		expectSst   string
		expectDst   string
		expectRules string
		expectError error
	}{
		{tz: "GMT0",
			expectSst:   "Standard Time: GMT (UTC +00:00)",
			expectRules: "(No Daylight Saving Time rules)",
			expectError: nil,
		},
		{tz: "EAT-3",
			expectSst:   "Standard Time: EAT (UTC +03:00)",
			expectRules: "(No Daylight Saving Time rules)",
			expectError: nil,
		},
	}

	for _, tt := range tests {

		testname := tt.tz
		t.Run(testname, func(t *testing.T) {
			ans, err := HumanReadableTZ(tt.tz)
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
