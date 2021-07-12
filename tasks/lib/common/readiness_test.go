package common

import (
	"encoding/json"
	"testing"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/handler/tasks"
	"github.com/stretchr/testify/assert"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestInvalidObjName(t *testing.T) {
	initDelay, _ := json.Marshal(5)
	numRetries, _ := json.Marshal(3)
	intervalSeconds, _ := json.Marshal(5)
	objRefs, _ := json.Marshal([]ObjRef{
		{
			Kind:      "deploy",
			Namespace: tasks.StringPointer("default"),
			Name:      "hello world",
			WaitFor:   tasks.StringPointer("condition=available"),
		},
	})
	_, err := MakeTask(&v2alpha2.TaskSpec{
		Task: LibraryName + "/" + ReadinessTaskName,
		With: map[string]apiextensionsv1.JSON{
			"initialDelaySeconds": {Raw: initDelay},
			"numRetries":          {Raw: numRetries},
			"intervalSeconds":     {Raw: intervalSeconds},
			"objRefs":             {Raw: objRefs},
		},
	})
	assert.Error(t, err)
}
