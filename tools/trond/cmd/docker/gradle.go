package docker

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
)

var (
	dockerNamePattern = regexp.MustCompile(`^[a-z0-9]+(?:[._-][a-z0-9]+)*$`)
	versionTagPattern = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$`)
)

func validateDockerGradleFlags(org, artifact, version, network, platform string) error {
	if !dockerNamePattern.MatchString(org) {
		return fmt.Errorf("invalid org %q: only lowercase letters, digits, '.', '_' and '-' are allowed", org)
	}
	if !dockerNamePattern.MatchString(artifact) {
		return fmt.Errorf("invalid artifact %q: only lowercase letters, digits, '.', '_' and '-' are allowed", artifact)
	}
	if !versionTagPattern.MatchString(version) {
		return fmt.Errorf("invalid version %q: must match Docker tag format", version)
	}
	if network != "mainnet" && network != "nile" && network != "" {
		return fmt.Errorf("invalid network %q: allowed values are mainnet or nile", network)
	}
	if platform != "linux/amd64" && platform != "linux/arm64" && platform != "" {
		return fmt.Errorf("invalid platform %q: allowed values are linux/amd64 or linux/arm64", platform)
	}
	return nil
}

func runGradleDockerTask(task, org, artifact, version, network, platform string) error {
	if err := validateDockerGradleFlags(org, artifact, version, network, platform); err != nil {
		return err
	}

	args := []string{
		"--no-daemon",
		task,
		fmt.Sprintf("-PdockerOrgName=%s", org),
		fmt.Sprintf("-PdockerArtifactName=%s", artifact),
		fmt.Sprintf("-Prelease.releaseVersion=%s", version),
	}

	if network != "" {
		args = append(args, fmt.Sprintf("-Pnetwork=%s", network))
	}
	if platform != "" {
		args = append(args, fmt.Sprintf("-Pplatform=%s", platform))
	}

	cmd := exec.Command("./gradlew", args...)
	cmd.Dir = "./tools/gradlew"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gradle task %s failed: %w", task, err)
	}
	return nil
}
