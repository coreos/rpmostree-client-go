package imgref

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOstreeImageReference(t *testing.T) {
	equivalent := []string{
		"ostree-unverified-image:docker://quay.io/exampleos/blah",
		"ostree-unverified-registry:quay.io/exampleos/blah",
	}
	for _, ir := range equivalent {
		ir, err := Parse(ir)
		assert.Nil(t, err)
		assert.True(t, ir.Sigverify.AllowInsecure)
		assert.Len(t, ir.Sigverify.OstreeRemote, 0)
		assert.Equal(t, ir.Imgref.Transport, "registry")
		assert.Equal(t, ir.Imgref.Image, "quay.io/exampleos/blah")
	}

	equivalent = []string{
		"ostree-remote-registry:fedora:quay.io/fedora/fedora-coreos:stable",
		"ostree-remote-image:fedora:registry:quay.io/fedora/fedora-coreos:stable",
		"ostree-remote-image:fedora:docker://quay.io/fedora/fedora-coreos:stable",
	}
	for _, ir := range equivalent {
		ir, err := Parse(ir)
		assert.Nil(t, err)
		assert.True(t, ir.Sigverify.AllowInsecure)
		assert.Equal(t, ir.Sigverify.OstreeRemote, "fedora")
		assert.Equal(t, ir.Imgref.Transport, "registry")
		assert.Equal(t, ir.Imgref.Image, "quay.io/fedora/fedora-coreos:stable")
	}
}
