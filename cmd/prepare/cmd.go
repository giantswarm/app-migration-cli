package prepare

import (
  "errors"

  //	"github.com/fatih/color"
  "github.com/fatih/color"
  "github.com/spf13/cobra"

  "github.com/giantswarm/app-migration-cli/pkg/cluster"
  "github.com/giantswarm/app-migration-cli/pkg/apps"

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
  namespace from deletion by the capi-migration (enabled by default)

  Run a migration from gauss to golem:

  ./app-migration-cli prepare -s gauss -d golem -n wc1 -o org-foobar
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
  newCommand.mainCommand.Flags().StringVarP(&flags.orgNamespace, "org-namespace", "o", "", "Namespace of organization in capi, eg. org-foobar")
  newCommand.mainCommand.Flags().StringVarP(&flags.dumpFile, "output-file", "f", "", "Name of the file where the app/cm dump will be stored")
  newCommand.mainCommand.Flags().BoolVarP(&flags.finalizer, "finalizer", "z", true, "Apply finalizers to the source namespace. Setting this might result in the deletion of the ns during the infrastructre migration")

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

  if flags.finalizer {
    err = mcs.SrcMC.SetFinalizerOnNamespace()
    if err != nil {
      return microerror.Mask(err)
    }
    color.Yellow("Finalizer set on NS: %s-%s", mcs.SrcMC.Name, mcs.SrcMC.Namespace)
  }

  mcs.Apps, err = apps.GetAppCRs(mcs.SrcMC.KubernetesClient, mcs.WcName)
  if err != nil {
    if errors.Is(err, apps.EmptyAppsError) {
      color.Red("⚠  Warning")
      color.Red("⚠  No apps targeted for migration.")
      color.Red("⚠  The migration will continue but no apps.application.giantswarm.io CRs will get transferred")
      color.Red("⚠  Warning")

      return nil
    }

    return microerror.Mask(err)
  }

  err = mcs.DumpApps(flags.dumpFile)
  if err != nil {
    return microerror.Mask(err)
  }

  color.Green("Apps (%d) and config is dumped and migrated to disk: %s", len(mcs.Apps), mcs.AppYamlFile(flags.dumpFile))

  return nil
}
