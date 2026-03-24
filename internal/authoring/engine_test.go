// SPDX-License-Identifier: Apache-2.0

package authoring

import (
	"strings"
	"testing"

	"github.com/hbraswelrh/pacman/internal/consts"
)

// T511: NewAuthoringEngine creates an engine with correct
// template, initial step index at 0, and role context.
func TestNewAuthoringEngine(t *testing.T) {
	t.Parallel()
	tmpl := threatCatalogTemplate()
	engine := NewAuthoringEngine(
		tmpl,
		"Security Engineer",
		[]string{"threat modeling", "penetration testing"},
		nil,
	)
	if engine == nil {
		t.Fatal("NewAuthoringEngine returned nil")
	}
	if engine.currentStep != 0 {
		t.Errorf(
			"currentStep = %d, want 0",
			engine.currentStep,
		)
	}
	if engine.roleName != "Security Engineer" {
		t.Errorf(
			"roleName = %q, want %q",
			engine.roleName, "Security Engineer",
		)
	}
	if engine.IsComplete() {
		t.Error("engine should not be complete initially")
	}
}

// T512: CurrentStep returns the first step with
// role-personalized explanation.
func TestCurrentStep(t *testing.T) {
	t.Parallel()
	tmpl := threatCatalogTemplate()
	engine := NewAuthoringEngine(
		tmpl,
		"Security Engineer",
		[]string{"threat modeling"},
		nil,
	)
	step := engine.CurrentStep()
	if step == nil {
		t.Fatal("CurrentStep returned nil")
	}
	if step.Name != consts.SectionMetadata {
		t.Errorf(
			"step name = %q, want %q",
			step.Name, consts.SectionMetadata,
		)
	}
	// Role-personalized explanation should mention the
	// role.
	if !strings.Contains(
		step.RoleExplanation, "Security Engineer",
	) {
		t.Errorf(
			"RoleExplanation should contain role "+
				"name, got %q",
			step.RoleExplanation,
		)
	}
}

// T513: SetFieldValue records value for a field in the
// current step; returns error for unknown field name.
func TestEngineSetFieldValue(t *testing.T) {
	t.Parallel()
	tmpl := threatCatalogTemplate()
	engine := NewAuthoringEngine(
		tmpl, "Security Engineer", nil, nil,
	)
	err := engine.SetFieldValue("name", "my-catalog")
	if err != nil {
		t.Fatalf("SetFieldValue returned error: %v", err)
	}

	// Unknown field returns error.
	err = engine.SetFieldValue(
		"nonexistent_field", "value",
	)
	if err == nil {
		t.Error(
			"SetFieldValue should return error " +
				"for unknown field",
		)
	}
}

// T514: CompleteStep advances to next step after all
// required fields are filled.
func TestCompleteStepAdvances(t *testing.T) {
	t.Parallel()
	tmpl := threatCatalogTemplate()
	engine := NewAuthoringEngine(
		tmpl, "Security Engineer", nil, nil,
	)

	// Fill all required fields for metadata step.
	fillMetadataFields(t, engine)

	errs, err := engine.CompleteStep()
	if err != nil {
		t.Fatalf("CompleteStep returned error: %v", err)
	}
	if len(errs) != 0 {
		t.Errorf(
			"CompleteStep returned %d errors, want 0",
			len(errs),
		)
	}
	// Should now be on the next step.
	step := engine.CurrentStep()
	if step == nil {
		t.Fatal("CurrentStep returned nil after advance")
	}
	if step.Name != consts.SectionScope {
		t.Errorf(
			"next step = %q, want %q",
			step.Name, consts.SectionScope,
		)
	}
}

// T515: CompleteStep with missing required fields returns
// validation errors listing the missing fields.
func TestCompleteStepMissingFields(t *testing.T) {
	t.Parallel()
	tmpl := threatCatalogTemplate()
	engine := NewAuthoringEngine(
		tmpl, "Security Engineer", nil, nil,
	)

	// Do not fill required fields. CompleteStep should
	// report them.
	errs, err := engine.CompleteStep()
	if err != nil {
		t.Fatalf("CompleteStep returned error: %v", err)
	}
	if len(errs) == 0 {
		t.Error(
			"CompleteStep should return errors for " +
				"missing required fields",
		)
	}
	// Should still be on the same step.
	step := engine.CurrentStep()
	if step.Name != consts.SectionMetadata {
		t.Errorf(
			"should remain on %q, got %q",
			consts.SectionMetadata, step.Name,
		)
	}
}

// T516: GetSuggestions returns example values from the
// template for the current role.
func TestGetSuggestions(t *testing.T) {
	t.Parallel()
	tmpl := threatCatalogTemplate()
	engine := NewAuthoringEngine(
		tmpl, "Security Engineer", nil, nil,
	)
	suggestions := engine.GetSuggestions("name")
	if len(suggestions) == 0 {
		t.Error(
			"GetSuggestions should return at least " +
				"one suggestion",
		)
	}
}

// T517: Progress returns correct completed/total counts.
func TestProgress(t *testing.T) {
	t.Parallel()
	tmpl := threatCatalogTemplate()
	engine := NewAuthoringEngine(
		tmpl, "Security Engineer", nil, nil,
	)
	completed, total := engine.Progress()
	if completed != 0 {
		t.Errorf("completed = %d, want 0", completed)
	}
	if total != len(tmpl.Steps) {
		t.Errorf(
			"total = %d, want %d",
			total, len(tmpl.Steps),
		)
	}

	// Complete the first step.
	fillMetadataFields(t, engine)
	_, err := engine.CompleteStep()
	if err != nil {
		t.Fatalf("CompleteStep error: %v", err)
	}
	completed, _ = engine.Progress()
	if completed != 1 {
		t.Errorf("completed = %d, want 1", completed)
	}
}

// T518: BuildArtifact assembles all completed sections
// into an AuthoredArtifact with correct metadata.
func TestBuildArtifact(t *testing.T) {
	t.Parallel()
	tmpl := threatCatalogTemplate()
	engine := NewAuthoringEngine(
		tmpl, "Security Engineer", nil, nil,
	)

	// Complete all steps.
	completeAllSteps(t, engine)

	artifact := engine.BuildArtifact()
	if artifact == nil {
		t.Fatal("BuildArtifact returned nil")
	}
	if artifact.ArtifactType !=
		consts.ArtifactThreatCatalog {
		t.Errorf(
			"ArtifactType = %q, want %q",
			artifact.ArtifactType,
			consts.ArtifactThreatCatalog,
		)
	}
	if artifact.AuthoringRole != "Security Engineer" {
		t.Errorf(
			"AuthoringRole = %q, want %q",
			artifact.AuthoringRole,
			"Security Engineer",
		)
	}
	if len(artifact.Sections) == 0 {
		t.Error("artifact should have sections")
	}
}

// T519: ArtifactTemplates returns templates for all
// supported artifact types with non-empty step lists.
func TestArtifactTemplates(t *testing.T) {
	t.Parallel()
	templates := ArtifactTemplates()
	if len(templates) != 12 {
		t.Fatalf(
			"ArtifactTemplates length = %d, want 12",
			len(templates),
		)
	}
	for name, tmpl := range templates {
		if len(tmpl.Steps) == 0 {
			t.Errorf(
				"template %q has no steps", name,
			)
		}
	}
}

// T520: ThreatCatalog template has steps for metadata,
// scope, capabilities, and threats sections in order.
func TestThreatCatalogTemplate(t *testing.T) {
	t.Parallel()
	templates := ArtifactTemplates()
	tmpl, ok := templates[consts.ArtifactThreatCatalog]
	if !ok {
		t.Fatal("ThreatCatalog template not found")
	}
	expected := []string{
		consts.SectionMetadata,
		consts.SectionScope,
		consts.SectionCapabilities,
		consts.SectionThreats,
	}
	if len(tmpl.Steps) != len(expected) {
		t.Fatalf(
			"steps = %d, want %d",
			len(tmpl.Steps), len(expected),
		)
	}
	for i, name := range expected {
		if tmpl.Steps[i].Name != name {
			t.Errorf(
				"Step[%d].Name = %q, want %q",
				i, tmpl.Steps[i].Name, name,
			)
		}
	}
}

// T521: Engine with role "Security Engineer" personalizes
// step explanations to reference security concerns.
func TestPersonalizedExplanation(t *testing.T) {
	t.Parallel()
	templates := ArtifactTemplates()
	tmpl := templates[consts.ArtifactThreatCatalog]
	engine := NewAuthoringEngine(
		tmpl,
		"Security Engineer",
		[]string{"threat modeling"},
		nil,
	)
	step := engine.CurrentStep()
	if step.RoleExplanation == "" {
		t.Error("RoleExplanation should not be empty")
	}
	if !strings.Contains(
		step.RoleExplanation, "Security Engineer",
	) {
		t.Errorf(
			"RoleExplanation should mention role, "+
				"got %q",
			step.RoleExplanation,
		)
	}
}

// T522: CompleteStep on the last step sets IsComplete flag.
func TestCompleteStepLastStep(t *testing.T) {
	t.Parallel()
	tmpl := threatCatalogTemplate()
	engine := NewAuthoringEngine(
		tmpl, "Security Engineer", nil, nil,
	)

	completeAllSteps(t, engine)

	if !engine.IsComplete() {
		t.Error("engine should be complete")
	}
	// CurrentStep should return nil when complete.
	step := engine.CurrentStep()
	if step != nil {
		t.Error(
			"CurrentStep should return nil when " +
				"complete",
		)
	}
}

// threatCatalogTemplate returns a minimal template for
// testing.
func threatCatalogTemplate() ArtifactTemplate {
	return ArtifactTemplate{
		ArtifactType: consts.ArtifactThreatCatalog,
		Steps: []AuthoringStep{
			{
				Name:        consts.SectionMetadata,
				Description: "Define artifact metadata",
				Fields: []StepField{
					{
						Name:         "name",
						Description:  "Artifact name",
						FieldType:    "string",
						Required:     true,
						ExampleValue: "ACME.WEB.THR01",
					},
					{
						Name:        "description",
						Description: "Artifact description",
						FieldType:   "string",
						Required:    true,
						ExampleValue: "Threat catalog " +
							"for web application",
					},
				},
			},
			{
				Name:        consts.SectionScope,
				Description: "Define the scope",
				Fields: []StepField{
					{
						Name:         "scope",
						Description:  "Scope definition",
						FieldType:    "string",
						Required:     true,
						ExampleValue: "Web application",
					},
				},
			},
			{
				Name:        consts.SectionCapabilities,
				Description: "Define capabilities",
				Fields: []StepField{
					{
						Name:         "capability",
						Description:  "Capability name",
						FieldType:    "string",
						Required:     true,
						ExampleValue: "Authentication",
					},
				},
			},
			{
				Name:        consts.SectionThreats,
				Description: "Define threats",
				Fields: []StepField{
					{
						Name:         "threat",
						Description:  "Threat description",
						FieldType:    "string",
						Required:     true,
						ExampleValue: "SQL Injection",
					},
				},
			},
		},
		Layer: consts.LayerThreatsControls,
	}
}

// fillMetadataFields fills all required fields for the
// metadata step.
func fillMetadataFields(t *testing.T, e *AuthoringEngine) {
	t.Helper()
	if err := e.SetFieldValue(
		"name", "ACME.WEB.THR01",
	); err != nil {
		t.Fatalf("SetFieldValue name: %v", err)
	}
	if err := e.SetFieldValue(
		"description", "Test threat catalog",
	); err != nil {
		t.Fatalf("SetFieldValue description: %v", err)
	}
}

// completeAllSteps walks through all steps filling required
// fields and completing each one.
func completeAllSteps(t *testing.T, e *AuthoringEngine) {
	t.Helper()
	for !e.IsComplete() {
		step := e.CurrentStep()
		if step == nil {
			break
		}
		for _, f := range step.Fields {
			if f.Required {
				if err := e.SetFieldValue(
					f.Name, f.ExampleValue,
				); err != nil {
					t.Fatalf(
						"SetFieldValue %s: %v",
						f.Name, err,
					)
				}
			}
		}
		errs, err := e.CompleteStep()
		if err != nil {
			t.Fatalf("CompleteStep error: %v", err)
		}
		if len(errs) != 0 {
			t.Fatalf(
				"CompleteStep returned %d errors",
				len(errs),
			)
		}
	}
}
