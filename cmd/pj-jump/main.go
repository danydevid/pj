// Command pj-jump jumps to a registered project.
//
// Invoked by the pj manager as `pj jump <project>`; the manager resolves
// the plugin binary and execs it. See docs/PLUGIN_SPEC.md.
package main

import (
	"os"

	"github.com/danydevid/pj/internal/plugin/jump"
)

func main() {
	os.Exit(jump.Run(os.Args[1:]))
}
