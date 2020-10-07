package main_test

import (
	"testing"

	cmd "github.com/gilbsgilbs/glloq/cmd"
	"github.com/stretchr/testify/assert"
)

func TestCli(t *testing.T) {
	env := []string{
		"GLLOQ_POLL_DELAY=100",
		"GLLOQ_TIMEOUT=5",
	}

	t.Run("runs", func(t *testing.T) {
		exitCode, err := cmd.RunGlloq(env, []string{})
		assert.Equal(t, 0, exitCode)
		assert.Nil(t, err)
	})

	t.Run("runs concurrent run", func(t *testing.T) {
		key := t.Name()
		go func() {
			if _, err := cmd.RunGlloq(
				append(env, "GLLOQ_KEY="+key),
				[]string{"sleep", "10"},
			); err != nil {
				panic(err)
			}
		}()

		exitCode, err := cmd.RunGlloq(
			append(env, "GLLOQ_KEY="+key),
			[]string{"sleep", "1"},
		)

		assert.Equal(t, 0, exitCode)
		assert.Nil(t, err)
	})

	t.Run("forwards errors", func(t *testing.T) {
		cmdArgs := []string{"sh", "-c", "exit 22"}

		exitCode, err := cmd.RunGlloq(env, cmdArgs)

		assert.Equal(t, 22, exitCode)
		assert.Nil(t, err)
	})
}
