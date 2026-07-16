package main

import (
	"fmt"
	"testing"
)

func TestValidateClientExtractionResultsRejectsDuplicates(t *testing.T) {
	if err := validateClientExtractionResults(nil, []string{"example/game", "example/game"}, "fixture"); err == nil {
		t.Fatal("duplicate extractor result was accepted")
	}
}

func TestStaleGeneratedClientContractsClassificationSurvivesWrapping(t *testing.T) {
	stale := fmt.Errorf("comparison failed: %w", staleGeneratedClientContracts("stale"))
	if !isStaleGeneratedClientContracts(stale) {
		t.Fatal("wrapped stale error was not classified")
	}
	if isStaleGeneratedClientContracts(fmt.Errorf("extractor failed")) {
		t.Fatal("infrastructure error was classified as stale")
	}
}
