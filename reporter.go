package main

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/andygrunwald/go-jira"
)

type Issue struct {
	Key         string
	Summary     string
	Status      string
	Unfinished  bool
	LastComment string
}

type Progress struct {
	Finished   []*Issue
	Unfinished []*Issue
}

type Report map[string]*Progress

func (r Report) Add(username string, issue jira.Issue) {
	p, ok := r[username]
	if !ok {
		p = new(Progress)
		r[username] = p
	}

	i := &Issue{
		Key:     issue.Key,
		Summary: issue.Fields.Summary,
		Status:  issue.Fields.Status.Name,
	}
	if issue.Fields.Comments != nil {
		comments := issue.Fields.Comments.Comments
		if len(comments) > 0 {
			c := comments[len(comments)-1]
			i.LastComment = strings.ReplaceAll(c.Body, "\r\n", " ")
		}
	}

	switch i.Status {
	case "关闭", "完成":
		p.Finished = append(p.Finished, i)
	case "取消":
	default:
		p.Unfinished = append(p.Unfinished, i)
	}

}

func (r Report) Markdown(recentDays int) (string, error) {
	text := `
# Jira 报告（最近 {{$.RecentDays}} 天更新过的任务）
{{- range $username, $progress:= $.Report}}

## {{$username}}

- 已完成:

    {{- range $progress.Finished}}
    + [{{.Summary}}](https://jira.meiqia.com/browse/{{.Key}}): {{.Status}}
    {{- end}}

- 未完成:

    {{- range $progress.Unfinished}}
    + [{{.Summary}}](https://jira.meiqia.com/browse/{{.Key}}): {{.Status}} ({{.LastComment}})
    {{- end}}

{{- end}}
`

	tmpl, err := template.New("").Parse(text)
	if err != nil {
		return "", err
	}

	data := struct {
		RecentDays int
		Report     Report
	}{
		RecentDays: recentDays,
		Report:     r,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

type Reporter struct {
	client *jira.Client
}

func NewReporter(baseURL, username, password string) (*Reporter, error) {
	tp := jira.BasicAuthTransport{
		Username: username,
		Password: password,
	}

	client, err := jira.NewClient(tp.Client(), baseURL)
	if err != nil {
		return nil, err
	}

	return &Reporter{
		client: client,
	}, nil
}

func (r *Reporter) GetIssues(jql string) ([]jira.Issue, error) {
	issues, _, err := r.client.Issue.Search(jql, &jira.SearchOptions{
		MaxResults: 100,
		Fields:     []string{"assignee", "summary", "status", "comment"},
	})
	if err != nil {
		return nil, err
	}
	return issues, nil
}

func (r *Reporter) GenReport(issues []jira.Issue) Report {
	report := Report(make(map[string]*Progress))
	for _, issue := range issues {
		username := issue.Fields.Assignee.DisplayName
		report.Add(username, issue)
	}
	return report
}
