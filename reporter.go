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

type Status struct {
	Finished map[string]struct{}
	Canceled map[string]struct{}
}

func NewStatus(finishedStatus, canceledStatus string) *Status {
	status := &Status{
		Finished: make(map[string]struct{}),
		Canceled: make(map[string]struct{}),
	}

	add := func(m map[string]struct{}, s string) {
		for _, part := range strings.Split(s, ",") {
			m[strings.TrimSpace(part)] = struct{}{}
		}
	}

	add(status.Finished, finishedStatus)
	add(status.Canceled, canceledStatus)

	return status
}

type Report struct {
	status *Status
	data   map[string]*Progress
}

func NewReport(status *Status) *Report {
	return &Report{
		status: status,
		data:   make(map[string]*Progress),
	}
}

func (r *Report) Add(username string, issue jira.Issue) {
	p, ok := r.data[username]
	if !ok {
		p = new(Progress)
		r.data[username] = p
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

	if _, ok := r.status.Finished[i.Status]; ok {
		p.Finished = append(p.Finished, i)
		return
	}

	if _, ok := r.status.Canceled[i.Status]; !ok {
		// Issues that are not cancelled will be categorized as Unfinished.
		p.Unfinished = append(p.Unfinished, i)
	}
}

func (r *Report) Markdown(dateRange string) (string, error) {
	text := `
# Jira 报告（{{$.DateRange}} 更新过的任务）
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
		DateRange string
		Report    map[string]*Progress
	}{
		DateRange: dateRange,
		Report:    r.data,
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

func (r *Reporter) GenReport(issues []jira.Issue, status *Status) *Report {
	report := NewReport(status)
	for _, issue := range issues {
		username := issue.Fields.Assignee.DisplayName
		report.Add(username, issue)
	}
	return report
}
