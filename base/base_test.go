package base

import (
	"testing"

	"github.com/iter8-tools/handler/utils"
	"github.com/stretchr/testify/assert"
)

func init() {
	log = utils.GetLogger()
}

func TestInterpolate(t *testing.T) {
	tags := Tags{
		M: map[string]string{"revision": "revision1", "container": "super-container"},
	}
	str := `hello {{ index . "revision" }} world`
	interpolated, err := tags.Interpolate(&str)
	assert.NoError(t, err)
	assert.Equal(t, "hello revision1 world", interpolated)

	tags = Tags{}
	interpolated, err = tags.Interpolate(&str)
	assert.NoError(t, err)
	assert.Equal(t, str, interpolated)

	tags = Tags{
		M: map[string]string{"revision": "revision1", "container": "super-container"},
	}
	str = `hello {{{ romeo . "revision" alpha tango }} world`
	_, err = tags.Interpolate(&str)
	assert.Error(t, err)

	str = `hello {{ index . 0 }} world`
	_, err = tags.Interpolate(&str)
	assert.Error(t, err)
}
