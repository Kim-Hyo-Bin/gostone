package config

import (
	"flag"
	"os"
)

// ParseFlags registers and parses gostone global flags, returning load options.
// Remaining args are not consumed (subcommands can be added later).
func ParseFlags() Options {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.Usage = func() {
		_, _ = os.Stderr.WriteString("Usage:\n  " + os.Args[0] + " [flags]\n\nFlags:\n")
		fs.PrintDefaults()
	}

	var opts Options
	fs.StringVar(&opts.ConfigFile, "config-file", "", "path to gostone.conf (overrides default search paths)")
	fs.StringVar(&opts.ConfigFile, "c", "", "shorthand for -config-file")
	fs.StringVar(&opts.ConfigDir, "config-dir", "", "directory of *.conf fragments merged after the main file (like Keystone conf.d)")
	_ = fs.Parse(os.Args[1:])
	return opts
}
