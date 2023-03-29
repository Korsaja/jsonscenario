package pipeline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/Korsaja/jsonscenario/internal/ffuncs"
	"github.com/Korsaja/jsonscenario/pkg/models"

	"github.com/pkg/errors"
)

type actionFunc func(args []string) (string, error)

var actionByStep = map[models.ActionFile]actionFunc{
	models.Create:    ffuncs.CreateFile,        // return filename
	models.Remove:    ffuncs.RemoveFile,        // return removed filename
	models.Rename:    ffuncs.RenameFile,        // return new path to filename
	models.CTime:     ffuncs.CTimeFile,         // return string date
	models.Write:     ffuncs.WriteString,       // write strings in args
	models.Condition: ffuncs.ValidateCondition, // if condition true run next step else set result to failed in next steps
}

type Pipeline struct {
	configPath, out string
	Steps           Steps       `json:"steps"`
	StepsResult     StepsResult `json:"steps_result"`

	L *log.Logger `json:"-"`
}

type StepsResult struct {
	Total      int    `json:"total"`
	Success    int    `json:"success"`
	Failed     int    `json:"failed"`
	FailedStep int    `json:"failed_step,omitempty"`
	Error      string `json:"error,omitempty"`
}

func NewPipeline(path, out string, logger *log.Logger) (*Pipeline, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	logger.Println("pipeline build...")
	pipe := &Pipeline{configPath: path, out: out, L: logger}
	if err = json.NewDecoder(f).Decode(pipe); err != nil {
		return nil, fmt.Errorf("failed parse config %s: %w", path, err)
	}

	if err = f.Close(); err != nil {
		return nil, err
	}

	return pipe, pipe.Steps.Validate()
}

func (pipe *Pipeline) Do() error {
	pipe.StepsResult.Total = len(pipe.Steps)
	for i, step := range pipe.Steps {
		action := actionByStep[step.Name]
		if action == nil {
			return errors.Errorf("invalid action step %s", step.Name.String())
		}

		pipe.L.Printf("run step %d action %s", i+1, step.Name.String())

		args := make([]string, 0)
		resultAction, err := action(step.Args)
		if errors.Is(err, ffuncs.ErrFalseCondition) {
			step.Result = ffuncs.Success.String()

			pipe.StepsResult.Success++
			pipe.StepsResult.FailedStep = i + 1
			pipe.StepsResult.Error = errors.Cause(err).Error()

			err = pipe.closeStepsWithError(i+1, nil)
			if flushErr := pipe.flush(); flushErr != nil {
				return flushErr
			}
			return err
		}
		if err != nil {
			pipe.StepsResult.FailedStep = i + 1
			pipe.StepsResult.Error = errors.Cause(err).Error()

			err = pipe.closeStepsWithError(i, err)
			if flushErr := pipe.flush(); flushErr != nil {
				return flushErr
			}
			return err
		}

		if resultAction != "" {
			args = append(args, resultAction)
		}
		// set result status
		step.Result = ffuncs.Success.String()
		pipe.StepsResult.Success++

		currAction := pipe.Steps[i].Name
		nextAction := pipe.nextAction(i + 1)
		// check correct call ctime func in pipe
		if currAction == models.CTime && nextAction != models.Condition {
			// set params to next step skip result action
			args = make([]string, len(step.Args))
			copy(args, step.Args)
		}

		if currAction == models.CTime && nextAction == models.Condition {
			args = append(args, step.Args...)
		}

		if currAction != models.CTime && nextAction == models.Condition {
			timeOperation, err := ffuncs.CTimeFile(args)
			if err != nil {
				pipe.StepsResult.FailedStep = i + 1
				pipe.StepsResult.Error = errors.Cause(err).Error()

				err = pipe.closeStepsWithError(i, err)
				if flushErr := pipe.flush(); flushErr != nil {
					return flushErr
				}
				return err
			}
			args = append([]string{timeOperation}, args...)
		}
		pipe.addArgs(i+1, args)
	}
	return pipe.flush()
}

func (pipe *Pipeline) flush() error {
	pipe.L.Printf("pipeline flush from memory to %s", pipe.out)
	const (
		flag = os.O_RDWR | os.O_CREATE | os.O_TRUNC
		perm = 0644
	)
	b, err := json.MarshalIndent(pipe, "", "\t")
	if err != nil {
		return err
	}

	f, err := os.OpenFile(pipe.out, flag, perm)
	if err != nil {
		return err
	}

	written, err := io.Copy(f, bytes.NewReader(b))
	if err != nil {
		return err
	}

	if err = f.Sync(); err != nil {
		return err
	}

	pipe.L.Printf("save to %s bytes %d", pipe.out, written)

	return f.Close()
}

func (pipe *Pipeline) addArgs(idx int, args []string) {
	if idx > len(pipe.Steps)-1 {
		return
	}
	pipe.L.Printf("set args %+v to next step num %d action %s", args, idx, pipe.Steps[idx].Name)
	// if condition args add to end else to begin
	if pipe.Steps[idx].Name == models.Condition {
		pipe.Steps[idx].Args = append(pipe.Steps[idx].Args, args...)
	} else {
		pipe.Steps[idx].Args = append(args, pipe.Steps[idx].Args...)
	}
}

func (pipe *Pipeline) nextAction(nextIndex int) models.ActionFile {
	if nextIndex > len(pipe.Steps)-1 {
		return models.Unknown
	}
	return pipe.Steps[nextIndex].Name
}

func (pipe *Pipeline) closeStepsWithError(from int, err error) error {
	// set failed steps
	for i := from; i < len(pipe.Steps); i++ {
		pipe.Steps[i].Result = ffuncs.Failed.String()
		pipe.StepsResult.Failed++
	}
	if err != nil {
		return errors.Wrapf(err, "step %d failed", from)
	}
	return nil
}
