package prepare

import (
	//	"github.com/fatih/color"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/giantswarm/app-migration-cli/pkg/cluster"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

var (
  flags = &Flags{}
)

const (
  // CommandUse indicates the general syntax of the command
  CommandUse = "prepare"

  // CommandShort describes the command in a short list
  CommandShort = "Run the first stage of an app migration"

  // CommandLong documents the command in full length
  CommandLong = `In the preparation phase the apps and additional config will
be written to disk and finalizers will protect the
namespace from deletion thourgh capi-migration (enabled by default)

Run a migration from gauss to golem:

  ./app-migration-cli prepare -s gauss -d golem -n wc1
  `
)

// Config represents the configuration used to create a new command.
type Config struct {
  // Settings.
  MainCommand *cobra.Command
  Logger micrologger.Logger
}

type Command struct {
  // Dependencies.
  logger micrologger.Logger

  // Settings.
  mainCommand *cobra.Command
}

// New creates a new configured command.
func New(config Config) (*Command, error) {
  // Dependencies.
  if config.Logger == nil {
    return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
  }

  newCommand := &Command{
    // Dependencies.
    logger: config.Logger,

    // Internals.
    mainCommand: nil,
  }

  newCommand.mainCommand = &cobra.Command{
    Use:   CommandUse,
    Short: CommandShort,
    Long:  CommandLong,
    RunE:  newCommand.Execute,
  }

  newCommand.mainCommand.Flags().StringVarP(&flags.srcMC, "source", "s", "", "Name of the source MC")
  newCommand.mainCommand.Flags().StringVarP(&flags.dstMC, "destination", "d", "", "Name of the destination MC")
  newCommand.mainCommand.Flags().StringVarP(&flags.wcName, "wc-name", "n", "", "Name of the WC to migrate")
  newCommand.mainCommand.Flags().StringVarP(&flags.orgNamespace, "org-name", "o", "", "Name of organization Namespace in capi, eg. org-foobar")
  newCommand.mainCommand.Flags().BoolVarP(&flags.noFinalizer, "no-finalizer", "f", false, "Apply no finalizers to the source namespace/cluster. This might result in the deletion of the ns")

  return newCommand, nil
}

func (c *Command) CobraCommand() *cobra.Command {
  return c.mainCommand
}

func (c *Command) Execute(cmd *cobra.Command, args []string) error {

  err := flags.Validate()
  if err != nil {
    return microerror.Mask(err)
  }

  //c.logger = debug.MustWrapDebugLogger(c.logger, "error")

  err = c.execute()
  if err != nil {
    return microerror.Mask(err)
  }

  return nil
}

func (c *Command) execute() error {
  mcs, err := cluster.Login(flags.srcMC, flags.dstMC)
  if err != nil {
    return microerror.Mask(err)
  }
  mcs.WcName = flags.wcName
  mcs.SrcMC.Namespace = flags.wcName
  mcs.OrgNamespace = flags.orgNamespace

  if ! flags.noFinalizer {
    err = mcs.SrcMC.SetFinalizerOnNamespace()
    if err != nil {
      return microerror.Mask(err)
    }
    color.Yellow("Finalizer set on NS: %s-%s", mcs.SrcMC.Name, mcs.SrcMC.Namespace)
  }

  mcs.Apps, err = mcs.FetchApps()
  if err != nil {
    return microerror.Mask(err)
  }

  err = mcs.DumpApps()
  if err != nil {
    return microerror.Mask(err)
  }

  // todo: remove after successfull migration
  //if ! flags.noFinalizer {
  //  mcs.SrcMC.RemoveFinalizerOnNamespace()
  //  if err != nil {
  //    return microerror.Mask(err)
  //  }
 // }


  return nil
}
