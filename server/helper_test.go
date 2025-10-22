package server

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/olive-io/gflow/server/testdata"
)

func Test_parseDoc(t *testing.T) {
	data, err := testdata.Open("swagger.json")
	if !assert.NoError(t, err) {
		return
	}

	runner, endpoints, err := parseDoc(string(data), "json")
	if !assert.NoError(t, err) {
		return
	}

	assert.NotNil(t, runner)
	assert.NotNil(t, endpoints)

	t.Logf("%+v\n", len(endpoints))

	data, err = testdata.Open("swagger.yaml")
	if !assert.NoError(t, err) {
		return
	}

	runner, endpoints, err = parseDoc(string(data), "yaml")
	if !assert.NoError(t, err) {
		return
	}

	assert.NotNil(t, runner)
	assert.NotNil(t, endpoints)

	t.Logf("%+v\n", len(endpoints))
}
