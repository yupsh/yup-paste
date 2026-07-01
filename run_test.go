package main

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestRun(t *testing.T) {
	cases := []struct {
		files      map[string]string
		name       string
		version    string
		stdin      string
		wantOut    string
		wantErrSub string
		args       []string
		wantCode   int
	}{
		{
			// Default (parallel) mode on a single stream is a passthrough:
			// each input line is its own output row.
			name:    "stdin parallel passthrough",
			args:    []string{"paste"},
			stdin:   "alpha\nbeta\ngamma\n",
			wantOut: "alpha\nbeta\ngamma\n",
		},
		{
			// Serial mode (-s) collapses the whole stream into one tab-joined
			// row.
			name:    "serial tab join",
			args:    []string{"paste", "-s"},
			stdin:   "alpha\nbeta\ngamma\n",
			wantOut: "alpha\tbeta\tgamma\n",
		},
		{
			// -d only takes effect on serial joins; the delimiter list is
			// cycled byte by byte between lines.
			name:    "serial custom delimiter",
			args:    []string{"paste", "-s", "-d", ","},
			stdin:   "one\ntwo\nthree\n",
			wantOut: "one,two,three\n",
		},
		{
			name:    "single file parallel passthrough",
			args:    []string{"paste", "/in.txt"},
			files:   map[string]string{"/in.txt": "one\ntwo\n"},
			wantOut: "one\ntwo\n",
		},
		{
			// Multiple files concatenate into one stream; serial mode then
			// joins every line of that stream into a single row.
			name: "multiple file sources serial",
			args: []string{"paste", "-s", "/a.txt", "/b.txt"},
			files: map[string]string{
				"/a.txt": "alpha\n",
				"/b.txt": "beta\n",
			},
			wantOut: "alpha\tbeta\n",
		},
		{
			name:    "version flag reports injected version",
			version: "1.2.3",
			args:    []string{"paste", "--version"},
			wantOut: "paste version 1.2.3\n",
		},
		{
			name:       "unknown flag errors",
			args:       []string{"paste", "--nope"},
			wantCode:   1,
			wantErrSub: "paste:",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			for path, content := range tc.files {
				if err := afero.WriteFile(fs, path, []byte(content), 0o644); err != nil {
					t.Fatalf("write fixture %s: %v", path, err)
				}
			}

			var out, errOut bytes.Buffer
			code := run(tc.version, tc.args, strings.NewReader(tc.stdin), &out, &errOut, fs)

			if code != tc.wantCode {
				t.Fatalf("exit code = %d, want %d (stderr=%q)", code, tc.wantCode, errOut.String())
			}
			if tc.wantErrSub == "" && out.String() != tc.wantOut {
				t.Fatalf("stdout = %q, want %q", out.String(), tc.wantOut)
			}
			if tc.wantErrSub != "" && !strings.Contains(errOut.String(), tc.wantErrSub) {
				t.Fatalf("stderr = %q, want substring %q", errOut.String(), tc.wantErrSub)
			}
		})
	}
}

func Test_main(t *testing.T) {
	origExit, origRun := osExit, runCLI
	t.Cleanup(func() { osExit, runCLI = origExit, origRun })

	gotCode := -1
	osExit = func(code int) { gotCode = code }
	runCLI = func(string, []string, io.Reader, io.Writer, io.Writer, afero.Fs) int { return 7 }

	main()

	if gotCode != 7 {
		t.Fatalf("main propagated exit code %d, want 7", gotCode)
	}
}
