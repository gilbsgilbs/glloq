package main_test

import (
	"testing"
	"time"

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
		done := false
		go func() {
			if _, err := cmd.RunGlloq(env, []string{"sleep", ".3"}); err != nil {
				panic(err)
			}
			done = true
		}()

		// Wait a bit so that the routine can start
		time.Sleep(50 * time.Millisecond)

		// The routine should be running, not done yet.
		assert.False(t, done)

		exitCode, err := cmd.RunGlloq(env, []string{"echo", "ok"})

		// The command should have waited for the routine to finish.
		assert.True(t, done)

		// Check the command exit code.
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
