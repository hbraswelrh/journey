// SPDX-License-Identifier: Apache-2.0

package roles

import (
	"testing"

	"github.com/hbraswelrh/pacman/internal/consts"
)

// T201: PredefinedRoles returns the required minimum list.
func TestPredefinedRolesContainsRequiredRoles(t *testing.T) {
	t.Parallel()

	roles := PredefinedRoles()

	required := []string{
		consts.RoleSecurityEngineer,
		consts.RoleComplianceOfficer,
		consts.RoleCISO,
		consts.RoleDeveloper,
		consts.RolePlatformEngineer,
		consts.RolePolicyAuthor,
		consts.RoleAuditor,
	}

	if len(roles) < len(required) {
		t.Fatalf(
			"expected at least %d predefined roles, "+
				"got %d",
			len(required), len(roles),
		)
	}

	roleNames := make(map[string]bool)
	for _, r := range roles {
		roleNames[r.Name] = true
	}

	for _, name := range required {
		if !roleNames[name] {
			t.Errorf(
				"missing required predefined role: %s",
				name,
			)
		}
	}
}

// T202: Role struct contains required fields.
func TestRoleStructHasRequiredFields(t *testing.T) {
	t.Parallel()

	roles := PredefinedRoles()
	for _, role := range roles {
		if role.Name == "" {
			t.Error("role has empty Name")
		}
		if role.Description == "" {
			t.Errorf(
				"role %s has empty Description",
				role.Name,
			)
		}
		if len(role.DefaultKeywords) == 0 {
			t.Errorf(
				"role %s has no DefaultKeywords",
				role.Name,
			)
		}
		if len(role.DefaultLayers) == 0 {
			t.Errorf(
				"role %s has no DefaultLayers",
				role.Name,
			)
		}
		if role.Source != SourcePredefined {
			t.Errorf(
				"role %s should be SourcePredefined",
				role.Name,
			)
		}
	}
}

// T218: MatchRole with exact predefined name returns exact
// match with high confidence.
func TestMatchRoleExact(t *testing.T) {
	t.Parallel()

	result := MatchRole("Security Engineer")

	if result.Type != MatchExact {
		t.Fatalf(
			"expected MatchExact, got %d", result.Type,
		)
	}
	if result.Role == nil {
		t.Fatal("expected non-nil Role")
	}
	if result.Role.Name != consts.RoleSecurityEngineer {
		t.Errorf(
			"expected %s, got %s",
			consts.RoleSecurityEngineer,
			result.Role.Name,
		)
	}
	if result.Confidence != 1.0 {
		t.Errorf(
			"expected confidence 1.0, got %f",
			result.Confidence,
		)
	}
}

// T219: MatchRole with partial match does NOT assume the
// generic role.
func TestMatchRolePartial(t *testing.T) {
	t.Parallel()

	result := MatchRole("Product Security Engineer")

	if result.Type != MatchPartial {
		t.Fatalf(
			"expected MatchPartial, got %d",
			result.Type,
		)
	}
	if result.Role == nil {
		t.Fatal("expected non-nil Role for partial match")
	}
	// The partial match should identify Security Engineer
	// but with less than 1.0 confidence.
	if result.Confidence >= 1.0 {
		t.Errorf(
			"partial match should have confidence < 1.0,"+
				" got %f",
			result.Confidence,
		)
	}
}

// T220: MatchRole with completely unknown title returns no
// match.
func TestMatchRoleNone(t *testing.T) {
	t.Parallel()

	result := MatchRole("Underwater Basket Weaver")

	if result.Type != MatchNone {
		t.Fatalf(
			"expected MatchNone, got %d", result.Type,
		)
	}
	if result.Role != nil {
		t.Error("expected nil Role for no match")
	}
}

// MatchRole with empty input returns no match.
func TestMatchRoleEmpty(t *testing.T) {
	t.Parallel()

	result := MatchRole("")

	if result.Type != MatchNone {
		t.Fatalf(
			"expected MatchNone for empty input, got %d",
			result.Type,
		)
	}
}

// MatchRole is case-insensitive.
func TestMatchRoleCaseInsensitive(t *testing.T) {
	t.Parallel()

	result := MatchRole("security engineer")

	if result.Type != MatchExact {
		t.Fatalf(
			"expected MatchExact for lowercase input, "+
				"got %d",
			result.Type,
		)
	}
}
