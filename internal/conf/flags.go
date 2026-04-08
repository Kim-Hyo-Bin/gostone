package conf

import (
	"flag"
	"os"
)

// ParseFlagsFromArgs parses global flags and returns unconsumed arguments (e.g. subcommands).
func ParseFlagsFromArgs(args []string) (Options, []string) {
	fs := flag.NewFlagSet("gostone", flag.ExitOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		_, _ = os.Stderr.WriteString("Usage:\n  gostone [flags] [version|db_sync|bootstrap|fernet-keygen]\n\n" +
			"  version        print binary version, commit, and Go runtime, then exit.\n" +
			"  db_sync        apply database schema only (like keystone-manage db_sync), then exit.\n" +
			"  bootstrap      migrate + seed admin + catalog (like keystone-manage bootstrap); refuses non-empty user table.\n" +
			"  fernet-keygen  write next numbered Fernet key under a Keystone-style key_repository directory, then exit.\n\nFlags:\n")
		fs.PrintDefaults()
	}

	var opts Options
	fs.StringVar(&opts.ConfigFile, "config-file", "", "path to gostone.conf (overrides default search paths)")
	fs.StringVar(&opts.ConfigFile, "c", "", "shorthand for -config-file")
	fs.StringVar(&opts.ConfigDir, "config-dir", "", "directory of *.conf fragments merged after the main file (like Keystone conf.d)")
	_ = fs.Parse(args)
	return opts, fs.Args()
}
