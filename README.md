# Jira-Reporter

A simple tool to generate report from Jira for our team.

## Installation

```bash
$ go get -u github.com/Meiqia/jira-exporter
```

Help:

```bash
$ jira-exporter -h
jira-reporter [flags] assignee [assignee2 [assignee3 [...]]]
  -config string
        config file in YAML (default "~/.jira-reporter/config.yaml")
  -host string
        base URL of Jira server
  -issueType string
        Jira issue type
  -lastUpdated string
        last updated
  -project string
        Jira project
  -user string
        Jira username
```

## Example Usage

### Peter's report for the last 7 days

Copy [config.yaml](config.yaml) into `~/.jira-exporter` and fill in your Jira username and password, then execute:

```bash
$ jira-exporter Peter
```

### Peter's report for the last 2 days

```bash
$ jira-exporter -lastUpdated=-2d Peter
```

### Input password in the prompt

Leave `password` empty in `~/.jira-exporter/config.yaml`, then execute:

```bash
$ jira-exporter Peter
Password: 
```
