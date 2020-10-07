package main_test

import (
	"runtime"
	"testing"

	"github.com/gilbsgilbs/glloq"
	cmd "github.com/gilbsgilbs/glloq/cmd"
	"github.com/stretchr/testify/assert"
)

func TestCli(t *testing.T) {
	env := []string{
		"GLLOQ_DSN=file://.lock",
		"GLLOQ_POLL_DELAY=1",
		"GLLOQ_TIMEOUT=5",
	}

	var sleep []string
	if runtime.GOOS == "windows" {
		sleep = []string{"timeout"}
	} else {
		sleep = []string{"sleep"}
	}

	t.Run("test DSN not set", func(t *testing.T) {
		exitCode, err := cmd.RunGlloq([]string{}, []string{})
		assert.Equal(t, 1, exitCode)
		assert.Equal(t, glloq.ErrDSNNotSet, err)
	})

	t.Run("runs", func(t *testing.T) {
		exitCode, err := cmd.RunGlloq(env, []string{})
		assert.Equal(t, 0, exitCode)
		assert.Nil(t, err)
	})

	t.Run("runs concurrent run", func(t *testing.T) {
		go func() {
			if _, err := cmd.RunGlloq(
				env,
				append(sleep, "5"),
			); err != nil {
				panic(err)
			}
		}()

		exitCode, err := cmd.RunGlloq(env, append(sleep, "1"))

		assert.Equal(t, 0, exitCode)
		assert.Nil(t, err)
	})

	t.Run("forwards errors", func(t *testing.T) {
		var cmdArgs []string
		if runtime.GOOS == "windows" {
			cmdArgs = []string{"exit", "22"}
		} else {
			cmdArgs = []string{"sh", "-c", "exit 22"}
		}

		exitCode, err := cmd.RunGlloq(env, cmdArgs)

		assert.Equal(t, 22, exitCode)
		assert.Nil(t, err)
	})
}
