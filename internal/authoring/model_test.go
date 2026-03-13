// SPDX-License-Identifier: Apache-2.0

package authoring

import (
	"testing"

	"github.com/hbraswelrh/pacman/internal/consts"
)

// T501: NewAuthoredArtifact creates an artifact with the
// given type, schema definition, schema version, and empty
// sections.
func TestNewAuthoredArtifact(t *testing.T) {
	t.Parallel()
	a := NewAuthoredArtifact(
		consts.ArtifactThreatCatalog,
		consts.SchemaThreatCatalog,
		"v0.20.0",
		"Security Engineer",
	)
	if a.ArtifactType != consts.ArtifactThreatCatalog {
		t.Errorf(
			"ArtifactType = %q, want %q",
			a.ArtifactType,
			consts.ArtifactThreatCatalog,
		)
	}
	if a.SchemaDef != consts.SchemaThreatCatalog {
		t.Errorf(
			"SchemaDef = %q, want %q",
			a.SchemaDef,
			consts.SchemaThreatCatalog,
		)
	}
	if a.SchemaVersion != "v0.20.0" {
		t.Errorf(
			"SchemaVersion = %q, want %q",
			a.SchemaVersion, "v0.20.0",
		)
	}
	if a.AuthoringRole != "Security Engineer" {
		t.Errorf(
			"AuthoringRole = %q, want %q",
			a.AuthoringRole, "Security Engineer",
		)
	}
	if len(a.Sections) != 0 {
		t.Errorf(
			"Sections length = %d, want 0",
			len(a.Sections),
		)
	}
	if a.Status != ValidationStatus(
		consts.ValidationStatusNotValidated,
	) {
		t.Errorf(
			"Status = %q, want %q",
			a.Status,
			consts.ValidationStatusNotValidated,
		)
	}
	if a.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

// T502: ArtifactTypeToSchema maps each supported artifact
// type to its CUE schema definition.
func TestArtifactTypeToSchema(t *testing.T) {
	t.Parallel()
	tests := []struct {
		artifactType string
		wantSchema   string
	}{
		{
			consts.ArtifactGuidanceCatalog,
			consts.SchemaGuidanceCatalog,
		},
		{
			consts.ArtifactControlCatalog,
			consts.SchemaControlCatalog,
		},
		{
			consts.ArtifactThreatCatalog,
			consts.SchemaThreatCatalog,
		},
		{
			consts.ArtifactPolicy,
			consts.SchemaPolicy,
		},
		{
			consts.ArtifactMappingDocument,
			consts.SchemaMappingDocument,
		},
		{
			consts.ArtifactEvaluationLog,
			consts.SchemaEvaluationLog,
		},
	}
	for _, tt := range tests {
		got, ok := ArtifactTypeToSchema(tt.artifactType)
		if !ok {
			t.Errorf(
				"ArtifactTypeToSchema(%q) not found",
				tt.artifactType,
			)
			continue
		}
		if got != tt.wantSchema {
			t.Errorf(
				"ArtifactTypeToSchema(%q) = %q, "+
					"want %q",
				tt.artifactType, got,
				tt.wantSchema,
			)
		}
	}
}

// T502 (negative): unknown type returns false.
func TestArtifactTypeToSchemaUnknown(t *testing.T) {
	t.Parallel()
	_, ok := ArtifactTypeToSchema("UnknownType")
	if ok {
		t.Error(
			"ArtifactTypeToSchema should return " +
				"false for unknown type",
		)
	}
}

// T503: SupportedArtifactTypes returns the six artifact
// types that have published CUE schemas.
func TestSupportedArtifactTypes(t *testing.T) {
	t.Parallel()
	types := SupportedArtifactTypes()
	if len(types) != 6 {
		t.Fatalf(
			"SupportedArtifactTypes length = %d, "+
				"want 6",
			len(types),
		)
	}
	expected := map[string]bool{
		consts.ArtifactGuidanceCatalog: true,
		consts.ArtifactControlCatalog:  true,
		consts.ArtifactThreatCatalog:   true,
		consts.ArtifactPolicy:          true,
		consts.ArtifactMappingDocument: true,
		consts.ArtifactEvaluationLog:   true,
	}
	for _, at := range types {
		if !expected[at] {
			t.Errorf(
				"unexpected artifact type: %q", at,
			)
		}
	}
}

// T504: AddSection appends an ArtifactSection with the
// given name and empty field values.
func TestAddSection(t *testing.T) {
	t.Parallel()
	a := NewAuthoredArtifact(
		consts.ArtifactThreatCatalog,
		consts.SchemaThreatCatalog,
		"v0.20.0",
		"Security Engineer",
	)
	a.AddSection(consts.SectionMetadata)
	if len(a.Sections) != 1 {
		t.Fatalf(
			"Sections length = %d, want 1",
			len(a.Sections),
		)
	}
	if a.Sections[0].Name != consts.SectionMetadata {
		t.Errorf(
			"Section name = %q, want %q",
			a.Sections[0].Name,
			consts.SectionMetadata,
		)
	}
	if len(a.Sections[0].Fields) != 0 {
		t.Errorf(
			"Fields length = %d, want 0",
			len(a.Sections[0].Fields),
		)
	}
}

// T505: SetFieldValue records a value for a named field
// within a section; returns error for unknown section.
func TestSetFieldValue(t *testing.T) {
	t.Parallel()
	a := NewAuthoredArtifact(
		consts.ArtifactThreatCatalog,
		consts.SchemaThreatCatalog,
		"v0.20.0",
		"Security Engineer",
	)
	a.AddSection(consts.SectionMetadata)

	err := a.SetFieldValue(
		consts.SectionMetadata, "name", "my-catalog",
	)
	if err != nil {
		t.Fatalf("SetFieldValue returned error: %v", err)
	}
	if a.Sections[0].Fields["name"] != "my-catalog" {
		t.Errorf(
			"Fields[name] = %q, want %q",
			a.Sections[0].Fields["name"],
			"my-catalog",
		)
	}

	// Unknown section returns error.
	err = a.SetFieldValue(
		"nonexistent", "name", "value",
	)
	if err == nil {
		t.Error(
			"SetFieldValue should return error " +
				"for unknown section",
		)
	}
}

// T506: ValidationStatus transitions.
func TestValidationStatusTransitions(t *testing.T) {
	t.Parallel()
	a := NewAuthoredArtifact(
		consts.ArtifactThreatCatalog,
		consts.SchemaThreatCatalog,
		"v0.20.0",
		"Security Engineer",
	)
	if a.Status != StatusNotValidated {
		t.Errorf(
			"initial status = %q, want %q",
			a.Status, StatusNotValidated,
		)
	}

	a.SetStatus(StatusPartial)
	if a.Status != StatusPartial {
		t.Errorf(
			"status = %q, want %q",
			a.Status, StatusPartial,
		)
	}

	a.SetStatus(StatusValid)
	if a.Status != StatusValid {
		t.Errorf(
			"status = %q, want %q",
			a.Status, StatusValid,
		)
	}

	a.SetStatus(StatusInvalid)
	if a.Status != StatusInvalid {
		t.Errorf(
			"status = %q, want %q",
			a.Status, StatusInvalid,
		)
	}
}

// T507: StepField with Required=true and empty value is
// reported by IncompleteFields.
func TestIncompleteFields(t *testing.T) {
	t.Parallel()
	step := AuthoringStep{
		Name: "metadata",
		Fields: []StepField{
			{
				Name:     "name",
				Required: true,
			},
			{
				Name:     "description",
				Required: true,
			},
			{
				Name:     "optional_tag",
				Required: false,
			},
		},
	}

	// No values set — both required fields incomplete.
	values := map[string]string{}
	incomplete := IncompleteFields(step, values)
	if len(incomplete) != 2 {
		t.Fatalf(
			"IncompleteFields = %d, want 2",
			len(incomplete),
		)
	}

	// Set one required field.
	values["name"] = "my-catalog"
	incomplete = IncompleteFields(step, values)
	if len(incomplete) != 1 {
		t.Fatalf(
			"IncompleteFields = %d, want 1",
			len(incomplete),
		)
	}
	if incomplete[0] != "description" {
		t.Errorf(
			"incomplete field = %q, want %q",
			incomplete[0], "description",
		)
	}

	// Set both required fields — none incomplete.
	values["description"] = "A threat catalog"
	incomplete = IncompleteFields(step, values)
	if len(incomplete) != 0 {
		t.Errorf(
			"IncompleteFields = %d, want 0",
			len(incomplete),
		)
	}
}

// T508: ArtifactTemplate defines ordered steps for a
// ThreatCatalog with correct section names.
func TestArtifactTemplateStructure(t *testing.T) {
	t.Parallel()
	tmpl := ArtifactTemplate{
		ArtifactType: consts.ArtifactThreatCatalog,
		Steps: []AuthoringStep{
			{Name: consts.SectionMetadata},
			{Name: consts.SectionScope},
			{Name: consts.SectionCapabilities},
			{Name: consts.SectionThreats},
		},
		Layer: consts.LayerThreatsControls,
	}
	if tmpl.ArtifactType != consts.ArtifactThreatCatalog {
		t.Errorf(
			"ArtifactType = %q, want %q",
			tmpl.ArtifactType,
			consts.ArtifactThreatCatalog,
		)
	}
	if len(tmpl.Steps) != 4 {
		t.Fatalf(
			"Steps length = %d, want 4",
			len(tmpl.Steps),
		)
	}
	expectedSteps := []string{
		consts.SectionMetadata,
		consts.SectionScope,
		consts.SectionCapabilities,
		consts.SectionThreats,
	}
	for i, step := range tmpl.Steps {
		if step.Name != expectedSteps[i] {
			t.Errorf(
				"Step[%d].Name = %q, want %q",
				i, step.Name, expectedSteps[i],
			)
		}
	}
	if tmpl.Layer != consts.LayerThreatsControls {
		t.Errorf(
			"Layer = %d, want %d",
			tmpl.Layer, consts.LayerThreatsControls,
		)
	}
}
