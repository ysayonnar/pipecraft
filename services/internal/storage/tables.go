package storage

import "time"

type PipelinesTable struct {
	PipelineId int64
	Status     string
	Repository string
	Branch     string
	Commit     string
	CreatedAt  time.Time
}

type LogsTable struct {
	LogId         int64
	CommandNumber int
	CommandName   string
	Command       string
	Results       string
	FinalStatus   string
	PipelineId    int64
}
