package domain

import "time"

type Testcase struct {
	Name          string        `json:"name"`
	Running       bool          `json:"running"`
	Verifications int           `json:"verifications"`
	Expectations  []Expectation `json:"expectation"`
	LastExecution time.Time     `json:"last_execution"`

	// Expectations, that match one of the patterns but didn't match one of the
	// expected expectations
	AdditionalExpectations []Expectation `json:"additional_expectations"`
}

// Fulfilled returns the number of fulfilled expectations.
func (t Testcase) Fulfilled() int {
	fulfilled := 0
	for _, e := range t.Expectations {
		if e.Fulfilled {
			fulfilled++
		}
	}
	return fulfilled
}

// Unfulfilled returns the unfulfilled expectations.
func (t Testcase) Unfulfilled() []Expectation {
	var unfulfilled []Expectation
	for _, e := range t.Expectations {
		if !e.Fulfilled {
			unfulfilled = append(unfulfilled, e)
		}
	}
	return unfulfilled
}
