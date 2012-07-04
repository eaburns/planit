package pddl

import (
	"testing"
	"os"
)

const (
	testDomainFile = "p01-domain.pddl"
	testProblemFile = "p01.pddl"
)

func TestParseDomain(t *testing.T) {
	file, err := os.Open(testDomainFile)
	if err != nil {
		t.Error(err)
	}
	_, err = ParseDomain(testDomainFile, file)
	if err != nil {
		t.Error(err)
	}
}

func TestPrintDomain(t *testing.T) {
	file, err := os.Open(testDomainFile)
	if err != nil {
		t.Error(err)
	}
	d, err := ParseDomain(testDomainFile, file)
	PrintDomain(os.Stdout, d)
}

func TestParseProblem(t *testing.T) {
	file, err := os.Open(testProblemFile)
	if err != nil {
		t.Error(err)
	}
	_, err = ParseProblem(testProblemFile, file)
	if err != nil {
		t.Error(err)
	}
}

func TestPrintProblem(t *testing.T) {
	file, err := os.Open(testProblemFile)
	if err != nil {
		t.Error(err)
	}
	p, err := ParseProblem(testProblemFile, file)
	if err != nil {
		t.Error(err)
	}
	PrintProblem(os.Stdout, p)
}
