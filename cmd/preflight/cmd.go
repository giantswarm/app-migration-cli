package preflight

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/app-migration-cli/pkg/apps"
	"github.com/giantswarm/app-migration-cli/pkg/cluster"
)

var (
	flags = &Flags{}
)

const (
	// CommandUse indicates the general syntax of the command
	CommandUse = "preflight"

	// CommandShort describes the command in a short list
	CommandShort = "Check all prerequistes for a smooth app migration"

	// CommandLong documents the command in full length
	CommandLong = `Check if the source MC has apps which can be migrated. Checks
  if the destination MC has the neccessary resources ready to allow running the apps. It operates read-only.

  Check a migration from gauss to golem:

  ./app-migration-cli preflight -s gauss -d golem -n wc1
  `
)

// Config represents the configuration used to create a new command.
type Config struct {
	// Settings.
	MainCommand *cobra.Command
	Logger      micrologger.Logger
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
	color.Yellow("Validating access to both MCs for app migration: %s/%s -> %s\n", flags.srcMC, flags.wcName, flags.dstMC)

	mcs, err := cluster.Login(flags.srcMC, flags.dstMC)
	if err != nil {
		return microerror.Mask(err)
	}
	mcs.WcName = flags.wcName

	color.Green("Access to both MCs validated")

	health, err := mcs.SrcMC.GetWCHealth(mcs.WcName)
	if err != nil {
		return microerror.Mask(err)
	}
	color.Green("WorkloadCluster State is healthy: %s", health)

	apps, err := apps.GetAppCRs(mcs.SrcMC.KubernetesClient, mcs.WcName)
	if err != nil {
		return microerror.Mask(err)
	}
	color.Yellow(". Found %d apps for migration", len(apps))

	return nil
}
