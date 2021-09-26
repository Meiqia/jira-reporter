# Jira-Reporter

A simple tool to generate report from Jira for our team.


## Installation

```bash
$ go get -u github.com/Meiqia/jira-reporter
```

Usage:

```bash
$ jira-reporter -h
jira-reporter [flags] assignee [assignee2 [assignee3 [...]]]
  -baseURL string
        base URL of Jira server
  -config string
        config file in YAML (default "~/.jira-reporter/config.yaml")
  -issueType string
        Jira issue type (comma-separated)
  -project string
        Jira project (comma-separated)
  -updatedBetween string
        date range between which issues have been updated, e.g. "2021-10-01~2021-10-10" (precedes updatedSince)
  -updatedSince string
        date after which issues have been updated, e.g. "-7d"
  -username string
        Jira username
```


## Example Usage

### Peter's report for the last 7 days

Copy [config.yaml](config.yaml) into `~/.jira-reporter` and fill in your Jira username and password, then execute:

```bash
$ jira-reporter Peter
```

### Peter's report for the last 2 days

```bash
$ jira-reporter -updatedSince=-2d Peter
```

### Peter's report between 2021-10-01 and 2021-10-10

```bash
$ jira-reporter -updatedBetween=2021-10-01~2021-10-10 Peter
```

### Input password in the prompt

Leave `password` empty in `~/.jira-reporter/config.yaml`, then execute:

```bash
$ jira-reporter Peter
Password: 
```
