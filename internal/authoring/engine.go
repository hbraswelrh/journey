// SPDX-License-Identifier: Apache-2.0

package authoring

import (
	"fmt"
	"strings"

	"github.com/hbraswelrh/gemara-user-journey/internal/consts"
)

// AuthoringEngine manages the step-by-step guided authoring
// flow for a Gemara artifact. It tracks the current step,
// records field values, provides role-personalized guidance,
// and assembles the final artifact.
type AuthoringEngine struct {
	template    ArtifactTemplate
	currentStep int
	fieldValues map[int]map[string]string
	roleName    string
	keywords    []string
	isComplete  bool
}

// NewAuthoringEngine creates an engine for the given
// artifact template with role-specific context. The blocks
// parameter is reserved for future content block
// integration and may be nil.
func NewAuthoringEngine(
	template ArtifactTemplate,
	roleName string,
	keywords []string,
	_ interface{},
) *AuthoringEngine {
	fv := make(map[int]map[string]string)
	for i := range template.Steps {
		fv[i] = make(map[string]string)
	}
	return &AuthoringEngine{
		template:    template,
		currentStep: 0,
		fieldValues: fv,
		roleName:    roleName,
		keywords:    keywords,
		isComplete:  false,
	}
}

// CurrentStep returns the current authoring step with
// role-personalized explanations. Returns nil if authoring
// is complete.
func (e *AuthoringEngine) CurrentStep() *AuthoringStep {
	if e.isComplete ||
		e.currentStep >= len(e.template.Steps) {
		return nil
	}
	step := e.template.Steps[e.currentStep]
	step.RoleExplanation = personalizeExplanation(
		&step, e.roleName,
	)
	return &step
}

// SetFieldValue records a value for a field in the current
// step. Returns an error if the field is not defined in the
// current step or if authoring is complete.
func (e *AuthoringEngine) SetFieldValue(
	fieldName string,
	value string,
) error {
	if e.isComplete {
		return fmt.Errorf("authoring is complete")
	}
	step := e.template.Steps[e.currentStep]
	found := false
	for _, f := range step.Fields {
		if f.Name == fieldName {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf(
			"field %q not found in step %q",
			fieldName, step.Name,
		)
	}
	e.fieldValues[e.currentStep][fieldName] = value
	return nil
}

// ValidationError represents a validation problem with a
// specific field, including a fix suggestion.
type ValidationError struct {
	// FieldPath identifies the field with the error.
	FieldPath string `yaml:"field_path"`
	// Message describes the validation error.
	Message string `yaml:"message"`
	// FixSuggestion provides actionable guidance to fix
	// the error.
	FixSuggestion string `yaml:"fix_suggestion"`
}

// CompleteStep validates the current step and advances to
// the next one if all required fields are filled. Returns
// validation errors for any missing required fields.
func (e *AuthoringEngine) CompleteStep() (
	[]ValidationError, error,
) {
	if e.isComplete {
		return nil, fmt.Errorf("authoring is complete")
	}
	step := e.template.Steps[e.currentStep]
	values := e.fieldValues[e.currentStep]

	// Check for missing required fields.
	missing := IncompleteFields(step, values)
	if len(missing) > 0 {
		var errs []ValidationError
		for _, name := range missing {
			errs = append(errs, ValidationError{
				FieldPath: step.Name + "." + name,
				Message: fmt.Sprintf(
					"required field %q is empty",
					name,
				),
				FixSuggestion: fmt.Sprintf(
					"Provide a value for the %q "+
						"field in the %q section",
					name, step.Name,
				),
			})
		}
		return errs, nil
	}

	// Mark step as completed and advance.
	e.template.Steps[e.currentStep].Completed = true
	e.currentStep++
	if e.currentStep >= len(e.template.Steps) {
		e.isComplete = true
	}
	return nil, nil
}

// GetSuggestions returns suggested values for the named
// field based on the template's example values and the
// user's role context.
func (e *AuthoringEngine) GetSuggestions(
	fieldName string,
) []string {
	if e.isComplete {
		return nil
	}
	step := e.template.Steps[e.currentStep]
	var suggestions []string
	for _, f := range step.Fields {
		if f.Name == fieldName && f.ExampleValue != "" {
			suggestions = append(
				suggestions, f.ExampleValue,
			)
		}
	}
	return suggestions
}

// Progress returns the number of completed steps and the
// total number of steps.
func (e *AuthoringEngine) Progress() (int, int) {
	completed := 0
	for _, step := range e.template.Steps {
		if step.Completed {
			completed++
		}
	}
	return completed, len(e.template.Steps)
}

// IsComplete returns true if all authoring steps have been
// completed.
func (e *AuthoringEngine) IsComplete() bool {
	return e.isComplete
}

// BuildArtifact assembles the completed sections into an
// AuthoredArtifact with the correct metadata.
func (e *AuthoringEngine) BuildArtifact() *AuthoredArtifact {
	schemaDef, _ := ArtifactTypeToSchema(
		e.template.ArtifactType,
	)
	artifact := NewAuthoredArtifact(
		e.template.ArtifactType,
		schemaDef,
		"", // Schema version set by caller.
		e.roleName,
	)
	for i, step := range e.template.Steps {
		artifact.AddSection(step.Name)
		for fieldName, value := range e.fieldValues[i] {
			_ = artifact.SetFieldValue(
				step.Name, fieldName, value,
			)
		}
	}
	return artifact
}

// personalizeExplanation generates a role-specific
// explanation for the given step.
func personalizeExplanation(
	step *AuthoringStep,
	roleName string,
) string {
	if roleName == "" {
		return step.Description
	}
	base := step.Description
	if step.RoleExplanation != "" {
		base = step.RoleExplanation
	}
	if !strings.Contains(base, roleName) {
		return fmt.Sprintf(
			"As a %s, %s",
			roleName,
			lowerFirst(base),
		)
	}
	return base
}

// lowerFirst converts the first character of s to
// lowercase.
func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// ArtifactTemplates returns the predefined authoring
// templates for all supported artifact types.
func ArtifactTemplates() map[string]ArtifactTemplate {
	return map[string]ArtifactTemplate{
		consts.ArtifactGuidanceCatalog: {
			ArtifactType: consts.ArtifactGuidanceCatalog,
			Layer:        consts.LayerVectorsGuidance,
			Steps: []AuthoringStep{
				metadataStep(
					consts.ArtifactGuidanceCatalog,
				),
				scopeStep(
					consts.ArtifactGuidanceCatalog,
				),
				guidanceItemsStep(),
			},
		},
		consts.ArtifactVectorCatalog: {
			ArtifactType: consts.ArtifactVectorCatalog,
			Layer:        consts.LayerVectorsGuidance,
			Steps: []AuthoringStep{
				metadataStep(
					consts.ArtifactVectorCatalog,
				),
				scopeStep(
					consts.ArtifactVectorCatalog,
				),
				vectorsStep(),
			},
		},
		consts.ArtifactPrincipleCatalog: {
			ArtifactType: consts.ArtifactPrincipleCatalog,
			Layer:        consts.LayerVectorsGuidance,
			Steps: []AuthoringStep{
				metadataStep(
					consts.ArtifactPrincipleCatalog,
				),
				scopeStep(
					consts.ArtifactPrincipleCatalog,
				),
				principlesStep(),
			},
		},
		consts.ArtifactControlCatalog: {
			ArtifactType: consts.ArtifactControlCatalog,
			Layer:        consts.LayerThreatsControls,
			Steps: []AuthoringStep{
				metadataStep(
					consts.ArtifactControlCatalog,
				),
				scopeStep(
					consts.ArtifactControlCatalog,
				),
				controlsStep(),
			},
		},
		consts.ArtifactThreatCatalog: {
			ArtifactType: consts.ArtifactThreatCatalog,
			Layer:        consts.LayerThreatsControls,
			Steps: []AuthoringStep{
				metadataStep(
					consts.ArtifactThreatCatalog,
				),
				scopeStep(
					consts.ArtifactThreatCatalog,
				),
				capabilitiesStep(),
				threatsStep(),
			},
		},
		consts.ArtifactCapabilityCatalog: {
			ArtifactType: consts.ArtifactCapabilityCatalog,
			Layer:        consts.LayerThreatsControls,
			Steps: []AuthoringStep{
				metadataStep(
					consts.ArtifactCapabilityCatalog,
				),
				scopeStep(
					consts.ArtifactCapabilityCatalog,
				),
				capabilitiesStep(),
			},
		},
		consts.ArtifactPolicy: {
			ArtifactType: consts.ArtifactPolicy,
			Layer:        consts.LayerRiskPolicy,
			Steps: []AuthoringStep{
				metadataStep(consts.ArtifactPolicy),
				scopeStep(consts.ArtifactPolicy),
				policyCriteriaStep(),
			},
		},
		consts.ArtifactRiskCatalog: {
			ArtifactType: consts.ArtifactRiskCatalog,
			Layer:        consts.LayerRiskPolicy,
			Steps: []AuthoringStep{
				metadataStep(
					consts.ArtifactRiskCatalog,
				),
				scopeStep(
					consts.ArtifactRiskCatalog,
				),
				risksStep(),
			},
		},
		consts.ArtifactMappingDocument: {
			ArtifactType: consts.ArtifactMappingDocument,
			Layer:        consts.LayerVectorsGuidance,
			Steps: []AuthoringStep{
				metadataStep(
					consts.ArtifactMappingDocument,
				),
				mappingsStep(),
			},
		},
		consts.ArtifactEvaluationLog: {
			ArtifactType: consts.ArtifactEvaluationLog,
			Layer:        consts.LayerEvaluation,
			Steps: []AuthoringStep{
				metadataStep(
					consts.ArtifactEvaluationLog,
				),
				scopeStep(
					consts.ArtifactEvaluationLog,
				),
				evaluationsStep(),
			},
		},
		consts.ArtifactEnforcementLog: {
			ArtifactType: consts.ArtifactEnforcementLog,
			Layer:        consts.LayerEnforcement,
			Steps: []AuthoringStep{
				metadataStep(
					consts.ArtifactEnforcementLog,
				),
				scopeStep(
					consts.ArtifactEnforcementLog,
				),
				actionsStep(),
			},
		},
		consts.ArtifactAuditLog: {
			ArtifactType: consts.ArtifactAuditLog,
			Layer:        consts.LayerAudit,
			Steps: []AuthoringStep{
				metadataStep(
					consts.ArtifactAuditLog,
				),
				scopeStep(
					consts.ArtifactAuditLog,
				),
				auditResultsStep(),
			},
		},
	}
}

// metadataStep returns the standard metadata authoring
// step for the given artifact type.
func metadataStep(artifactType string) AuthoringStep {
	return AuthoringStep{
		Name: consts.SectionMetadata,
		Description: "Define the artifact metadata " +
			"including name, description, and version",
		Fields: []StepField{
			{
				Name: "name",
				Description: "Unique identifier for " +
					"the artifact following Gemara " +
					"naming conventions",
				FieldType: "string",
				Required:  true,
				ExampleValue: exampleName(
					artifactType,
				),
				HelpText: "Use the format " +
					"ORG.PROJECT.COMPONENT.TYPE##",
			},
			{
				Name: "description",
				Description: "Human-readable " +
					"description of what this " +
					"artifact covers",
				FieldType:    "string",
				Required:     true,
				ExampleValue: exampleDesc(artifactType),
			},
			{
				Name:         "version",
				Description:  "Artifact version",
				FieldType:    "string",
				Required:     false,
				ExampleValue: "1.0.0",
			},
		},
	}
}

// scopeStep returns the standard scope authoring step.
func scopeStep(artifactType string) AuthoringStep {
	return AuthoringStep{
		Name: consts.SectionScope,
		Description: "Define the scope and boundaries " +
			"of this artifact",
		Fields: []StepField{
			{
				Name: "scope",
				Description: "What this artifact " +
					"covers — the systems, processes, " +
					"or domains in scope",
				FieldType: "string",
				Required:  true,
				ExampleValue: exampleScope(
					artifactType,
				),
			},
			{
				Name: "boundary",
				Description: "What is explicitly " +
					"excluded from scope",
				FieldType:    "string",
				Required:     false,
				ExampleValue: "Third-party SaaS integrations",
			},
		},
	}
}

// guidanceItemsStep returns the guidance items step.
func guidanceItemsStep() AuthoringStep {
	return AuthoringStep{
		Name: consts.SectionGuidanceItems,
		Description: "Define the guidance items — " +
			"reusable best practices, standards, or " +
			"requirements",
		Fields: []StepField{
			{
				Name: "item_id",
				Description: "Guidance item " +
					"identifier",
				FieldType:    "string",
				Required:     true,
				ExampleValue: "GDN-001",
			},
			{
				Name: "item_description",
				Description: "What this guidance " +
					"item requires or recommends",
				FieldType: "string",
				Required:  true,
				ExampleValue: "All authentication " +
					"mechanisms must support " +
					"multi-factor authentication",
			},
		},
	}
}

// controlsStep returns the controls definition step.
func controlsStep() AuthoringStep {
	return AuthoringStep{
		Name: consts.SectionControls,
		Description: "Define the security controls " +
			"that mitigate identified threats",
		Fields: []StepField{
			{
				Name:         "control_id",
				Description:  "Control identifier",
				FieldType:    "string",
				Required:     true,
				ExampleValue: "CTL-001",
			},
			{
				Name: "control_description",
				Description: "What this control " +
					"implements or enforces",
				FieldType: "string",
				Required:  true,
				ExampleValue: "Enforce input " +
					"validation on all user-supplied " +
					"data",
			},
		},
	}
}

// capabilitiesStep returns the capabilities definition
// step for threat catalogs.
func capabilitiesStep() AuthoringStep {
	return AuthoringStep{
		Name: consts.SectionCapabilities,
		Description: "Define the capabilities — " +
			"functional areas that threats target",
		Fields: []StepField{
			{
				Name: "capability_name",
				Description: "Name of the capability " +
					"or functional area",
				FieldType:    "string",
				Required:     true,
				ExampleValue: "Authentication",
			},
			{
				Name: "capability_description",
				Description: "What this capability " +
					"provides to the system",
				FieldType: "string",
				Required:  true,
				ExampleValue: "User identity " +
					"verification and session " +
					"management",
			},
		},
	}
}

// threatsStep returns the threats definition step.
func threatsStep() AuthoringStep {
	return AuthoringStep{
		Name: consts.SectionThreats,
		Description: "Define the threats that target " +
			"the identified capabilities",
		Fields: []StepField{
			{
				Name: "threat_id",
				Description: "Threat identifier " +
					"following Gemara naming conventions",
				FieldType:    "string",
				Required:     true,
				ExampleValue: "THR-001",
			},
			{
				Name: "threat_description",
				Description: "Description of the " +
					"threat scenario",
				FieldType: "string",
				Required:  true,
				ExampleValue: "SQL injection via " +
					"unvalidated user input in search " +
					"parameters",
			},
			{
				Name: "target_capability",
				Description: "The capability this " +
					"threat targets",
				FieldType:    "string",
				Required:     true,
				ExampleValue: "Authentication",
			},
		},
	}
}

// policyCriteriaStep returns the policy criteria step.
func policyCriteriaStep() AuthoringStep {
	return AuthoringStep{
		Name: consts.SectionPolicyCriteria,
		Description: "Define the policy criteria — " +
			"rules and requirements that govern " +
			"compliance",
		Fields: []StepField{
			{
				Name: "criterion_id",
				Description: "Policy criterion " +
					"identifier",
				FieldType:    "string",
				Required:     true,
				ExampleValue: "POL-001",
			},
			{
				Name: "criterion_description",
				Description: "What this policy " +
					"criterion requires",
				FieldType: "string",
				Required:  true,
				ExampleValue: "All production " +
					"systems must pass security " +
					"review before deployment",
			},
			{
				Name: "adherence_timeline",
				Description: "Timeline for achieving " +
					"adherence to this criterion",
				FieldType:    "string",
				Required:     false,
				ExampleValue: "90 days from policy adoption",
			},
		},
	}
}

// mappingsStep returns the mappings definition step.
func mappingsStep() AuthoringStep {
	return AuthoringStep{
		Name: consts.SectionMappings,
		Description: "Define the mappings between " +
			"source and target artifacts",
		Fields: []StepField{
			{
				Name: "source_ref",
				Description: "Reference to the " +
					"source artifact or item",
				FieldType:    "string",
				Required:     true,
				ExampleValue: "GDN-001",
			},
			{
				Name: "target_ref",
				Description: "Reference to the " +
					"target artifact or item",
				FieldType:    "string",
				Required:     true,
				ExampleValue: "CTL-001",
			},
			{
				Name: "relationship",
				Description: "Type of relationship " +
					"(implements, equivalent, subsumes)",
				FieldType:    "enum",
				Required:     true,
				ExampleValue: consts.RelImplements,
			},
		},
	}
}

// evaluationsStep returns the evaluations definition step.
func evaluationsStep() AuthoringStep {
	return AuthoringStep{
		Name: consts.SectionEvaluations,
		Description: "Record evaluation results " +
			"against policy criteria",
		Fields: []StepField{
			{
				Name: "evaluation_id",
				Description: "Evaluation entry " +
					"identifier",
				FieldType:    "string",
				Required:     true,
				ExampleValue: "EVAL-001",
			},
			{
				Name: "criterion_ref",
				Description: "Reference to the " +
					"policy criterion being evaluated",
				FieldType:    "string",
				Required:     true,
				ExampleValue: "POL-001",
			},
			{
				Name: "result",
				Description: "Evaluation result " +
					"(pass, fail, partial, not_assessed)",
				FieldType:    "enum",
				Required:     true,
				ExampleValue: "pass",
			},
			{
				Name: "evidence",
				Description: "Evidence supporting " +
					"the evaluation result",
				FieldType: "string",
				Required:  false,
				ExampleValue: "Automated scan report " +
					"from 2026-03-01",
			},
		},
	}
}

// vectorsStep returns the vectors definition step.
func vectorsStep() AuthoringStep {
	return AuthoringStep{
		Name: consts.SectionVectors,
		Description: "Define the attack vectors — " +
			"methods and pathways through which " +
			"threats may be realized",
		Fields: []StepField{
			{
				Name:         "vector_id",
				Description:  "Vector identifier",
				FieldType:    "string",
				Required:     true,
				ExampleValue: "VEC-001",
			},
			{
				Name: "vector_description",
				Description: "Description of the " +
					"attack vector or technique",
				FieldType: "string",
				Required:  true,
				ExampleValue: "Credential stuffing " +
					"via compromised password databases",
			},
			{
				Name: "vector_group",
				Description: "Group this vector " +
					"belongs to",
				FieldType:    "string",
				Required:     true,
				ExampleValue: "Credential Attacks",
			},
		},
	}
}

// principlesStep returns the principles definition step.
func principlesStep() AuthoringStep {
	return AuthoringStep{
		Name: consts.SectionPrinciples,
		Description: "Define the principles — " +
			"foundational values that guide " +
			"governance, design, and operational " +
			"decisions",
		Fields: []StepField{
			{
				Name:         "principle_id",
				Description:  "Principle identifier",
				FieldType:    "string",
				Required:     true,
				ExampleValue: "PRC-001",
			},
			{
				Name: "principle_title",
				Description: "Title describing the " +
					"principle at a glance",
				FieldType:    "string",
				Required:     true,
				ExampleValue: "Defense in Depth",
			},
			{
				Name: "principle_description",
				Description: "Explanation of the " +
					"principle and its expected outcomes",
				FieldType: "string",
				Required:  true,
				ExampleValue: "Apply multiple layers " +
					"of security controls to protect " +
					"against single points of failure",
			},
			{
				Name: "principle_group",
				Description: "Group this principle " +
					"belongs to",
				FieldType:    "string",
				Required:     true,
				ExampleValue: "Secure Design",
			},
		},
	}
}

// risksStep returns the risks definition step.
func risksStep() AuthoringStep {
	return AuthoringStep{
		Name: consts.SectionRisks,
		Description: "Define the risks — potential " +
			"negative impacts resulting from threats",
		Fields: []StepField{
			{
				Name:         "risk_id",
				Description:  "Risk identifier",
				FieldType:    "string",
				Required:     true,
				ExampleValue: "RSK-001",
			},
			{
				Name: "risk_description",
				Description: "Description of the " +
					"risk scenario",
				FieldType: "string",
				Required:  true,
				ExampleValue: "Unauthorized access " +
					"to production systems via " +
					"compromised credentials",
			},
			{
				Name: "risk_severity",
				Description: "Assessed severity " +
					"level (Low, Medium, High, " +
					"Critical)",
				FieldType:    "enum",
				Required:     true,
				ExampleValue: "High",
			},
			{
				Name: "risk_group",
				Description: "Risk category this " +
					"risk belongs to",
				FieldType:    "string",
				Required:     true,
				ExampleValue: "Access Control",
			},
		},
	}
}

// actionsStep returns the enforcement actions step.
func actionsStep() AuthoringStep {
	return AuthoringStep{
		Name: consts.SectionActions,
		Description: "Record enforcement actions " +
			"taken in response to noncompliance " +
			"findings",
		Fields: []StepField{
			{
				Name: "disposition",
				Description: "Enforcement outcome " +
					"(Undetermined, Enforced, " +
					"Tolerated, Clear)",
				FieldType:    "enum",
				Required:     true,
				ExampleValue: "Enforced",
			},
			{
				Name: "method_ref",
				Description: "Reference to the " +
					"enforcement method in the Policy",
				FieldType:    "string",
				Required:     true,
				ExampleValue: "ACME.SEC.POL01.ENF01",
			},
			{
				Name: "message",
				Description: "Additional context " +
					"about the enforcement action",
				FieldType: "string",
				Required:  false,
				ExampleValue: "Deployment blocked " +
					"by admission controller",
			},
		},
	}
}

// auditResultsStep returns the audit results step.
func auditResultsStep() AuthoringStep {
	return AuthoringStep{
		Name: consts.SectionResults,
		Description: "Record audit results against " +
			"the defined criteria",
		Fields: []StepField{
			{
				Name:         "result_id",
				Description:  "Audit result identifier",
				FieldType:    "string",
				Required:     true,
				ExampleValue: "AUD-001",
			},
			{
				Name: "result_type",
				Description: "Type of audit result " +
					"(Gap, Finding, Observation, " +
					"Strength)",
				FieldType:    "enum",
				Required:     true,
				ExampleValue: "Finding",
			},
			{
				Name: "result_description",
				Description: "Detailed explanation " +
					"of the audit result",
				FieldType: "string",
				Required:  true,
				ExampleValue: "TLS 1.2 still in use " +
					"for internal service " +
					"communication",
			},
		},
	}
}

// exampleName returns an example artifact name for the
// given artifact type.
func exampleName(artifactType string) string {
	switch artifactType {
	case consts.ArtifactGuidanceCatalog:
		return "ACME.WEB.GDN01"
	case consts.ArtifactVectorCatalog:
		return "ACME.WEB.VEC01"
	case consts.ArtifactPrincipleCatalog:
		return "ACME.WEB.PRC01"
	case consts.ArtifactControlCatalog:
		return "ACME.WEB.CTL01"
	case consts.ArtifactThreatCatalog:
		return "ACME.WEB.THR01"
	case consts.ArtifactCapabilityCatalog:
		return "ACME.WEB.CAP01"
	case consts.ArtifactPolicy:
		return "ACME.WEB.POL01"
	case consts.ArtifactRiskCatalog:
		return "ACME.WEB.RSK01"
	case consts.ArtifactMappingDocument:
		return "ACME.WEB.MAP01"
	case consts.ArtifactEvaluationLog:
		return "ACME.WEB.EVAL01"
	case consts.ArtifactEnforcementLog:
		return "ACME.WEB.ENF01"
	case consts.ArtifactAuditLog:
		return "ACME.WEB.AUD01"
	default:
		return "ACME.PROJ.ART01"
	}
}

// exampleDesc returns an example description for the given
// artifact type.
func exampleDesc(artifactType string) string {
	switch artifactType {
	case consts.ArtifactGuidanceCatalog:
		return "Guidance catalog for web application " +
			"security best practices"
	case consts.ArtifactVectorCatalog:
		return "Attack vectors targeting web " +
			"application infrastructure"
	case consts.ArtifactPrincipleCatalog:
		return "Secure design principles for web " +
			"application architecture"
	case consts.ArtifactControlCatalog:
		return "Control catalog for web application " +
			"security controls"
	case consts.ArtifactThreatCatalog:
		return "Threat catalog for web application " +
			"threat assessment"
	case consts.ArtifactCapabilityCatalog:
		return "System capabilities for web " +
			"application platform"
	case consts.ArtifactPolicy:
		return "Security policy for web application " +
			"deployment"
	case consts.ArtifactRiskCatalog:
		return "Organizational risks for web " +
			"application operations"
	case consts.ArtifactMappingDocument:
		return "Mapping between guidance items and " +
			"implementing controls"
	case consts.ArtifactEvaluationLog:
		return "Evaluation log for policy compliance " +
			"assessment"
	case consts.ArtifactEnforcementLog:
		return "Enforcement actions for noncompliance " +
			"findings"
	case consts.ArtifactAuditLog:
		return "Audit results for compliance posture " +
			"review"
	default:
		return "Gemara artifact"
	}
}

// exampleScope returns an example scope for the given
// artifact type.
func exampleScope(artifactType string) string {
	switch artifactType {
	case consts.ArtifactGuidanceCatalog:
		return "Web application authentication and " +
			"authorization"
	case consts.ArtifactVectorCatalog:
		return "Web application attack techniques " +
			"and exploitation pathways"
	case consts.ArtifactPrincipleCatalog:
		return "Secure design principles for cloud " +
			"native applications"
	case consts.ArtifactControlCatalog:
		return "Web application input validation and " +
			"output encoding"
	case consts.ArtifactThreatCatalog:
		return "Web application attack surface"
	case consts.ArtifactCapabilityCatalog:
		return "Kubernetes platform features and " +
			"components"
	case consts.ArtifactPolicy:
		return "Production deployment pipeline"
	case consts.ArtifactRiskCatalog:
		return "Organizational risk categories for " +
			"cloud operations"
	case consts.ArtifactEvaluationLog:
		return "Q1 2026 compliance assessment"
	case consts.ArtifactEnforcementLog:
		return "Enforcement actions for Q1 2026 " +
			"evaluation findings"
	case consts.ArtifactAuditLog:
		return "Annual compliance audit for cloud " +
			"infrastructure"
	default:
		return "System under assessment"
	}
}
