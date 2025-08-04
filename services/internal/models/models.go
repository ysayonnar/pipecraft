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
