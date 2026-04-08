package bootstrap

import (
	"errors"
	"flag"
	"os"
)

// ParseBootstrapFlags parses arguments after the `bootstrap` subcommand (keystone-manage bootstrap–style).
func ParseBootstrapFlags(args []string) (Options, error) {
	fs := flag.NewFlagSet("bootstrap", flag.ExitOnError)
	fs.SetOutput(os.Stderr)
	o := DefaultBootstrapOptions()
	fs.StringVar(&o.AdminPassword, "bootstrap-password", "", "admin user password (or set GOSTONE_BOOTSTRAP_ADMIN_PASSWORD)")
	fs.StringVar(&o.AdminUsername, "bootstrap-username", o.AdminUsername, "admin user name")
	fs.StringVar(&o.DomainName, "bootstrap-domain-name", o.DomainName, "initial domain name")
	fs.StringVar(&o.ProjectName, "bootstrap-project-name", o.ProjectName, "initial project name")
	fs.StringVar(&o.RegionID, "bootstrap-region-id", "", "catalog region id when seeding empty services (default: [service] region_id or RegionOne)")
	fs.StringVar(&o.AdminRoleName, "bootstrap-admin-role", o.AdminRoleName, "name of the admin role")
	_ = fs.Parse(args)

	if o.AdminPassword == "" {
		o.AdminPassword = os.Getenv("GOSTONE_BOOTSTRAP_ADMIN_PASSWORD")
	}
	if o.AdminPassword == "" {
		return o, errors.New("bootstrap password required: use --bootstrap-password or GOSTONE_BOOTSTRAP_ADMIN_PASSWORD")
	}
	return o, nil
}
