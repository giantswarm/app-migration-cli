package apply

import (
	"errors"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/app-migration-cli/pkg/cluster"

	"github.com/giantswarm/backoff"
)

var (
	flags = &Flags{}
)

const (
	// CommandUse indicates the general syntax of the command
	CommandUse = "apply"

	// CommandShort describes the command in a short list
	CommandShort = "Run the second stage of an app migration"

	// CommandLong documents the command in full length
	CommandLong = `In the apply phase the apps and additional config will
  be read from disk and applied to the newly created capi WC

  Run a migration from gauss to golem:

  ./app-migration-cli apply -f test25-apps.yaml -d golem -n wc1
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

	newCommand.mainCommand.Flags().StringVarP(&flags.sourceFile, "dump-file", "f", "", "Filename that contains the yaml-resources for migration")
	newCommand.mainCommand.Flags().StringVarP(&flags.dstMC, "destination", "d", "", "Name of the destination MC")
	newCommand.mainCommand.Flags().StringVarP(&flags.srcMC, "source", "s", "", "Name of the source MC")
	newCommand.mainCommand.Flags().StringVarP(&flags.wcName, "wc-name", "n", "", "Name of the WC to migrate")
	newCommand.mainCommand.Flags().StringVarP(&flags.orgNamespace, "org-namespace", "o", "", "Namespace of organization in capi, eg. org-foobar")
	newCommand.mainCommand.Flags().BoolVarP(&flags.finalizer, "finalizer", "z", false, "Remove finalizers in the sourceMC. Setting this might result in leftover finalizers")

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
	mcs.BackOff = backoff.NewMaxRetries(15, 3*time.Second)

	err = mcs.ApplyCAPIApps(flags.sourceFile)
	if err != nil {
		if errors.Is(err, cluster.MigrationFileEmpty) {
			color.Red("⚠  Warning")
			color.Red("⚠  No apps targeted for migration")
			color.Red("⚠  The given file was empty")
			color.Red("⚠  Warning")

			return nil
		}

		return microerror.Mask(err)
	}

	if flags.finalizer {
		err = mcs.SrcMC.RemoveFinalizerOnNamespace()
		if err != nil {
			return microerror.Mask(err)
		}
		color.Yellow("Finalizer removed on NS: %s/%s", mcs.SrcMC.Name, mcs.SrcMC.Namespace)
	}

	return nil
}
