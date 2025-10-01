package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	ErrInvalidOSARCH           = errors.New("invalid os/arch configuration")
	ErrUnsupportedTargetOSARCH = errors.New("unable to find go dist to support target os/arch combination(s)")
)

type OSARCH struct {
	OS   string
	ARCH string
}

func NewOSARCH() OSARCH {
	return OSARCH{"", ""}
}

type GoDist struct {
	GOOS         string `json:"GOOS"`
	GOARCH       string `json:"GOARCH"`
	CgoSupported bool   `json:"CgoSupported"`
	FirstClass   bool   `json:"FirstClass"`
}

type BuildConfig struct {
	ProjectDir string
	OutputDir  string
	BinaryName string
	Targets    []OSARCH
}

func (d GoDist) GOOSEnv() string {
	return fmt.Sprintf("GOOS=%s", d.GOOS)
}

func (d GoDist) GOARCHEnv() string {
	return fmt.Sprintf("GOARCH=%s", d.GOARCH)
}

func NewConfig() BuildConfig {
	return BuildConfig{
		ProjectDir: "./",
		OutputDir:  "./build",
		BinaryName: "build",
		Targets:    []OSARCH{},
	}
}

func getTargetBuilds(targets []OSARCH, allDists []GoDist) []GoDist {

	if len(targets) == 0 {
		return allDists
	}
	targetDists := []GoDist{}

	for _, target := range targets {
		for _, dist := range allDists {
			if target.ARCH == "" {
				if target.OS == dist.GOOS {
					targetDists = append(targetDists, dist)
				}
			} else {
				if target.OS == dist.GOOS && target.ARCH == dist.GOARCH {
					targetDists = append(targetDists, dist)
				}

			}
		}
	}

	return targetDists
}

func getBuildOptions(ctx context.Context, targets []OSARCH) ([]GoDist, error) {
	cmd := exec.CommandContext(ctx, "go", "tool", "dist", "list", "-json")
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	rawJson, err := cmd.Output()

	if err != nil {
		return []GoDist{}, fmt.Errorf("dist: %w", err)
	}

	var supportedDists []GoDist
	if err := json.Unmarshal(rawJson, &supportedDists); err != nil {
		return nil, fmt.Errorf("json parse: %w", err)
	}

	if len(targets) == 0 {
		return supportedDists, nil
	}

	targetDists := getTargetBuilds(targets, supportedDists)

	if len(targetDists) > 0 {
		return targetDists, nil
	} else {
		return []GoDist{}, ErrUnsupportedTargetOSARCH
	}
}

func Build(config BuildConfig, dist GoDist) error {

	outputDir := "./build"

	filename := fmt.Sprintf("%s-%s_%s", config.BinaryName, dist.GOOS, dist.GOARCH)

	if dist.GOOS == "windows" || dist.GOOS == "nt" {
		filename += ".exe"
	}

	fp := filepath.Join(outputDir, filename)

	cmd := exec.Command("go", "build", "-o", fp)
	cmd.Dir = config.ProjectDir
	cmd.Env = append(os.Environ(),
		dist.GOOSEnv(),
		dist.GOARCHEnv(),
	)

	return nil

}

func parseStringToOSARCH(rawStr string) (OSARCH, error) {

	if rawStr == "" {
		return OSARCH{}, ErrInvalidOSARCH
	}

	strLower := strings.ToLower(rawStr)
	splitStr := strings.Split(strLower, "/")

	if len(splitStr) == 1 {
		return OSARCH{
			OS:   splitStr[0],
			ARCH: "",
		}, nil
	} else if len(splitStr) == 2 {
		return OSARCH{
			OS:   splitStr[0],
			ARCH: splitStr[1],
		}, nil
	} else {
		return OSARCH{}, ErrInvalidOSARCH
	}

}

func getProjectName(projFp string) (string, error) {
	if projFp == "." {

		return os.Getwd()

	}

	return filepath.Base(projFp), nil
}

func main() {
	log.SetFlags(0)

	ctx := context.Background()

	var targetOS []OSARCH
	var targetOSRaw []string

	targetOSARCHFunc := func(v string) error {

		osarch, err := parseStringToOSARCH(v)

		if err == ErrInvalidOSARCH {
			fmt.Fprintf(os.Stderr, "Unable to parse %s to valid OS/ARCH\n", v)
			return nil
		} else if err != nil {
			return fmt.Errorf("parse osarch: %w", err)
		}

		targetOSRaw = append(targetOSRaw, v)

		targetOS = append(targetOS,
			osarch)
		return nil
	}

	flag.Func("target",
		"Specify what OS to target. Additional specifier can be supplied with <os>/<arch>.",
		targetOSARCHFunc)

	var outputDir string
	flag.StringVar(&outputDir, "o", "", "Specify the output directory to build in.")

	var binaryName string
	flag.StringVar(&binaryName, "n", "", "Specify the name of the binary build file(s)")

	projectDir := "."
	if len(flag.Args()) > 1 {
		projectDir = flag.Args()[1]
	}

	projectName, err := getProjectName(projectDir)

	if err != nil {
		log.Fatalln("project name:", err)
	}

	if outputDir == "" {
		outputDir = filepath.Join(projectDir, "build")
	}

	buildDists, err := getBuildOptions(ctx, targetOS)

	if err == ErrUnsupportedTargetOSARCH {
		log.Fatalln("Unsupported targets: ", strings.Join(targetOSRaw, "\n"), "\n", err)
	} else if err != nil {
		log.Fatalln("build options:", err)
	}

	config := NewConfig()
	config.Targets = targetOS
	config.BinaryName = projectName
	config.OutputDir = outputDir
	config.ProjectDir = projectDir

	for _, dist := range buildDists {
		Build(config, dist)

	}

}
