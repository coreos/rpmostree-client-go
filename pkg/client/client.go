package client

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/containers/image/v5/docker/reference"
)

// Status houses zero or more RpmOstreeDeployments
// Subset of `rpm-ostree status --json`
// https://github.com/projectatomic/rpm-ostree/blob/bce966a9812df141d38e3290f845171ec745aa4e/src/daemon/rpmostreed-deployment-utils.c#L227
type Status struct {
	Deployments []Deployment
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

type Client struct {
	clientid string
}

// NewClient creates a new rpm-ostree client.
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