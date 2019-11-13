package create

import (
	"github.com/jenkins-x/jx/pkg/cmd/helper"
	"github.com/jenkins-x/jx/pkg/cmd/importcmd"
	"github.com/spf13/cobra"

	"github.com/jenkins-x/jx/pkg/cmd/opts"
	"github.com/jenkins-x/jx/pkg/cmd/templates"
)

// CreateOptions contains the command line options
type CreateOptions struct {
	*opts.CommonOptions

	DisableImport bool
	OutDir        string
}

// CreateProjectOptions contains the command line options
type CreateProjectOptions struct {
	importcmd.ImportOptions

	DisableImport bool
	OutDir        string
}

var (
	create_resources = `Valid resource types include:

	* archetype
	* cluster
	* env
	* git
	* spring (aka 'springboot')
    `

	create_long = templates.LongDesc(`
		Creates a new resource.

		` + create_resources + `
`)
)

// NewCmdCreate creates a command object for the "create" command
func NewCmdCreate(commonOpts *opts.CommonOptions) *cobra.Command {
	options := &CreateOptions{
		CommonOptions: commonOpts,
	}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new resource",
		Long:  create_long,
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			helper.CheckErr(err)
		},
	}

	cmd.AddCommand(NewCmdCreateAddon(commonOpts))
	cmd.AddCommand(NewCmdCreateArchetype(commonOpts))
	cmd.AddCommand(NewCmdCreateBranchPattern(commonOpts))
	cmd.AddCommand(NewCmdCreateCamel(commonOpts))
	cmd.AddCommand(NewCmdCreateChat(commonOpts))
	cmd.AddCommand(NewCmdCreateCodeship(commonOpts))
	cmd.AddCommand(NewCmdCreateCluster(commonOpts))
	cmd.AddCommand(NewCmdCreateDevPod(commonOpts))
	cmd.AddCommand(NewCmdCreateDockerAuth(commonOpts))
	cmd.AddCommand(NewCmdCreateDocs(commonOpts))
	cmd.AddCommand(NewCmdCreateDomain(commonOpts))
	cmd.AddCommand(NewCmdCreateEnv(commonOpts))
	cmd.AddCommand(NewCmdCreateEtcHosts(commonOpts))
	cmd.AddCommand(NewCmdCreateGkeServiceAccount(commonOpts))
	cmd.AddCommand(NewCmdCreateGit(commonOpts))
	cmd.AddCommand(NewCmdCreateIssue(commonOpts))
	cmd.AddCommand(NewCmdCreateJenkins(commonOpts))
	cmd.AddCommand(NewCmdCreateJHipster(commonOpts))
	cmd.AddCommand(NewCmdCreateLile(commonOpts))
	cmd.AddCommand(NewCmdCreateMicro(commonOpts))
	cmd.AddCommand(NewCmdCreatePostPreviewJob(commonOpts))
	cmd.AddCommand(NewCmdCreateProject(commonOpts))
	cmd.AddCommand(NewCmdCreatePullRequest(commonOpts))
	cmd.AddCommand(NewCmdCreateQuickstart(commonOpts))
	cmd.AddCommand(NewCmdCreateQuickstartLocation(commonOpts))
	cmd.AddCommand(NewCmdCreateMLQuickstart(commonOpts))
	cmd.AddCommand(NewCmdCreateSpring(commonOpts))
	cmd.AddCommand(NewCmdCreateStep(commonOpts))
	cmd.AddCommand(NewCmdCreateTeam(commonOpts))
	cmd.AddCommand(NewCmdCreateTerraform(commonOpts))
	cmd.AddCommand(NewCmdCreateToken(commonOpts))
	cmd.AddCommand(NewCmdCreateTracker(commonOpts))
	cmd.AddCommand(NewCmdCreateUser(commonOpts))
	cmd.AddCommand(NewCmdCreateVault(commonOpts))
	cmd.AddCommand(NewCmdCreateVariable(commonOpts))

	return cmd
}

// Run implements this command
func (o *CreateOptions) Run() error {
	return o.Cmd.Help()
}

// DoImport imports the project created at the given directory
func (o *CreateProjectOptions) ImportCreatedProject(outDir string) error {
	if o.DisableImport {
		return nil
	}
	importOptions := &o.ImportOptions
	importOptions.Dir = outDir
	importOptions.DisableDotGitSearch = true
	return importOptions.Run()
}

func (o *CreateProjectOptions) addCreateAppFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&o.DisableImport, "no-import", "", false, "Disable import after the creation")
	cmd.Flags().StringVarP(&o.OutDir, opts.OptionOutputDir, "o", "", "Directory to output the project to. Defaults to the current directory")

	o.AddImportFlags(cmd, true)
}
