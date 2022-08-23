package client

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	c := NewClient("test")
	if c.clientid != "test" {
		panic("mismatched client id")
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
