package client

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/containers/image/v5/docker/reference"
	yaml "gopkg.in/yaml.v3"
)

// Status summarizes the current worldview of the rpm-ostree daemon.
// The deployment list is the primary data.
type Status struct {
	// Deployments is the list of bootable filesystem trees.
	Deployments []Deployment
	// Transaction is the active transaction, if any.
	Transaction *[]string
}

// Deployment represents a bootable filesystem tree
type Deployment struct {
	ID                      string   `json:"id"`
	OSName                  string   `json:"osname"`
	Serial                  int32    `json:"serial"`
	Checksum                string   `json:"checksum"`
	Version                 string   `json:"version"`
	Timestamp               uint64   `json:"timestamp"`
	Booted                  bool     `json:"booted"`
	Staged                  bool     `json:"staged"`
	LiveReplaced            string   `json:"live-replaced,omitempty"`
	Origin                  string   `json:"origin"`
	CustomOrigin            []string `json:"custom-origin"`
	ContainerImageReference string   `json:"container-image-reference"`
}

// Client is a handle for interacting with an rpm-ostree based system.
type Client struct {
	clientid string
}

// NewClient creates a new rpm-ostree client.  The client identifier should be a short, unique and ideally machine-readable string.
// This could be as simple as `examplecorp-management-agent`.
// If you want to be more verbose, you could use a URL, e.g. `https://gitlab.com/examplecorp/management-agent`.
func NewClient(id string) Client {
	return Client{
		clientid: id,
	}
}

func (client *Client) newCmd(args ...string) *exec.Cmd {
	r := exec.Command("rpm-ostree", args...)
	r.Env = append(r.Env, "RPMOSTREE_CLIENT_ID", client.clientid)
	return r
}

func (client *Client) run(args ...string) error {
	c := client.newCmd(args...)
	return c.Run()
}

// RpmOstreeVersion returns the running rpm-ostree version number
func (client *Client) RpmOstreeVersion() (string, error) {
	var q struct {
		Version string `json:"version"`
	}
	c := client.newCmd("--version")
	buf, err := c.Output()
	if err != nil {
		return "", err
	}

	if err := yaml.Unmarshal(buf, &q); err != nil {
		return "", fmt.Errorf("failed to parse `rpm-ostree --version` output: %w", err)
	}

	return q.Version, nil
}

func compareVersionStrings(required, actual string) (bool, error) {
	verparts := strings.Split(actual, ".")
	verlen := len(verparts)
	requiredparts := strings.Split(required, ".")
	for i, req := range requiredparts {
		if i >= verlen {
			break
		}
		reqv, err := strconv.Atoi(req)
		if err != nil {
			return false, err
		}
		actualv, err := strconv.Atoi(verparts[i])
		if err != nil {
			return false, err
		}
		if actualv < reqv {
			return false, nil
		}
	}
	return true, nil
}

// RpmOstreeVersionEqualOrGreater checks whether the version of rpm-ostree is new enough.
func (client *Client) RpmOstreeVersionEqualOrGreater(required string) (bool, error) {
	actual, err := client.RpmOstreeVersion()
	if err != nil {
		return false, err
	}

	return compareVersionStrings(required, actual)
}

// QueryStatus loads the current system state.
func (client *Client) QueryStatus() (*Status, error) {
	var q Status
	c := client.newCmd("status", "--json")
	buf, err := c.Output()
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(buf, &q); err != nil {
		return nil, fmt.Errorf("failed to parse `rpm-ostree status --json` output: %w", err)
	}

	return &q, nil
}

// GetBootedDeployment finds the booted deployment, or returns an error if none is found.
func (s *Status) GetBootedDeployment() (*Deployment, error) {
	for num := range s.Deployments {
		deployment := s.Deployments[num]
		if deployment.Booted {
			return &deployment, nil
		}
	}
	return nil, fmt.Errorf("no booted deployment found")
}

// GetStagedDeployment finds the staged deployment, or returns nil if none is found.
func (s *Status) GetStagedDeployment() *Deployment {
	for num := range s.Deployments {
		deployment := s.Deployments[num]
		if deployment.Staged {
			return &deployment
		}
	}
	return nil
}

// Remove the pending deployment.
func (client *Client) RemovePendingDeployment() error {
	return client.run("cleanup", "-p")
}

// Remove the rollback deployment.
func (client *Client) RemoveRollbackDeployment() error {
	return client.run("cleanup", "-r")
}

// ChangeKernelArguments adjusts the kernel arguments.
func (client *Client) ChangeKernelArguments(toAdd []string, toRemove []string) error {
	args := []string{"kargs"}
	for _, arg := range toRemove {
		args = append(args, "--delete="+arg)
	}
	for _, arg := range toAdd {
		args = append(args, "--append="+arg)
	}
	return client.run(args...)
}

// ChangePackages installs or removes packages.
func (client *Client) ChangePackages(toAdd []string, toRemove []string) error {
	args := []string{}
	if len(toAdd) == 0 {
		args = append(args, "uninstall")
		args = append(args, toRemove...)
		for _, pkg := range toAdd {
			args = append(args, "--install")
			args = append(args, pkg)
		}
	} else {
		args = append(args, "install")
		args = append(args, toAdd...)
		for _, pkg := range toRemove {
			args = append(args, "--uninstall")
			args = append(args, pkg)
		}
	}
	return client.run(args...)
}

// OverrideRemove uninstalls base packages, optionally installing new ones at the same time.
func (client *Client) OverrideRemove(toRemove []string, toInstall []string) error {
	args := []string{"override", "remove"}
	args = append(args, toRemove...)
	for _, pkg := range toInstall {
		args = append(args, "--install")
		args = append(args, pkg)
	}
	return client.run(args...)
}

// OverrideRemove drops overrides, optionally uninstalling new ones at the same time.
func (client *Client) OverrideReset(toReset []string, toUninstall []string) error {
	args := []string{"override", "reset"}
	args = append(args, toReset...)
	for _, pkg := range toUninstall {
		args = append(args, "--uninstall")
		args = append(args, pkg)
	}
	return client.run(args...)
}

// RebaseToContainerImage switches to the target container image
func (client *Client) RebaseToContainerImage(target reference.Reference) error {
	return client.run("rebase", "--experimental", fmt.Sprintf("ostree-image-signed:docker://%s", target.String()))
}

// RebaseToContainerImageAllowUnsigned switches to the target container image, ignoring lack of image signatures.
func (client *Client) RebaseToContainerImageAllowUnsigned(target reference.Reference) error {
	return client.run("rebase", "--experimental", fmt.Sprintf("ostree-unverified-registry:%s", target.String()))
}
