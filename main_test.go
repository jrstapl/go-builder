package main

import (
	"os"
	"runtime"
	"slices"
	"strings"
	"testing"
)

var currentWD, _ = os.Getwd()

// just example input, not taken from true
// go tool dist list
var testingDists = []GoDist{
	GoDist{
		GOOS:         "windows",
		GOARCH:       "x86",
		CgoSupported: true,
		FirstClass:   true,
	},
	GoDist{
		GOOS:         "darwin",
		GOARCH:       "arm64",
		CgoSupported: true,
		FirstClass:   true,
	},
	GoDist{
		GOOS:         "linux",
		GOARCH:       "x86",
		CgoSupported: true,
		FirstClass:   true,
	},
	GoDist{
		GOOS:         "linux",
		GOARCH:       "arm64",
		CgoSupported: true,
		FirstClass:   true,
	},
	GoDist{
		GOOS:         "bsd",
		GOARCH:       "arm64",
		CgoSupported: true,
		FirstClass:   false,
	},
}

func TestGetTargetBuilds(t *testing.T) {
	testCases := []struct {
		name    string
		targets []OSARCH
		dists   []GoDist
		wants   []GoDist
	}{
		{
			name: "windows only",
			targets: []OSARCH{
				OSARCH{
					OS:   "windows",
					ARCH: "",
				},
			},
			dists: testingDists,
			wants: []GoDist{
				GoDist{
					GOOS:         "windows",
					GOARCH:       "x86",
					CgoSupported: true,
					FirstClass:   true,
				},
			},
		},
		{
			name: "linux only",
			targets: []OSARCH{
				OSARCH{
					OS:   "linux",
					ARCH: "",
				},
			},
			dists: testingDists,
			wants: []GoDist{
				GoDist{
					GOOS:         "linux",
					GOARCH:       "x86",
					CgoSupported: true,
					FirstClass:   true,
				},
				GoDist{
					GOOS:         "linux",
					GOARCH:       "arm64",
					CgoSupported: true,
					FirstClass:   true,
				},
			},
		},
		{
			name: "linux x86",
			targets: []OSARCH{
				OSARCH{
					OS:   "linux",
					ARCH: "x86",
				},
			},
			dists: testingDists,
			wants: []GoDist{
				GoDist{
					GOOS:         "linux",
					GOARCH:       "x86",
					CgoSupported: true,
					FirstClass:   true,
				},
			},
		},
		{
			name:    "empty targets",
			targets: []OSARCH{},
			dists:   testingDists,
			wants:   testingDists,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := getTargetBuilds(tc.targets, tc.dists)

			// could be in one line but it will end up wrapping anyway
			cmp := slices.CompareFunc(res, tc.wants, func(a GoDist, b GoDist) int {
				if a.GOOS == b.GOOS {
					if a.GOARCH == b.GOARCH {
						if a.CgoSupported == b.CgoSupported {
							if a.FirstClass == b.FirstClass {
								return 0
							}
							return -4
						}
						return -3
					}

					return -2

				}

				return -1
			})

			if cmp != 0 {
				t.Logf("Incorrect target dist returned (cmp: %d), wanted:\n%v\ngot:\n%v\n", cmp, tc.wants, res)
				t.Fail()
			}
		})
	}
}

func TestParseStringToOSARCH(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		wants OSARCH
		err   error
	}{
		{
			name:  "windows/x86",
			input: "windows/x86",
			wants: OSARCH{OS: "windows", ARCH: "x86"},
			err:   nil,
		},
		{
			name:  "WINDOWS/X86",
			input: "WINDOWS/x86",
			wants: OSARCH{OS: "windows", ARCH: "x86"},
			err:   nil,
		},
		{
			name:  "windows",
			input: "windows",
			wants: OSARCH{OS: "windows", ARCH: ""},
			err:   nil,
		},
		{
			name:  "blank",
			input: "",
			wants: OSARCH{OS: "", ARCH: ""},
			err:   ErrInvalidOSARCH,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := parseStringToOSARCH(tc.input)

			if res != tc.wants {
				t.Logf("Incorrect OSARCH formulated, wanted: %v got: %v\n", tc.wants, res)
				t.Fail()
			} else if err != tc.err {
				t.Logf("Incorrect error returned, wanted: %v got: %v\n", tc.err, err)
				t.Fail()
			}
		})
	}

}

func TestGetProjectName(t *testing.T) {
	windowsPath := "C:/Users/username/projects/myproject"
	unixPath := "/usr/home/username/projects/myproject"
	if runtime.GOOS == "windows" {
		windowsPath = strings.ReplaceAll(windowsPath, "/", "\\")
		unixPath = strings.ReplaceAll(unixPath, "/", "\\")
	}

	testCases := []struct {
		name  string
		input string
		wants string
		err   error
	}{
		{
			name:  "Windows example",
			input: `C:/Users/username/projects/myproject`,
			wants: "myproject",
			err:   nil,
		},
		{
			name:  "Unix-style example",
			input: "/usr/home/username/projects/myproject",
			wants: "myproject",
			err:   nil,
		},
		{
			name:  "current dir (.)",
			input: ".",
			wants: currentWD,
			err:   nil,
		},
	}

	for _, tc := range testCases {
		res, err := getProjectName(tc.input)

		if res != tc.wants {
			t.Logf("Incorrect project name formulated, wanted: %v got: %v\n", tc.wants, res)
			t.Fail()
		} else if err != tc.err {
			t.Logf("Incorrect error returned, wanted: %v got: %v\n", tc.err, err)
			t.Fail()
		}
	}

}
