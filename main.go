package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/abiosoft/ishell/v2"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v2"
)

const (
	defaultConfigFile = "~/.jira-reporter/config.yaml"
)

type Options struct {
	ConfigFile     string `yaml:"-"`
	BaseURL        string `yaml:"baseURL"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	Project        string `yaml:"project"`
	IssueType      string `yaml:"issueType"`
	UpdatedSince   string `yaml:"updatedSince"`
	UpdatedBetween string `yaml:"updatedBetween"`
}

type userFlags struct {
	Options

	args []string
}

func main() {
	var flags userFlags
	flag.StringVar(&flags.ConfigFile, "config", defaultConfigFile, "config file in YAML")
	flag.StringVar(&flags.BaseURL, "baseURL", "", "base URL of Jira server")
	flag.StringVar(&flags.Username, "username", "", "Jira username")
	flag.StringVar(&flags.Project, "project", "", "Jira project (comma-separated)")
	flag.StringVar(&flags.IssueType, "issueType", "", "Jira issue type (comma-separated)")
	flag.StringVar(&flags.UpdatedSince, "updatedSince", "", `date after which issues have been updated, e.g. "-7d"`)
	flag.StringVar(&flags.UpdatedBetween, "updatedBetween", "", "date range between which issues have been updated, e.g. \"2021-10-01~2021-10-10\" (precedes `updatedSince`)")

	flag.Usage = func() {
		fmt.Println(`jira-reporter [flags] assignee [assignee2 [assignee3 [...]]]`)
		flag.PrintDefaults()
	}

	flag.Parse()
	flags.args = flag.Args()

	if err := run(flags); err != nil {
		fmt.Fprintln(os.Stderr, err)
		flag.Usage()
		os.Exit(1)
	}
}

func run(flags userFlags) error {
	if len(flags.args) == 0 {
		return errors.New("require assignee(s)")
	}
	assignee := strings.Join(flags.args, ",")

	opts := &flags.Options
	if err := completeOptions(opts); err != nil {
		return err
	}

	reporter, err := NewReporter(opts.BaseURL, opts.Username, opts.Password)
	if err != nil {
		return err
	}

	jql, err := buildJQL(opts, assignee)
	if err != nil {
		return err
	}

	issues, err := reporter.GetIssues(jql)
	if err != nil {
		return err
	}

	report := reporter.GenReport(issues)
	markdown, err := report.Markdown(getDateRange(opts))
	if err != nil {
		return err
	}
	fmt.Println(markdown)

	return nil
}

func buildJQL(opts *Options, assignee string) (jql string, err error) {
	switch {
	case opts.UpdatedBetween != "":
		parts := strings.SplitN(opts.UpdatedBetween, "~", 2)
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid `-updatedBetween`: %q", opts.UpdatedBetween)
		}
		min, max := parts[0], parts[1]
		jql = fmt.Sprintf(
			"project in (%s) AND issuetype in (%s) AND assignee in (%s) AND updatedDate >= %s AND updatedDate <= %s ORDER BY updated DESC",
			opts.Project, opts.IssueType, assignee, min, max,
		)
	case opts.UpdatedSince != "":
		jql = fmt.Sprintf(
			"project in (%s) AND issuetype in (%s) AND assignee in (%s) AND updatedDate >= %s ORDER BY updated DESC",
			opts.Project, opts.IssueType, assignee, opts.UpdatedSince,
		)
	default:
		return "", errors.New("require `-updatedSince` or `-updatedBetween`")
	}
	return jql, nil
}

func completeOptions(opts *Options) (err error) {
	if err := fillOptionsByConfig(opts); err != nil {
		return err
	}
	return completeUsernameAndPassword(opts)
}

func fillOptionsByConfig(opts *Options) error {
	makeError := func(err error) error {
		if opts.ConfigFile == defaultConfigFile {
			fmt.Println(err)
			return nil
		}
		return err
	}

	expandedConfigFile, err := homedir.Expand(opts.ConfigFile)
	if err != nil {
		return makeError(err)
	}

	configFile, err := filepath.Abs(expandedConfigFile)
	if err != nil {
		return makeError(err)
	}

	b, err := ioutil.ReadFile(configFile)
	if err != nil {
		return makeError(err)
	}

	configOpts := new(Options)
	err = yaml.Unmarshal(b, configOpts)
	if err != nil {
		return makeError(err)
	}

	// Merge options.
	if opts.BaseURL == "" && configOpts.BaseURL != "" {
		opts.BaseURL = configOpts.BaseURL
	}
	if opts.Username == "" && configOpts.Username != "" {
		opts.Username = configOpts.Username
	}
	if opts.Password == "" && configOpts.Password != "" {
		opts.Password = configOpts.Password
	}
	if opts.Project == "" && configOpts.Project != "" {
		opts.Project = configOpts.Project
	}
	if opts.IssueType == "" && configOpts.IssueType != "" {
		opts.IssueType = configOpts.IssueType
	}
	if opts.UpdatedSince == "" && configOpts.UpdatedSince != "" {
		opts.UpdatedSince = configOpts.UpdatedSince
	}
	if opts.UpdatedBetween == "" && configOpts.UpdatedBetween != "" {
		opts.UpdatedBetween = configOpts.UpdatedBetween
	}

	return nil
}

func completeUsernameAndPassword(opts *Options) (err error) {
	shell := ishell.New()

	if opts.Username == "" {
		// Show the prompt for Username if not specified.
		shell.Print("Username: ")
		opts.Username, err = shell.ReadLineErr()
		if err != nil {
			return err
		}
	}

	if opts.Password == "" {
		// Show the prompt for Password if not specified.
		shell.Print("Password: ")
		opts.Password, err = shell.ReadPasswordErr()
		if err != nil {
			return err
		}
	}

	return nil
}

func getDateRange(opts *Options) string {
	switch {
	case opts.UpdatedBetween != "":
		return opts.UpdatedBetween
	case opts.UpdatedSince != "":
		return opts.UpdatedSince
	default:
		return ""
	}
}
