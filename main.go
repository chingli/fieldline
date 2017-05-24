package main

import "stj/fieldline/cmd"

func main() {
	cmd.Execute()
}

/*
import (
	"flag"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var usage string = `Fieldline is a tool for generating various field lines from discrete physical quantities.

Usage:

        fieldline command [arguments]

The commands are:

        server      run web server of fieldline
        scalar      visualization from a scalar field
        vector      visualization from a vector field
        tensor      visualization from a tensor field

Use "fieldline help [command]" for more information about a command.

Additional help topics:

        streamline         description of streamline
        hyperstreamline    description of hyperstreamline
        contourline        description of contourline

Use "fieldline help [topic]" for more information about that topic.`

func main() {
	flag.Parse()
	if len(os.Args) == 1 {
		fmt.Printf(usage)
	}
	if len(os.Args) >= 2 {
		if os.Args[1] == "server" {
			fmt.Printf("Starting a web server...\n")
		}
		if os.Args[1] == "tensor" {
			fmt.Printf("Please input the path of a exported discrete tensor field data file:\n")
		}
	}
	os.Exit(0)
}
*/
