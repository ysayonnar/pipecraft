package models

type ErrorResponse struct {
	Error string `json:"error"`
}

type RunPipelineRequest struct {
	RepositoryUrl string `json:"repository_url"`
	Branch        string `json:"branch"`
	Commit        string `json:"commit,omitempty"`
	Image         string `json:"image"`
}

type RunPipelineResponse struct {
	PipelineId int64 `json:"pipeline_id,omitempty"`
}

type PipelineStatusResponse struct {
	Status string `json:"status"`
}

type Logs struct {
	LogsId        int64  `json:"logs_id"`
	Command       string `json:"command_name"`
	CommandNumber int    `json:"command_number"`
	Results       string `json:"results"`
	FinalStatus   string `json:"final_status"`
	PipelineId    int64  `json:"pipeline_id"`
}
type PipelineLogsResponse struct {
	Logs []Logs `json:"logs"`
}
