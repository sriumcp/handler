package base

import (
	"testing"

	"github.com/iter8-tools/handler/utils"
	"github.com/stretchr/testify/assert"
)

func init() {
	log = utils.GetLogger()
}

func TestExtrapolate(t *testing.T) {
	tags := Tags{
		M: &map[string]string{"revision": "revision1", "container": "super-container"},
	}
	str := `hello {{ index . "revision" }} world`
	extrapolated, err := tags.Extrapolate(&str)
	assert.NoError(t, err)
	assert.Equal(t, "hello revision1 world", extrapolated)

	tags = Tags{}
	extrapolated, err = tags.Extrapolate(&str)
	assert.NoError(t, err)
	assert.Equal(t, str, extrapolated)

	tags = Tags{
		M: &map[string]string{"revision": "revision1", "container": "super-container"},
	}
	str = `hello {{{ romeo . "revision" alpha tango }} world`
	_, err = tags.Extrapolate(&str)
	assert.Error(t, err)

	str = `hello {{ index . 0 }} world`
	_, err = tags.Extrapolate(&str)
	assert.Error(t, err)
}
