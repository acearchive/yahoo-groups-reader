package flowed_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFlowed(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flowed Suite")
}
