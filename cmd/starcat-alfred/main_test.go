package main

import "testing"

func TestStableErrorCodeIgnoresHumanMessage(t *testing.T) {
	message := "starcat: STARCAT_ERROR REQUIRES_PRO: wording may change"
	if code := stableErrorCode(message); code != "REQUIRES_PRO" {
		t.Fatalf("stableErrorCode() = %q", code)
	}
}
