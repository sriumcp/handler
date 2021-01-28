package base

import (
	"bytes"
	"context"
	"errors"
	"html/template"

	"github.com/iter8-tools/handler/utils"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = utils.GetLogger()
}

// TaskMeta identifies Library and Name of a task.
type TaskMeta struct {
	// Library where this task is defined.
	Library string `json:"library" yaml:"library"`
	// Name (type) of this task.
	Task string `json:"task" yaml:"task"`
}

// Task defines common method signatures for every task.
type Task interface {
	Run(ctx context.Context) error
	DryRun()
	LocalRun(ctx context.Context) error
	LocallyRunnable() bool
	Extrapolate(tags *Tags) error
}

// Tags supports string extrapolation using tags.
type Tags struct {
	M *map[string]string
}

// Extrapolate str using tags.
func (tags *Tags) Extrapolate(str *string) (string, error) {
	if tags == nil || tags.M == nil { // return a copy of the string
		return *str, nil
	}
	var err error
	var templ *template.Template
	if templ, err = template.New("").Parse(*str); err == nil {
		buf := bytes.Buffer{}
		if err = templ.Execute(&buf, tags.M); err == nil {
			return string(buf.Bytes()), nil
		}
		log.Error("template execution error: ", err)
		return "", errors.New("cannot extrapolate string")
	}
	log.Error("template creation error: ", err)
	return "", errors.New("cannot extrapolate string")
}

// ContextKey is the type of key that will be used to index into context.
type ContextKey string
