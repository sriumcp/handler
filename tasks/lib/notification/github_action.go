package notification

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/handler/tasks"
)

const (
	// GHWorkflowTaskName is the name of the GitHub action request task
	GHWorkflowTaskName string = "github-workflow"
	// DefaultRef is default github reference (branch)
	DefaultRef string = "master"
)

// GHWorkflowInputs contain the name and arguments of the task.
type GHWorkflowInputs struct {
	Repository string                `json:"repository" yaml:"repository"`
	Workflow   string                `json:"workflow" yaml:"workflow"`
	Secret     string                `json:"secret" yaml:"secret"`
	Ref        *string               `json:"ref,omitempty" yaml:"ref,omitempty"`
	WFInputs   []v2alpha2.NamedValue `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Inputs     `json:",inline" yaml:",inline"`
}

// GHWorkflowTask encapsulates the task.
type GHWorkflowTask struct {
	tasks.TaskMeta `json:",inline" yaml:",inline"`
	With           GHWorkflowInputs `json:"with" yaml:"with"`
}

// MakeGHWorkflowTask converts an spec to a task.
func MakeGHWorkflowTask(t *v2alpha2.TaskSpec) (tasks.Task, error) {
	if t.Task != LibraryName+"/"+GHWorkflowTaskName {
		return nil, fmt.Errorf("library and task need to be '%s' and '%s'", LibraryName, GHWorkflowTaskName)
	}
	var jsonBytes []byte
	var task tasks.Task
	// convert t to jsonBytes
	jsonBytes, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	// convert jsonString to task
	task = &GHWorkflowTask{}
	err = json.Unmarshal(jsonBytes, &task)
	return task, err
}

// ToHTTPTask converts a GHWorkflowTask task to a HTTPTask
func (t *GHWorkflowTask) ToHTTPTask() *HTTPTask {
	authType := v2alpha2.BearerAuthType
	authtype := &authType
	secret := &t.With.Secret

	ref := DefaultRef
	if t.With.Ref != nil {
		ref = *t.With.Ref
	}

	// compose body of POST request
	body := ""
	body += "{"
	body += "\"ref\": \"" + ref + "\","
	body += "\"inputs\": {"
	numWFInputs := len(t.With.WFInputs)
	for i := 0; i < numWFInputs; i++ {
		body += "\"" + t.With.WFInputs[i].Name + "\": \"" + t.With.WFInputs[i].Value + "\""
		if i+1 < numWFInputs {
			body += ","
		}
	}
	body += "}"
	body += "}"

	tSpec := &HTTPTask{
		TaskMeta: tasks.TaskMeta{
			Library: LibraryName,
			Task:    GHWorkflowTaskName,
		},
		With: HTTPInputs{
			URL:      "https://api.github.com/repos/" + t.With.Repository + "/actions/workflows/" + t.With.Workflow + "/dispatches",
			AuthType: authtype,
			Secret:   secret,
			Headers: []v2alpha2.NamedValue{{
				Name:  "Accept",
				Value: "application/vnd.github.v3+json",
			}},
			Body: &body,
		},
	}

	if t.With.IgnoreFailure != nil {
		tSpec.With.IgnoreFailure = t.With.IgnoreFailure
	}

	log.Info("Dispatching GitHub workflow: ", tSpec.With.URL)
	log.Info(*tSpec.With.Body)

	return tSpec
}

// Run the task. Ignores failures unless the task indicates ignoreFailures: false
func (t *GHWorkflowTask) Run(ctx context.Context) error {
	return t.ToHTTPTask().Run(ctx)
}
