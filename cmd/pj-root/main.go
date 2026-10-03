// Command pj-root walks upward from cwd to the nearest .pjrc and cds there.
package main

import (
	"os"

	"github.com/danydevid/pj/internal/plugin/root"
)

func main() { os.Exit(root.Run(os.Args[1:])) }
