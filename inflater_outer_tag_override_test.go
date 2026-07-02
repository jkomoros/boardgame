package boardgame

import (
	"testing"

	"github.com/workfit/tester/assert"
)

// Synthetic fixtures for the outer-embedding-site struct-tag override
// mechanism added in P2.1 (spec §6.3.2). Independent of any specific
// behavior so the inflater extension can be verified in isolation.

type innerBehaviorWithTag struct {
	Color string `sanitize:"other:hidden" enum:"color"`
}

type structEmbeddingWithoutOverride struct {
	innerBehaviorWithTag
}

type structEmbeddingWithOverride struct {
	innerBehaviorWithTag `sanitize:"all:visible"`
}

type structEmbeddingWithUnrelatedTag struct {
	innerBehaviorWithTag `something:"else"`
}

func TestStructTagsForField_InnerTagWhenNoOuter(t *testing.T) {
	got := structTagsForField(&structEmbeddingWithoutOverride{}, "Color", []string{"sanitize", "enum"})
	assert.For(t).ThatActual(got["sanitize"]).Equals("other:hidden")
	assert.For(t).ThatActual(got["enum"]).Equals("color")
}

func TestStructTagsForField_OuterOverridesInner(t *testing.T) {
	got := structTagsForField(&structEmbeddingWithOverride{}, "Color", []string{"sanitize", "enum"})
	// Outer sanitize override should win.
	assert.For(t).ThatActual(got["sanitize"]).Equals("all:visible")
	// Inner enum tag unchanged (outer didn't set enum).
	assert.For(t).ThatActual(got["enum"]).Equals("color")
}

func TestStructTagsForField_UnrelatedOuterTagDoesNotMask(t *testing.T) {
	got := structTagsForField(&structEmbeddingWithUnrelatedTag{}, "Color", []string{"sanitize", "enum"})
	// Outer has only `something:`; the requested tags fall through to inner.
	assert.For(t).ThatActual(got["sanitize"]).Equals("other:hidden")
	assert.For(t).ThatActual(got["enum"]).Equals("color")
}

func TestStructTagsForField_DirectFieldUnaffected(t *testing.T) {
	type directField struct {
		Color string `sanitize:"direct"`
	}
	got := structTagsForField(&directField{}, "Color", []string{"sanitize"})
	// Direct field (no embedding) — current behavior is preserved.
	assert.For(t).ThatActual(got["sanitize"]).Equals("direct")
}

func TestStructTagsForField_MissingField(t *testing.T) {
	got := structTagsForField(&structEmbeddingWithoutOverride{}, "DoesNotExist", []string{"sanitize"})
	// Missing field returns empty string (existing behavior).
	assert.For(t).ThatActual(got["sanitize"]).Equals("")
}

func TestOuterEmbeddingTags_NilType(t *testing.T) {
	got := outerEmbeddingTags(nil, "Color", []string{"sanitize"})
	assert.For(t).ThatActual(len(got)).Equals(0)
}
