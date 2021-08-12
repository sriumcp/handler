package tasks

import (
	"context"

	"github.com/antonmedv/expr"
	"github.com/iter8-tools/etc3/api/v2alpha2"
)

func init() {
	log = GetLogger()
}

// Task defines common method signatures for every task.
type Task interface {
	Run(ctx context.Context) error
	GetCondition() *string
}

// Action is a slice of Tasks.
type Action []Task

// TaskMeta is common to all Tasks
type TaskMeta struct {
	Library   string  `json:"library" yaml:"library"`
	Task      string  `json:"task" yaml:"task"`
	Condition *string `json:"condition" yaml:"condition"`
}

// GetCondition returns condition from TaskMeta
func (tm TaskMeta) GetCondition() *string {
	return tm.Condition
}

// VersionInfo contains header and url information needed to send requests to each version.
type VersionInfo struct {
	Variables []v2alpha2.NamedValue `json:"variables,omitempty" yaml:"variables,omitempty"`
}

// Run the given action.
func (a *Action) Run(ctx context.Context) error {
	for i := 0; i < len(*a); i++ {
		log.Info("------ task starting")
		shouldRun := true
		exp, err := GetExperimentFromContext(ctx)
		if err != nil {
			return err
		}
		// if task has a condition
		if cond := (*a)[i].GetCondition(); cond != nil {
			// condition evaluates to false ... then shouldRun is false
			program, err := expr.Compile(*cond, expr.Env(exp), expr.AsBool())
			if err != nil {
				return err
			}

			output, err := expr.Run(program, exp)
			if err != nil {
				return err
			}

			shouldRun = output.(bool)
		}
		if shouldRun {
			err := (*a)[i].Run(ctx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GetDefaultTags creates interpolation.Tags from experiment referenced by context
func GetDefaultTags(ctx context.Context) *Tags {
	tags := NewTags()
	exp, err := GetExperimentFromContext(ctx)
	if err == nil {
		obj, err := exp.ToMap()
		if err == nil {
			tags = tags.
				With("this", obj).
				WithRecommendedVersionForPromotionDeprecated(&exp.Experiment)
		}
	} else {
		log.Warn("No experiment found in context")
	}

	return &tags
}
