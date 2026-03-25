// SPDX-License-Identifier: Apache-2.0

// Package authoring implements guided Gemara content
// authoring for the Gemara User Journey tutorial engine (US6). It
// provides step-by-step artifact creation with role-aware
// guidance, schema validation at each step, and YAML/JSON
// output generation.
package authoring

import (
	"fmt"
	"time"

	"github.com/hbraswelrh/gemara-user-journey/internal/consts"
)

// ValidationStatus tracks the validation state of an
// authored artifact.
type ValidationStatus string

const (
	// StatusNotValidated means validation has not been
	// run yet.
	StatusNotValidated ValidationStatus = ValidationStatus(
		consts.ValidationStatusNotValidated,
	)
	// StatusPartial means the artifact has been partially
	// validated (step-level).
	StatusPartial ValidationStatus = ValidationStatus(
		consts.ValidationStatusPartial,
	)
	// StatusValid means the artifact passed full
	// validation.
	StatusValid ValidationStatus = ValidationStatus(
		consts.ValidationStatusValid,
	)
	// StatusInvalid means the artifact failed validation.
	StatusInvalid ValidationStatus = ValidationStatus(
		consts.ValidationStatusInvalid,
	)
)

// StepField describes a single field within an authoring
// step. Each field has metadata for guidance and validation.
type StepField struct {
	// Name is the field identifier used in the artifact
	// structure.
	Name string `yaml:"name"`
	// Description explains what this field represents.
	Description string `yaml:"description"`
	// FieldType is the expected value type (string, list,
	// map, enum).
	FieldType string `yaml:"field_type"`
	// Required indicates whether this field must be filled
	// before the step can be completed.
	Required bool `yaml:"required"`
	// ExampleValue shows a sample value from Gemara
	// tutorials.
	ExampleValue string `yaml:"example_value"`
	// HelpText provides additional guidance sourced from
	// content blocks.
	HelpText string `yaml:"help_text"`
}

// AuthoringStep represents a single step in the guided
// authoring flow. Each step corresponds to a section of
// the artifact being authored.
type AuthoringStep struct {
	// Name identifies the step (matches a section name).
	Name string `yaml:"name"`
	// Description explains what this step accomplishes.
	Description string `yaml:"description"`
	// RoleExplanation is a role-specific "why this
	// matters" annotation.
	RoleExplanation string `yaml:"role_explanation"`
	// Fields lists the fields to be filled in this step.
	Fields []StepField `yaml:"fields"`
	// Completed indicates whether this step has been
	// finished.
	Completed bool `yaml:"completed"`
}

// ArtifactSection holds the completed field values for a
// section of an authored artifact.
type ArtifactSection struct {
	// Name is the section identifier (e.g., "metadata",
	// "scope").
	Name string `yaml:"name"`
	// Fields maps field names to their entered values.
	Fields map[string]string `yaml:"fields"`
}

// AuthoredArtifact is a Gemara-conformant document produced
// through the guided authoring flow.
type AuthoredArtifact struct {
	// ArtifactType is the Gemara artifact type (e.g.,
	// "ThreatCatalog").
	ArtifactType string `yaml:"artifact_type"`
	// SchemaDef is the target CUE schema definition (e.g.,
	// "#ThreatCatalog").
	SchemaDef string `yaml:"schema_def"`
	// SchemaVersion is the Gemara schema version used for
	// validation.
	SchemaVersion string `yaml:"schema_version"`
	// Sections holds the completed artifact sections.
	Sections []ArtifactSection `yaml:"sections"`
	// Status tracks the validation state.
	Status ValidationStatus `yaml:"status"`
	// AuthoringRole is the role of the user who authored
	// this artifact.
	AuthoringRole string `yaml:"authoring_role"`
	// CreatedAt is when authoring began.
	CreatedAt time.Time `yaml:"created_at"`
	// UpdatedAt is when the artifact was last modified.
	UpdatedAt time.Time `yaml:"updated_at"`
}

// ArtifactTemplate defines the authoring recipe for a
// specific artifact type. It provides the ordered sequence
// of steps, the fields within each step, and tutorial
// references.
type ArtifactTemplate struct {
	// ArtifactType is the Gemara artifact type this
	// template produces.
	ArtifactType string `yaml:"artifact_type"`
	// Steps defines the ordered authoring steps.
	Steps []AuthoringStep `yaml:"steps"`
	// Layer is the primary Gemara layer for this artifact
	// type.
	Layer int `yaml:"layer"`
	// TutorialRefs lists tutorial file paths relevant to
	// this artifact type.
	TutorialRefs []string `yaml:"tutorial_refs"`
}

// NewAuthoredArtifact creates a new AuthoredArtifact with
// the given type, schema definition, schema version, and
// authoring role. The artifact starts with no sections and
// a status of NotValidated.
func NewAuthoredArtifact(
	artifactType string,
	schemaDef string,
	schemaVersion string,
	authoringRole string,
) *AuthoredArtifact {
	now := time.Now()
	return &AuthoredArtifact{
		ArtifactType:  artifactType,
		SchemaDef:     schemaDef,
		SchemaVersion: schemaVersion,
		Sections:      []ArtifactSection{},
		Status:        StatusNotValidated,
		AuthoringRole: authoringRole,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// AddSection appends a new section with the given name and
// empty field values.
func (a *AuthoredArtifact) AddSection(name string) {
	a.Sections = append(a.Sections, ArtifactSection{
		Name:   name,
		Fields: make(map[string]string),
	})
	a.UpdatedAt = time.Now()
}

// SetFieldValue records a value for a named field within
// the specified section. Returns an error if the section
// is not found.
func (a *AuthoredArtifact) SetFieldValue(
	sectionName string,
	fieldName string,
	value string,
) error {
	for i := range a.Sections {
		if a.Sections[i].Name == sectionName {
			a.Sections[i].Fields[fieldName] = value
			a.UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf(
		"section %q not found in artifact", sectionName,
	)
}

// SetStatus updates the validation status of the artifact.
func (a *AuthoredArtifact) SetStatus(s ValidationStatus) {
	a.Status = s
	a.UpdatedAt = time.Now()
}

// artifactTypeSchemaMap maps artifact type names to their
// CUE schema definition strings.
var artifactTypeSchemaMap = map[string]string{
	consts.ArtifactGuidanceCatalog:   consts.SchemaGuidanceCatalog,
	consts.ArtifactVectorCatalog:     consts.SchemaVectorCatalog,
	consts.ArtifactPrincipleCatalog:  consts.SchemaPrincipleCatalog,
	consts.ArtifactControlCatalog:    consts.SchemaControlCatalog,
	consts.ArtifactThreatCatalog:     consts.SchemaThreatCatalog,
	consts.ArtifactCapabilityCatalog: consts.SchemaCapabilityCatalog,
	consts.ArtifactPolicy:            consts.SchemaPolicy,
	consts.ArtifactRiskCatalog:       consts.SchemaRiskCatalog,
	consts.ArtifactMappingDocument:   consts.SchemaMappingDocument,
	consts.ArtifactEvaluationLog:     consts.SchemaEvaluationLog,
	consts.ArtifactEnforcementLog:    consts.SchemaEnforcementLog,
	consts.ArtifactAuditLog:          consts.SchemaAuditLog,
}

// ArtifactTypeToSchema returns the CUE schema definition
// string for the given artifact type. Returns the schema
// and true if found, or an empty string and false if the
// type is not recognized.
func ArtifactTypeToSchema(
	artifactType string,
) (string, bool) {
	schema, ok := artifactTypeSchemaMap[artifactType]
	return schema, ok
}

// SupportedArtifactTypes returns the list of artifact types
// that have published CUE schemas and can be authored.
func SupportedArtifactTypes() []string {
	return []string{
		consts.ArtifactGuidanceCatalog,
		consts.ArtifactVectorCatalog,
		consts.ArtifactPrincipleCatalog,
		consts.ArtifactControlCatalog,
		consts.ArtifactThreatCatalog,
		consts.ArtifactCapabilityCatalog,
		consts.ArtifactPolicy,
		consts.ArtifactRiskCatalog,
		consts.ArtifactMappingDocument,
		consts.ArtifactEvaluationLog,
		consts.ArtifactEnforcementLog,
		consts.ArtifactAuditLog,
	}
}

// IncompleteFields returns the names of required fields in
// the given step that have no value (empty or absent) in
// the provided values map.
func IncompleteFields(
	step AuthoringStep,
	values map[string]string,
) []string {
	var missing []string
	for _, f := range step.Fields {
		if f.Required {
			v, ok := values[f.Name]
			if !ok || v == "" {
				missing = append(missing, f.Name)
			}
		}
	}
	return missing
}
