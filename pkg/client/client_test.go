package client

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

//go:embed test-fixtures/workstation-status.json
var workstationFixture string

//go:embed test-fixtures/fcos-container-status.json
var fcosFixture string

func TestNewClient(t *testing.T) {
	c := NewClient("test")
	if c.clientid != "test" {
		panic("mismatched client id")
	}
}

func TestCompareVersionStrings(t *testing.T) {
	for _, req := range []string{"2023.0", "2022.8", "3000.5", "3000.5.5"} {
		v, err := compareVersionStrings(req, "2022.7")
		assert.Nil(t, err)
		assert.False(t, v)
	}
	for _, req := range []string{"2021.0", "2022", "2022.5", "10.1", "0"} {
		v, err := compareVersionStrings(req, "2022.7")
		assert.Nil(t, err)
		assert.True(t, v)
	}
}

func TestParseVersion(t *testing.T) {
	verdata := `rpm-ostree:
  Version: '2022.10'
  Git: 6b302116c969397fd71899e3b9bb3b8c100d1af9
  Features:
   - rust
   - compose
   - rhsm
`
	var q rpmOstreeVersionData
	if err := yaml.Unmarshal([]byte(verdata), &q); err != nil {
		panic(err)
	}

	assert.Equal(t, "2022.10", q.Root.Version)
	assert.Contains(t, q.Root.Features, "rust")
	assert.NotContains(t, q.Root.Features, "container")
}

func TestParseWorkstation(t *testing.T) {
	var s Status

	if err := json.Unmarshal([]byte(workstationFixture), &s); err != nil {
		panic(err)
	}

	booted, err := s.GetBootedDeployment()
	assert.Nil(t, err)
	assert.NotNil(t, booted)

	assert.Equal(t, booted.GetBaseChecksum(), "229387d3c0bb8ad698228ca5702eca72aed8b298a7c800be1dc72bab160a9f7f")
}

func TestParseFcos(t *testing.T) {
	var s Status

	if err := json.Unmarshal([]byte(fcosFixture), &s); err != nil {
		panic(err)
	}

	booted, err := s.GetBootedDeployment()
	assert.Nil(t, err)
	assert.NotNil(t, booted)
	assert.Equal(t, booted.ContainerImageReference, "")

	assert.Equal(t, booted.GetBaseChecksum(), "a465c49fef185f8339d3cd5857e28386cfdc6516f68206912917c9dc3192d809")

	firstDeploy := s.Deployments[0]
	assert.Equal(t, firstDeploy.ContainerImageReference, "ostree-unverified-registry:quay.io/fedora/fedora-coreos:testing-devel")
}

// Stubbed out tests below that depend on a running rpm-ostree system

// func TestRpmOstreeVersion(t *testing.T) {
// 	c := NewClient("test")
// 	v, err := c.RpmOstreeVersion()
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println(v)
// }
