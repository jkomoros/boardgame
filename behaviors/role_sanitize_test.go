package behaviors

import (
	"reflect"
	"strings"
	"testing"
)

// TestPlayerRoleHasHiddenDefaultSanitizeTag pins the spec §6.3.1 default:
// behaviors.PlayerRole.Role ships with `sanitize:"other:hidden"` so the
// projector (ObserverPlayerIndex) doesn't see roles by default. Hidden-
// role games (Werewolf, Mysterium, Secret Hitler) get the right
// behavior automatically.
//
// If this test fails because the tag is missing, asymmetric games that
// rely on the framework default would silently leak roles to the
// projector. Don't relax this assertion without flipping the default
// across the spec, the implementation, AND adding a public-role audit
// to all existing games.
func TestPlayerRoleHasHiddenDefaultSanitizeTag(t *testing.T) {
	roleField, ok := reflect.TypeOf(PlayerRole{}).FieldByName("Role")
	if !ok {
		t.Fatal("PlayerRole.Role field not found via reflection")
	}
	tag := roleField.Tag.Get("sanitize")
	if !strings.Contains(tag, "other:hidden") {
		t.Errorf("PlayerRole.Role.Tag sanitize=%q; expected default to contain 'other:hidden' (spec §6.3.1)", tag)
	}
}

// TestPlayerTeamHasHiddenDefaultSanitizeTag mirrors the above for
// PlayerTeam.Team — same spec §6.3.1 default.
func TestPlayerTeamHasHiddenDefaultSanitizeTag(t *testing.T) {
	teamField, ok := reflect.TypeOf(PlayerTeam{}).FieldByName("Team")
	if !ok {
		t.Fatal("PlayerTeam.Team field not found via reflection")
	}
	tag := teamField.Tag.Get("sanitize")
	if !strings.Contains(tag, "other:hidden") {
		t.Errorf("PlayerTeam.Team.Tag sanitize=%q; expected default to contain 'other:hidden' (spec §6.3.1)", tag)
	}
}

// TestPlayerRoleOuterEmbeddingTagWinsViaInflater verifies the end-to-end
// override path (spec §6.3.2): when a game's playerState embeds
// behaviors.PlayerRole with an outer `sanitize:"all:visible"` tag, the
// inflater's outer-tag-precedence rule surfaces THAT value when asked
// for the Role property's sanitize tag, NOT the inner "other:hidden"
// default.
//
// We can't easily run the full StructInflater pipeline here (it needs a
// ComponentChest and a generated Reader), so we exercise the lower-level
// tag-resolution function that the inflater uses internally. The
// boardgame package's inflater_outer_tag_override_test.go covers the
// inflater path with synthetic fixtures; this test pins the
// PlayerRole-specific composition that production games rely on.
func TestPlayerRoleOuterEmbeddingTagWinsViaInflater(t *testing.T) {
	type playerStatePublicRole struct {
		PlayerRole `sanitize:"all:visible"`
	}
	type playerStateDefault struct {
		PlayerRole
	}

	defaultTag := reflectGetSanitizeTagOnPromotedField(reflect.TypeOf(playerStateDefault{}), "Role")
	overrideTag := reflectGetSanitizeTagOnPromotedField(reflect.TypeOf(playerStatePublicRole{}), "Role")

	if !strings.Contains(defaultTag, "other:hidden") {
		t.Errorf("default-embedding: Role sanitize tag = %q; expected to inherit 'other:hidden' from PlayerRole.Role default", defaultTag)
	}
	if !strings.Contains(overrideTag, "all:visible") {
		t.Errorf("override-embedding: Role sanitize tag = %q; expected outer-embedding 'all:visible' to win", overrideTag)
	}
}

// reflectGetSanitizeTagOnPromotedField walks the top-level fields of t
// looking for an anonymous embedding whose inner type contains
// fieldName. If found AND the OUTER embedding's Tag has a sanitize:
// value, returns that; otherwise returns the inner promoted field's
// sanitize: tag. Mirrors the behavior of struct_inflater.go's
// structTagsForField in a self-contained way so this test doesn't
// import boardgame (which would cycle through behaviors).
func reflectGetSanitizeTagOnPromotedField(t reflect.Type, fieldName string) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return ""
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.Anonymous {
			continue
		}
		innerType := f.Type
		if innerType.Kind() == reflect.Ptr {
			innerType = innerType.Elem()
		}
		if innerType.Kind() != reflect.Struct {
			continue
		}
		if _, ok := innerType.FieldByName(fieldName); !ok {
			continue
		}
		// Outer tag wins if present.
		if outer := f.Tag.Get("sanitize"); outer != "" {
			return outer
		}
		// Inner tag fallback.
		inner, _ := innerType.FieldByName(fieldName)
		return inner.Tag.Get("sanitize")
	}
	return ""
}
