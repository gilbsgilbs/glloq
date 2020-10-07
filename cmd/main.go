package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gilbsgilbs/glloq"
	"github.com/gilbsgilbs/glloq/anylocker"
)

func getGlloqOptionsFromEnv(envVars []string) (map[string]string, []string) {
	opts := map[string]string{}
	rest := []string{}

	for _, envVar := range envVars {
		if !strings.HasPrefix(envVar, "GLLOQ_") {
			rest = append(rest, envVar)
			continue
		}

		parts := strings.SplitN(envVar, "=", 2)
		key := parts[0]
		val := parts[1]

		key = strings.ToLower(strings.TrimPrefix(key, "GLLOQ_"))
		opts[key] = val
	}

	return opts, rest
}

func RunGlloq(env []string, args []string) (int, error) {
	opts, env := getGlloqOptionsFromEnv(env)

	dsn := opts["dsn"]
	if dsn == "" {
		return 1, glloq.ErrDSNNotSet
	}

	timeoutSeconds, _ := strconv.Atoi(opts["timeout"])

	lockerOptions := glloq.Options{
		DSN:     dsn,
		Key:     opts["key"],
		Timeout: time.Duration(timeoutSeconds) * time.Second,
		Params:  opts,
	}

	if err := glloq.UseLocker(
		&anylocker.Locker{},
		&lockerOptions,
		func() error {
			if len(args) == 0 {
				return nil
			}

			cmdName := args[0]
			cmdArgs := args[1:]
			cmd := exec.Command(cmdName, cmdArgs...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Env = env

			return cmd.Run()
		}); err != nil {
		if exitError, isExitError := err.(*exec.ExitError); isExitError {
			return exitError.ExitCode(), nil
		}

		return 1, err
	}

	return 0, nil
}

func main() {
	exitCode, err := RunGlloq(os.Environ(), os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(exitCode)
}
