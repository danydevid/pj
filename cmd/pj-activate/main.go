// Command pj-activate activates a project: cd into it and apply its .pjrc env.
package main

import (
	"os"

	"github.com/danydevid/pj/internal/plugin/activate"
)

func main() { os.Exit(activate.Run(os.Args[1:])) }
