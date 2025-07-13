package cmd

import (
	"github.com/spf13/cobra"
)

var options Options

var RootCmd = &cobra.Command{
	Use:   "gh-prtree",
	Short: "Generate Pull Request tree from GitHub repository",
	Long:  "A CLI tool to fetch Pull Requests from GitHub and generate a tree structure.",
	Run: func(cmd *cobra.Command, args []string) {
		ExecutePRTree(options)
	},
}

func Execute() {
	cobra.CheckErr(RootCmd.Execute())
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&options.Repository, "repository", "r", "", "Repository owner/name (defaults to the current working directory's repository)")
	RootCmd.PersistentFlags().StringSliceVarP(&options.Authors, "author", "a", []string{}, "Filter by author (can be specified multiple times: --author foo --author bar)")
	RootCmd.PersistentFlags().BoolVarP(&options.ExcludeDrafts, "exclude-drafts", "d", true, "Exclude draft pull requests")
	RootCmd.PersistentFlags().VarP(&options.Format, "format", "f", "Output format (mermaid|json|graphviz, default \"mermaid\")")
}
