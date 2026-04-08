package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/Kim-Hyo-Bin/gostone/internal/app"
	"github.com/Kim-Hyo-Bin/gostone/internal/bootstrap"
	"github.com/Kim-Hyo-Bin/gostone/internal/buildinfo"
	"github.com/Kim-Hyo-Bin/gostone/internal/conf"
	"github.com/Kim-Hyo-Bin/gostone/internal/token"
)

func main() {
	opts, rest := conf.ParseFlagsFromArgs(os.Args[1:])
	if len(rest) >= 1 && rest[0] == "version" {
		if len(rest) != 1 {
			log.Fatalf("usage: %s version", os.Args[0])
		}
		fmt.Printf("gostone %s\n", buildinfo.Version)
		fmt.Printf("commit %s\n", buildinfo.Commit)
		fmt.Printf("%s\n", runtime.Version())
		return
	}
	if len(rest) >= 1 && rest[0] == "fernet-keygen" {
		if len(rest) != 2 {
			log.Fatalf("usage: %s [flags] fernet-keygen <key_repository_dir>", os.Args[0])
		}
		id, err := token.WriteNextKeystoneFernetKey(rest[1])
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("gostone: wrote Fernet key file %d (restart to use as primary; keep older keys until tokens expire)", id)
		return
	}
	cfg, err := conf.Load(opts)
	if err != nil {
		log.Fatal(err)
	}
	switch {
	case len(rest) == 0:
		if err := app.Run(cfg); err != nil {
			log.Fatal(err)
		}
	case len(rest) >= 1 && rest[0] == "db_sync":
		if len(rest) != 1 {
			log.Fatalf("usage: %s [flags] db_sync", os.Args[0])
		}
		if err := app.DBSync(cfg); err != nil {
			log.Fatal(err)
		}
		log.Print("gostone: database schema synchronized")
	case len(rest) >= 1 && rest[0] == "bootstrap":
		bo, err := bootstrap.ParseBootstrapFlags(rest[1:])
		if err != nil {
			log.Fatal(err)
		}
		if err := app.Bootstrap(cfg, bo); err != nil {
			log.Fatal(err)
		}
		log.Print("gostone: bootstrap complete")
	default:
		log.Fatalf("usage: %s [flags] [version|db_sync|bootstrap [bootstrap-flags...]|fernet-keygen <dir>]", os.Args[0])
	}
}
