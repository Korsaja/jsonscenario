package pipeline

import (
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

// easy tests
func TestNewPipeline(t *testing.T) {
	in := "../testdata/steps_true_condition.json"
	out := "../testdata/result_true.json"
	pipe, err := NewPipeline(in, out, log.Default())
	require.NoError(t, err)
	require.NoError(t, pipe.Do())
}

func TestNewPipelineWithFalseCondition(t *testing.T) {
	in := "../testdata/steps_false_condition.json"
	out := "../testdata/result_false.json"
	pipe, err := NewPipeline(in, out, log.Default())
	require.NoError(t, err)
	require.NoError(t, pipe.Do())
}
