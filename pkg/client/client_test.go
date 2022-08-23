package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

// Stubbed out tests below that depend on a running rpm-ostree system

// func TestRpmOstreeVersion(t *testing.T) {
// 	c := NewClient("test")
// 	v, err := c.RpmOstreeVersion()
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println(v)
// }
