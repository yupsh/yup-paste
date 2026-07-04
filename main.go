// Command yup-paste is the CLI wrapper around github.com/gloo-foo/cmd-paste.
package main

import (
	clix "github.com/gloo-foo/cli"
	command "github.com/gloo-foo/cmd-paste"
	urf "github.com/urfave/cli/v3"
)

// version is the build version. It defaults to "dev" for local builds and is
// overridden at release time via the linker: -ldflags "-X main.version=<v>".
var version = "dev"

const (
	name           = "paste"
	flagDelimiters = "delimiters"
	flagSerial     = "serial"
)

// synopsis is the multi-line --help usage block; urfave/cli indents it three
// spaces, so the lines stay flush-left.
const synopsis = `paste [OPTIONS] [FILE...]

Write lines consisting of the sequentially corresponding lines from
each FILE, separated by TABs, to standard output.
With no FILE, or when FILE is -, read standard input.`

// spec declares the paste wrapper: a file-or-stdin filter with delimiter and
// serial flags.
var spec = clix.Spec{
	Name:     name,
	Summary:  "merge lines of files",
	Synopsis: synopsis,
	Build:    build,
	Flags:    flags(),
}

// flags builds a fresh set of the wrapper's flags. It is a constructor rather
// than a shared slice so each parse gets its own flag instances (urfave/cli
// retains a flag's set-state on its pointer across runs, which would otherwise
// leak between test invocations).
func flags() []urf.Flag {
	return []urf.Flag{
		&urf.StringFlag{
			Name:    flagDelimiters,
			Aliases: []string{"d"},
			Usage:   "use characters from LIST instead of TABs",
		},
		&urf.BoolFlag{
			Name:    flagSerial,
			Aliases: []string{"s"},
			Usage:   "paste one file at a time instead of in parallel",
		},
	}
}

// build maps the invocation to paste's pipeline: a file-or-stdin source into the
// paste command configured by the delimiter and serial flags.
func build(inv clix.Invocation) (clix.Source, clix.Command, error) {
	return clix.OperandsOrStdin(inv), command.Paste(options(inv.Args)...), nil
}

// options folds the parsed flags into paste's option values.
func options(c *urf.Command) []any {
	var opts []any
	if c.IsSet(flagDelimiters) {
		opts = append(opts, command.PasteDelimiter(c.String(flagDelimiters)))
	}
	if c.Bool(flagSerial) {
		opts = append(opts, command.PasteSerial)
	}
	return opts
}

// runMain is an indirection seam so main's wiring is testable without spawning
// the process; a test swaps it and restores it.
var runMain = clix.Main

func main() { runMain(spec, version) }
