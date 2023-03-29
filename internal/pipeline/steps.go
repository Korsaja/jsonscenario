package pipeline

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"

	"github.com/Korsaja/jsonscenario/pkg/models"
)

var ErrStepsAlready = errors.New("error steps already processed")

type Steps []*Step

func (steps Steps) Validate() error {
	for _, step := range steps {
		if step.Result != "" {
			return ErrStepsAlready
		}
	}
	return nil
}

type Step struct {
	Name   models.ActionFile `json:"name"`
	Result string            `json:"result,omitempty"`
	Args   []string          `json:"args,omitempty"`
}

func (s *Step) String() string {
	var buf bytes.Buffer
	_ = json.NewEncoder(&buf).Encode(s)
	return buf.String()
}
