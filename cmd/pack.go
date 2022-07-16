package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/gosuri/eve/logger"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var DefaultBuilder = "heroku/buildpacks:20"

type PackFlags struct {
	Env      []string
	EnvFiles []string
	Builder  string
	Image    string
}

func NewPackCMD(ctx context.Context, cancel context.CancelFunc) *cobra.Command {
	packFlags := &PackFlags{}
	cmd := &cobra.Command{
		Use:   "pack",
		Short: "Pack your project into a container using buildpacks",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runPack(ctx, cancel, packFlags); err != nil {
				logger.Errorf("error: %v", err)
				return
			}
		},
	}
	cmd.Flags().StringArrayVarP(&packFlags.Env, "env", "e", []string{}, "Build-time environment variable, in the form 'VAR=VALUE' or 'VAR'.\nWhen using latter value-less form, value will be taken from current\n  environment at the time this command is executed.\nThis flag may be specified multiple times and will override\n  individual values defined by --env-file."+stringArrayHelp("env")+"\nNOTE: These are NOT available at image runtime.")
	cmd.Flags().StringArrayVar(&packFlags.EnvFiles, "env-file", []string{}, "Build-time environment variables file\nOne variable per line, of the form 'VAR=VALUE' or 'VAR'\nWhen using latter value-less form, value will be taken from current\n  environment at the time this command is executed\nNOTE: These are NOT available at image runtime.\"")
	cmd.Flags().StringVar(&packFlags.Builder, "builder", "", "Builder to use for building the image")
	cmd.Flags().StringVar(&packFlags.Image, "image", "", "Name of the image to build")
	return cmd
}

func runPack(ctx context.Context, cancel context.CancelFunc, packFlags *PackFlags) (err error) {
	// Check if the image name is provided if not read it from IMAGE file state directory
	if packFlags.Image == "" {
		packFlags.Image, err = readvar("IMAGE")
		if err != nil {
			return errors.Wrap(err, "failed to read IMAGE variable")
		}
	}
	// Check if env-file is specified and if so, parse it
	if len(packFlags.EnvFiles) == 0 && varExists("ENV") { // if ENV is set, use it
		envFile := path.Join(globalFlags.Path, globalFlags.StateDir, "ENV") // path to the ENV file
		packFlags.EnvFiles = []string{envFile}
	}

	// check if the builder is provided if not read it from BUILDER file state directory
	if packFlags.Builder == "" {
		if varExists("BUILDER") {
			packFlags.Builder, _ = readvar("BUILDER")
		} else {
			// if BUILDER is not set, use the default one
			packFlags.Builder = DefaultBuilder

			// write the default builder to the state directory
			writevar("BUILDER", packFlags.Builder)
		}
	}

	c := []string{"build", packFlags.Image, "--builder", packFlags.Builder}
	env, err := parseEnv(packFlags.EnvFiles, packFlags.Env)
	if err != nil {
		return errors.Wrap(err, "error parsing environment variables")
	}

	for k, v := range env {
		c = append(c, "--env", k+"="+v)
	}

	logger.Debugf("runPack running: %s", strings.Join(c, " "))

	cmd := exec.CommandContext(ctx, "pack", c...)

	r, _ := cmd.StdoutPipe()
	err = cmd.Start()

	if err != nil {
		return errors.Wrap(err, "error starting pack")
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	if err := cmd.Wait(); err != nil {
		logger.Errorf("error: %v", err)
		return errors.Wrap(err, "error waiting for pack")
	}
	return nil
}

func parseEnv(envFiles []string, envVars []string) (map[string]string, error) {
	env := map[string]string{}
	for _, envFile := range envFiles {
		envFileVars, err := parseEnvFile(envFile)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse env file '%s'", envFile)
		}

		for k, v := range envFileVars {
			env[k] = v
		}
	}
	for _, envVar := range envVars {
		env = addEnvVar(env, envVar)
	}
	return env, nil
}

func parseEnvFile(filename string) (map[string]string, error) {
	out := make(map[string]string)
	f, err := ioutil.ReadFile(filepath.Clean(filename))
	if err != nil {
		return nil, errors.Wrapf(err, "open %s", filename)
	}
	for _, line := range strings.Split(string(f), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out = addEnvVar(out, line)
	}
	return out, nil
}

func addEnvVar(env map[string]string, item string) map[string]string {
	arr := strings.SplitN(item, "=", 2)
	if len(arr) > 1 {
		env[arr[0]] = arr[1]
	} else {
		env[arr[0]] = os.Getenv(arr[0])
	}
	return env
}
