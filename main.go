// Command harpo is a local secret broker for AI coding agents. It grants
// temporary, scoped, auditable access to credentials from existing vaults
// without exposing the vault to the agent. See harpo-mvp-spec.md.
package main

import (
	"os"

	"github.com/harpo-sh/harpo/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
