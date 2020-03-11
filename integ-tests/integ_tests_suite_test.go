package template_generation_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIntegTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IntegTests Suite")
}

func BeforeAll(fn func()) {
	first := true
	BeforeEach(func() {
		if first {
			fn()
			first = false
		}
	})
}
