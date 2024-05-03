package adapter

import (
	"github.com/rwirdemann/datafrog/app/domain"
)

type MemExpectationSource struct {
	testcase domain.Testcase
}

func NewMemExpectationSource(expectations []domain.Expectation) *MemExpectationSource {
	return &MemExpectationSource{testcase: domain.Testcase{Expectations: expectations}}
}

func (es MemExpectationSource) Get() domain.Testcase {
	return es.testcase
}

func (es *MemExpectationSource) Write(testcase domain.Testcase) error {
	es.testcase = testcase
	return nil
}
