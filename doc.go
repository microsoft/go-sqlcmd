// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package main

/*
Main package (main.go) is the entry point for the sqlcmd CLI application.

To follow the flow of this code:

 1. enter through main.go, (TEMPORARY: decision made whether to invoke the modern
    cobra CLI
 2. Then cmd/cmd.go, see the init() func `New` the `Root` cmd (and all its
    subcommands)
 3. The command-line is then parsed and internal.Initialize() runs (with
    the logging level, config file path, error handling and trace support passed
    into internal packages)
 4. Now go to the cmd/root/… folder structure, and read the DefineCommand
    function for the command (sqlcmd install, sqlcmd query etc.) being run
 5. Each cmd/root/... command has a `run` method that performs the action
 6. All the commands (cmd/root/…) use the /internal packages to abstract from error
    handling and trace (non-localized) logging (as can be seen from the `import`
    for each command (in /cmd/root/...)).

This code follows the Go Style Guide

  - https://google.github.io/styleguide/go/
  - https://go.dev/doc/effective_go
  - https://github.com/golang-standards/project-layout

Exceptions to Go Style Guide:

  - None
*/
