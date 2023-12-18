package cmd

import (

  "github.com/giantswarm/app-migration-cli/cmd/preflight"

  "github.com/giantswarm/microerror"
  "github.com/giantswarm/micrologger"
  "github.com/spf13/cobra"
)

const (
  // CommandUse indicates the general syntax of the command
  CommandUse = "app-migration-cli"

  // CommandShort describes the command in a short list
  CommandShort = "A Giantswarm tool to migrate apps between MCs"

  // CommandLong documents the command in full length
  CommandLong = `Migrate apps.application.giantswarm.io between two MCs`
)

// Config represents the configuration used to create a new root command.
type Config struct {
  // Dependencies.
  Logger micrologger.Logger

  // Settings.
}

type Command struct {
  // Internals.
  cobraCommand *cobra.Command

  // Settings/Preferences
}

// New creates a new root command.
func New(config Config) (*Command, error) {
  var err error

  newCommand := &Command{
    // Internals.
    cobraCommand: nil,
  }

  newCommand.cobraCommand = &cobra.Command{
    Use:               CommandUse,
    Short:             CommandShort,
    Long:              CommandLong,
    RunE:              newCommand.Execute,
  }

  var preflightCommand *preflight.Command
  {
    c := preflight.Config{
      MainCommand: newCommand.cobraCommand,
      Logger: config.Logger,
    }

    preflightCommand, err = preflight.New(c)
    if err != nil {
      return nil, microerror.Mask(err)
    }
  }

  newCommand.cobraCommand.AddCommand(preflightCommand.CobraCommand())

  return newCommand, nil
}

// CobraCommand returns the spf13/cobra command
func (c *Command) CobraCommand() *cobra.Command {
  return c.cobraCommand
}

// Execute is called to actuall run the main command
func (c *Command) Execute(cmd *cobra.Command, args []string) error {
  cmd.HelpFunc()(cmd, nil)
  return nil
}
