package main

import (
  "context"
  "fmt"
  "os"

  "github.com/giantswarm/microerror"
  "github.com/giantswarm/micrologger"

  "github.com/giantswarm/app-migration-cli/cmd"
)

func main() {
  err := mainE(context.Background())
  if err != nil {
    fmt.Fprintf(os.Stderr, "Error: %s\n\nTo increase verbosity, re-run with --level=debug\n", microerror.Pretty(err, true))
    os.Exit(2)
  }
}

func mainE(ctx context.Context) error {
  var err error

  // Create a new logger which is used by all packages.
  var newLogger micrologger.Logger
  {
    newLogger, err = micrologger.New(micrologger.Config{
      IOWriter: os.Stderr,
    })
    if err != nil {
      return microerror.Mask(err)
    }

    // set default log level to "error"
    //newLogger = debug.MustWrapDebugLogger(newLogger, "error")
  }

  var newCommand *cmd.Command
  {
    c := cmd.Config{
      Logger: newLogger,
    }

    newCommand, err = cmd.New(c)
    if err != nil {
      return microerror.Mask(err)
    }
  }

  newCommand.CobraCommand().SilenceErrors = true
  newCommand.CobraCommand().SilenceUsage = true
  newCommand.CobraCommand().CompletionOptions.DisableDefaultCmd = true

  err = newCommand.CobraCommand().Execute()
  if err != nil {
    return microerror.Mask(err)
  }

  return nil
}
