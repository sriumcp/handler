package common

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/handler/tasks"
	"github.com/stretchr/testify/assert"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestMakeFakeCommonTask(t *testing.T) {
	_, err := MakeTask(&v2alpha2.TaskSpec{
		Task: LibraryName + "/" + "fake",
	})
	assert.Error(t, err)
}

func TestMakeTask(t *testing.T) {
	b, _ := json.Marshal("echo")
	a, _ := json.Marshal([]string{"hello", "people", "of", "earth"})
	task, err := MakeTask(&v2alpha2.TaskSpec{
		Task: "common/exec",
		With: map[string]apiextensionsv1.JSON{
			"cmd":  {Raw: b},
			"args": {Raw: a},
		},
	})
	assert.NotEmpty(t, task)
	assert.NoError(t, err)
	assert.Equal(t, "earth", task.(*ExecTask).With.Args[3])
	log.Trace(task.(*ExecTask).With.Args)

	exp, _ := (&tasks.Builder{}).FromFile(tasks.CompletePath("../../", "testdata/experiment10.yaml")).Build()
	task.Run(context.WithValue(context.Background(), tasks.ContextKey("experiment"), exp))

	task, err = MakeTask(&v2alpha2.TaskSpec{
		Task: "common/run",
		With: map[string]apiextensionsv1.JSON{
			"cmd": {Raw: b},
		},
	})
	assert.Nil(t, task)
	assert.Error(t, err)
}

func TestMakeFakeBashTask(t *testing.T) {
	_, err := MakeBashTask(&v2alpha2.TaskSpec{
		Task: LibraryName + "/" + "fake",
	})
	assert.Error(t, err)
}

func TestExecTaskNoInterpolation(t *testing.T) {
	b, _ := json.Marshal("echo")
	a, _ := json.Marshal([]string{"hello", "{{ omg }}", "world"})
	c, _ := json.Marshal(true)
	task, err := MakeTask(&v2alpha2.TaskSpec{
		Task: "common/exec",
		With: map[string]apiextensionsv1.JSON{
			"cmd":                  {Raw: b},
			"args":                 {Raw: a},
			"disableInterpolation": {Raw: c},
		},
	})

	assert.NotEmpty(t, task)
	assert.NoError(t, err)
	assert.Equal(t, "world", task.(*ExecTask).With.Args[2])
	log.Trace(task.(*ExecTask).With.Args)

	exp, _ := (&tasks.Builder{}).FromFile(tasks.CompletePath("../../", "testdata/experiment10.yaml")).Build()
	task.Run(context.WithValue(context.Background(), tasks.ContextKey("experiment"), exp))
}

func TestMakeBashTask(t *testing.T) {
	script, _ := json.Marshal("echo hello")
	task, err := MakeTask(&v2alpha2.TaskSpec{
		Task: LibraryName + "/" + BashTaskName,
		With: map[string]apiextensionsv1.JSON{
			"script": {Raw: script},
		},
	})
	assert.NotEmpty(t, task)
	assert.NoError(t, err)
	assert.Equal(t, "echo hello", task.(*BashTask).With.Script)
}

func TestBashRun(t *testing.T) {
	exp, err := (&tasks.Builder{}).FromFile(filepath.Join("..", "..", "..", "testdata", "common", "bashexperiment.yaml")).Build()
	assert.NoError(t, err)
	actionSpec, err := exp.GetActionSpec("start")
	assert.NoError(t, err)
	// action, err := GetAction(exp, actionSpec)
	action, err := MakeTask(&actionSpec[0])
	assert.NoError(t, err)
	ctx := context.WithValue(context.Background(), tasks.ContextKey("experiment"), exp)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	err = action.Run(ctx)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(buf.String(), "\necho \"v1\"\n"))
}

func TestMakePromoteKubectlTask(t *testing.T) {
	namespace, _ := json.Marshal("default")
	recursive, _ := json.Marshal(true)
	manifest, _ := json.Marshal("promote.yaml")
	task, err := MakeTask(&v2alpha2.TaskSpec{
		Task: LibraryName + "/" + PromoteKubectlTaskName,
		With: map[string]apiextensionsv1.JSON{
			"namespace": {Raw: namespace},
			"recursive": {Raw: recursive},
			"manifest":  {Raw: manifest},
		},
	})
	assert.NotEmpty(t, task)
	assert.NoError(t, err)
	assert.Equal(t, "default", *task.(*PromoteKubectlTask).With.Namespace)
	assert.Equal(t, "promote.yaml", task.(*PromoteKubectlTask).With.Manifest)
	assert.Equal(t, true, *task.(*PromoteKubectlTask).With.Recursive)

	bTask := *task.(*PromoteKubectlTask).ToBashTask()
	assert.Equal(t, "kubectl apply --namespace default --recursive --filename promote.yaml", bTask.With.Script)
}

func TestMakeReadinessTask(t *testing.T) {
	initDelay, _ := json.Marshal(5)
	numRetries, _ := json.Marshal(3)
	intervalSeconds, _ := json.Marshal(5)
	objRefs, _ := json.Marshal([]ObjRef{
		{
			Kind:      "deploy",
			Namespace: tasks.StringPointer("default"),
			Name:      "hello",
			WaitFor:   tasks.StringPointer("condition=available"),
		},
		{
			Kind:      "deploy",
			Namespace: tasks.StringPointer("default"),
			Name:      "hello-candidate",
			WaitFor:   tasks.StringPointer("condition=available"),
		},
	})
	task, err := MakeTask(&v2alpha2.TaskSpec{
		Task: LibraryName + "/" + ReadinessTaskName,
		With: map[string]apiextensionsv1.JSON{
			"initialDelaySeconds": {Raw: initDelay},
			"numRetries":          {Raw: numRetries},
			"intervalSeconds":     {Raw: intervalSeconds},
			"objRefs":             {Raw: objRefs},
		},
	})
	assert.NotEmpty(t, task)
	assert.NoError(t, err)
	assert.Equal(t, int32(5), *task.(*ReadinessTask).With.InitialDelaySeconds)
	assert.Equal(t, int32(3), *task.(*ReadinessTask).With.NumRetries)
	assert.Equal(t, int32(5), *task.(*ReadinessTask).With.IntervalSeconds)
	assert.Equal(t, 2, len(task.(*ReadinessTask).With.ObjRefs))
}
