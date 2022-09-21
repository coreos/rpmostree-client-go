package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

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

// Stubbed out tests below that depend on a running rpm-ostree system

// func TestRpmOstreeVersion(t *testing.T) {
// 	c := NewClient("test")
// 	v, err := c.RpmOstreeVersion()
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println(v)
// }
