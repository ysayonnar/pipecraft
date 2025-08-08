package models

type ErrorResponse struct {
	Error string `json:"error"`
}

type RunPipelineRequest struct {
	RepositoryUrl string `json:"repository_url"`
	Branch        string `json:"branch"`
	Commit        string `json:"commit,omitempty"`
}

type RunPipelineResponse struct {
	PipelineId int64 `json:"pipeline_id,omitempty"`
}

type PipelineStatusResponse struct {
	Status string `json:"status"`
}

type Logs struct {
	LogsId        int64  `json:"logs_id"`
	CommandNumber int    `json:"command_number"`
	CommandName   string `json:"command_name"`
	Command       string `json:"command"`
	Results       string `json:"results"`
	FinalStatus   string `json:"final_status"`
}
type PipelineLogsResponse struct {
	Logs []Logs `json:"logs"`
}
