package jobs

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Step struct {
	Name string `yaml:"name"`
	Run  string `yaml:"run"`
}

type Job struct {
	Name  string
	Steps []Step
}

func ParseJobsOrdered(data []byte) ([]Job, error) {
	const op = "jobs.ParseJobsOrdered"

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("op: %s, err: %w", op, err)
	}

	var jobsNode *yaml.Node
	for i := 0; i < len(root.Content[0].Content); i += 2 {
		if root.Content[0].Content[i].Value == "jobs" {
			jobsNode = root.Content[0].Content[i+1]
			break
		}
	}

	if jobsNode == nil {
		return nil, fmt.Errorf("op: %s, err: no jobs found", op)
	}

	var jobs []Job
	for i := 0; i < len(jobsNode.Content); i += 2 {
		jobName := jobsNode.Content[i].Value
		jobBody := jobsNode.Content[i+1]

		var steps []Step
		for j := 0; j < len(jobBody.Content); j += 2 {
			if jobBody.Content[j].Value == "steps" {
				stepsNode := jobBody.Content[j+1]
				for _, stepNode := range stepsNode.Content {
					var step Step
					if err := stepNode.Decode(&step); err != nil {
						return nil, fmt.Errorf("op: %s, err: %w", op, err)
					}
					steps = append(steps, step)
				}
			}
		}

		jobs = append(jobs, Job{
			Name:  jobName,
			Steps: steps,
		})
	}

	return jobs, nil
}
