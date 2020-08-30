package repo

import (
    "os/exec"
    "regexp"
)

import "github.com/caring/go-packages/pkg/errors"

// isTerraformInstalled checks if Terraform is installed by searching for it
// in the directories named by the PATH environment variable. If its not found
// an error is returned, otherwise nil is returned.
func isTerraformInstalled() error {
    _, err := exec.LookPath("terraform")

    if err != nil {
        return errors.Wrap(err, "Terraform not found in PATH")
    }
    return nil
}
// TODO: Add timeout and better OS signal handling
// getTerraformVersion executes the command `terraform version` and
// returns the version parsed from the output
func getTerraformVersion() ([]byte, error) {
    tf := exec.Command("terraform", "version")
    out, err := tf.Output()

    if err != nil {
        return nil, errors.Wrap(err, "Error encountered while get Terraform version!")
    }

    pattern := regexp.MustCompile(`v[0-9]\.[0-9]+\.[0-9]+`)
    return pattern.Find(out), nil
}

// initTerraform initializes the Terraform directory of the newly
// created project repository. The var tfDir is the absolute path
// of the Terraform directory.
func initTerraform(tfDir string) error {
    tf := exec.Command("terraform", "init", tfDir)
    err := tf.Run()

    if err != nil {
        return errors.Wrap(err, "Error encountered while initializing Terraform directory!")
    }
    return nil
}
