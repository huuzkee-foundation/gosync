package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGosync(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gosync Suite!")
}
