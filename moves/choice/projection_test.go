package choice

import "testing"

func TestProjectionDescriptorsAreImmutable(t *testing.T) {
	base := EnumValues("Card").Excluding("Unknown")
	first := base.DiscloseExactAvailabilityToActor("public card catalogue")
	second := base.Excluding("Guard").DiscloseExactAvailabilityToActor("different review")

	firstDeclaration := first.Declaration()
	secondDeclaration := second.Declaration()
	if len(firstDeclaration.ExcludedValues) != 1 || firstDeclaration.ExcludedValues[0] != "Unknown" {
		t.Fatalf("first exclusions mutated: %v", firstDeclaration.ExcludedValues)
	}
	if len(secondDeclaration.ExcludedValues) != 2 {
		t.Fatalf("second exclusions = %v", secondDeclaration.ExcludedValues)
	}
	firstDeclaration.ExcludedValues[0] = "changed"
	if first.Declaration().ExcludedValues[0] != "Unknown" {
		t.Fatal("Declaration did not return a defensive copy")
	}
}
