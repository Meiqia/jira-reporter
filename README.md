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

### Using Config

Copy [config.yaml](config.yaml) into `~/.jira-exporter`, and fill in your Jira username and password.

```bash
$ jira-exporter <同事1>,<同事2>,...
```

### Using Flags

```bash
$ jira-exporter \
-baseURL=https://jira.meiqia.com -username=<YOUR-JIRA-USERNAME> \
-project="基础架构, LiveChat, 呼叫, 茶馆" \
-issueType="任务, 改进, 故事, 子任务, 故障, Bug" \
-lastUpdated="-7d" \
<同事1>,<同事2>,...
```
