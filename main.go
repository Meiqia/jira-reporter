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
	ConfigFile  string `yaml:"-"`
	BaseURL     string `yaml:"baseURL"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	Project     string `yaml:"project"`
	IssueType   string `yaml:"issueType"`
	LastUpdated string `yaml:"lastUpdated"`
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
	flag.StringVar(&flags.Project, "project", "", "Jira project")
	flag.StringVar(&flags.IssueType, "issueType", "", "Jira issue type")
	flag.StringVar(&flags.LastUpdated, "lastUpdated", "", "last updated")

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
		return errors.New("need assignee(s)")
	}
	assignee := strings.Join(flags.args, ",")

	if err := completeOptions(&flags.Options); err != nil {
		return err
	}

	reporter, err := NewReporter(flags.Options.BaseURL, flags.Options.Username, flags.Options.Password)
	if err != nil {
		panic(err)
	}

	jql := fmt.Sprintf(
		"project in (%s) AND issuetype in (%s) AND assignee in (%s) AND updatedDate >= %s ORDER BY updated DESC",
		flags.Options.Project, flags.Options.IssueType, assignee, flags.Options.LastUpdated,
	)
	issues, err := reporter.GetIssues(jql)
	if err != nil {
		panic(err)
	}

	report := reporter.GenReport(issues)
	fmt.Println(report.Markdown())

	return nil
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
	if opts.LastUpdated == "" && configOpts.LastUpdated != "" {
		opts.LastUpdated = configOpts.LastUpdated
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
