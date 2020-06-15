package cmd

import (
	"errors"
	"fmt"
	"log"

	"github.com/mitchellh/go-homedir"
	"github.com/rhd-gitops-example/services/pkg/git"
	"github.com/rhd-gitops-example/services/pkg/promotion"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tcnksm/go-gitconfig"
)

func makePromoteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "promote",
		Short: "promote from one environment to another",
		RunE:  promoteAction,
	}

	// Required flags
	cmd.Flags().String(
		fromFlag,
		"",
		"source Git repository",
	)
	logIfError(cmd.MarkFlagRequired(fromFlag))
	logIfError(viper.BindPFlag(fromFlag, cmd.Flags().Lookup(fromFlag)))

	cmd.Flags().String(
		toFlag,
		"",
		"destination Git repository",
	)
	logIfError(cmd.MarkFlagRequired(toFlag))
	logIfError(viper.BindPFlag(toFlag, cmd.Flags().Lookup(toFlag)))

	cmd.Flags().String(
		serviceFlag,
		"",
		"service name to promote",
	)
	logIfError(cmd.MarkFlagRequired(serviceFlag))
	logIfError(viper.BindPFlag(serviceFlag, cmd.Flags().Lookup(serviceFlag)))

	// Optional flags
	cmd.Flags().String(
		branchNameFlag,
		"",
		"the name of the branch on the destination repository for the pull request (auto-generated if empty)",
	)
	logIfError(viper.BindPFlag(branchNameFlag, cmd.Flags().Lookup(branchNameFlag)))

	cmd.Flags().String(
		cacheDirFlag,
		"~/.promotion/cache",
		"where to cache Git checkouts",
	)
	logIfError(viper.BindPFlag(cacheDirFlag, cmd.Flags().Lookup(cacheDirFlag)))

	cmd.Flags().String(
		emailFlag,
		"",
		"the email to use for commits when creating branches",
	)
	logIfError(viper.BindPFlag(emailFlag, cmd.Flags().Lookup(emailFlag)))

	cmd.Flags().String(
		msgFlag,
		"",
		"the msg to use on the resultant commit and pull request",
	)
	logIfError(viper.BindPFlag(msgFlag, cmd.Flags().Lookup(msgFlag)))

	cmd.Flags().String(
		nameFlag,
		"",
		"the name to use for commits when creating branches",
	)
	logIfError(viper.BindPFlag(nameFlag, cmd.Flags().Lookup(nameFlag)))

	cmd.Flags().Bool(
		debugFlag,
		false,
		"additional debug logging output",
	)
	logIfError(viper.BindPFlag(debugFlag, cmd.Flags().Lookup(debugFlag)))

	cmd.Flags().String(
		fromBranchFlag,
		"master",
		"branch on the source Git repository",
	)
	logIfError(viper.BindPFlag(fromBranchFlag, cmd.Flags().Lookup(fromBranchFlag)))

	cmd.Flags().Bool(
		insecureSkipVerifyFlag,
		false,
		"Insecure skip verify TLS certificate",
	)
	logIfError(viper.BindPFlag(insecureSkipVerifyFlag, cmd.Flags().Lookup(insecureSkipVerifyFlag)))

	cmd.Flags().Bool(
		keepCacheFlag,
		false,
		"whether to retain the locally cloned repositories in the cache directory",
	)
	logIfError(viper.BindPFlag(keepCacheFlag, cmd.Flags().Lookup(keepCacheFlag)))

	cmd.Flags().String(
		repoTypeFlag,
		"github",
		"the type of repository: github, gitlab or ghe",
	)
	logIfError(viper.BindPFlag(repoTypeFlag, cmd.Flags().Lookup(repoTypeFlag)))

	cmd.Flags().String(
		toBranchFlag,
		"master",
		"branch on the destination Git repository",
	)
	logIfError(viper.BindPFlag(toBranchFlag, cmd.Flags().Lookup(toBranchFlag)))
	return cmd
}

func logIfError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func promoteAction(c *cobra.Command, args []string) error {
	// Required flags
	fromRepo := viper.GetString(fromFlag)
	toRepo := viper.GetString(toFlag)
	service := viper.GetString(serviceFlag)

	// Optional flags
	newBranchName := viper.GetString(branchNameFlag)
	msg := viper.GetString(msgFlag)
	debug := viper.GetBool(debugFlag)
	fromBranch := viper.GetString(fromBranchFlag)
	insecureSkipVerify := viper.GetBool(insecureSkipVerifyFlag)
	keepCache := viper.GetBool(keepCacheFlag)
	repoType := viper.GetString(repoTypeFlag)
	toBranch := viper.GetString(toBranchFlag)

	cacheDir, err := homedir.Expand(viper.GetString(cacheDirFlag))
	if err != nil {
		return fmt.Errorf("failed to expand cacheDir path: %w", err)
	}

	author, err := newAuthor()
	if err != nil {
		return fmt.Errorf("unable to establish credentials: %w", err)
	}

	from := promotion.EnvLocation{
		RepoPath: fromRepo,
		Branch:   fromBranch,
	}
	to := promotion.EnvLocation{
		RepoPath: toRepo,
		Branch:   toBranch,
	}

	sm := promotion.New(cacheDir, author, promotion.WithDebug(debug), promotion.WithInsecureSkipVerify(insecureSkipVerify), promotion.WithRepoType(repoType))
	return sm.Promote(service, from, to, newBranchName, msg, keepCache)
}

func newAuthor() (*git.Author, error) {
	name := viper.GetString(nameFlag)
	email := viper.GetString(emailFlag)
	token := viper.GetString(githubTokenFlag)

	var err error
	if name == "" {
		name, err = gitconfig.Username()
		if err != nil {
			return nil, err
		}
	}

	if email == "" {
		email, err = gitconfig.Email()
		if err != nil {
			return nil, err
		}
	}

	// TODO: make this a multierror with both errors?
	if name == "" || email == "" {
		return nil, errors.New("unable to identify user and email for commits")
	}

	return &git.Author{Name: name, Email: email, Token: token}, nil
}
