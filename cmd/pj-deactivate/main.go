// Command pj-deactivate reverses pj-activate: unsets env vars recorded in
// PJ_ACTIVATED_KEYS and clears the PJ_PROJECT markers.
package main

import (
	"os"

	"github.com/danydevid/pj/internal/plugin/deactivate"
)

func main() { os.Exit(deactivate.Run(os.Args[1:])) }
