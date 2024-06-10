package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/jmattheis/goverter"
	"github.com/jmattheis/goverter/config"
	"github.com/jmattheis/goverter/enum"
)

type Strings []string

func (s Strings) String() string {
	return fmt.Sprint([]string(s))
}

func (s *Strings) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func Parse(args []string) (Command, error) {
	if len(args) == 0 {
		return nil, usageErr("invalid args", "unknown")
	}
	cmd := args[0]

	fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
	fs.Usage = func() {}
	fs.SetOutput(io.Discard)

	if err := fs.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return &Help{Usage: usage(cmd)}, nil
		}
		return nil, usageErr(err.Error(), cmd)
	}

	subArgs := fs.Args()
	if len(subArgs) == 0 {
		return nil, usageErr("missing command", cmd)
	}

	switch subArgs[0] {
	case "gen":
		return parseGen(cmd, subArgs[1:])
	case "version":
		return &Version{}, nil
	case "help":
		return &Help{Usage: usage(cmd)}, nil
	default:
		return nil, usageErr("unknown command "+subArgs[0], cmd)
	}
}

func parseGen(cmd string, args []string) (Command, error) {
	fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}

	var global Strings
	fs.Var(&global, "global", "")
	fs.Var(&global, "g", "")

	buildTags := fs.String("build-tags", "goverter", "")
	outputConstraint := fs.String("output-constraint", "!goverter", "")
	cwd := fs.String("cwd", "", "")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return &Help{Usage: usage(cmd)}, nil
		}
		return nil, usageErr(err.Error(), cmd)
	}

	patterns := fs.Args()

	if len(patterns) == 0 {
		return nil, usageErr("missing PATTERN", cmd)
	}

	c := goverter.GenerateConfig{
		PackagePatterns:       patterns,
		BuildTags:             *buildTags,
		OutputBuildConstraint: *outputConstraint,
		WorkingDir:            *cwd,
		EnumTransformers:      map[string]enum.Transformer{},
		Global: config.RawLines{
			Lines:    global,
			Location: "command line (-g, -global)",
		},
	}
	return &Generate{Config: &c}, nil
}

func usageErr(err, cmd string) error {
	return fmt.Errorf("Error: %s\n%s", err, usage(cmd))
}

func usage(cmd string) string {
	return fmt.Sprintf(`Usage:
  %s gen [OPTIONS] PACKAGE...
  %s help
  %s version

PACKAGE(s):
  Define the import paths goverter will use to search for converter interfaces.
  You can define multiple packages and use the special ... golang pattern to
  select multiple packages. See $ go help packages

OPTIONS:
  -build-tags [tags]: (default: goverter)
      a comma-separated list of additional build tags to consider satisfied
      during the loading of conversion interfaces. See 'go help buildconstraint'.
      Can be disabled by supplying an empty string.

  -cwd [value]:
      set the working directory

  -g [value], -global [value]:
      apply settings to all defined converters. For a list of available
      settings see: https://goverter.jmattheis.de/reference/settings

  -output-constraint [constraint]: (default: !goverter)
      A build constraint added to all files generated by goverter.
      Can be disabled by supplying an empty string.

Examples:
  %s gen ./example/simple ./example/complex
  %s gen ./example/...
  %s gen github.com/jmattheis/goverter/example/simple
  %s gen -g 'ignoreMissing no' -g 'skipCopySameType' ./simple

Documentation:
  Full documentation is available here: https://goverter.jmattheis.de`, cmd, cmd, cmd, cmd, cmd, cmd, cmd)
}
