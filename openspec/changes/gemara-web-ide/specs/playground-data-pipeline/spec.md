## ADDED Requirements

### Requirement: Export example artifact YAML in genwebdata
The `cmd/genwebdata` pipeline SHALL fetch example artifact YAML files from the upstream Gemara test data and embed their content as string constants in the generated TypeScript data module.

#### Scenario: Example artifacts are generated
- **WHEN** `make web-data` is run
- **THEN** the generated `journey-data.ts` includes a `playgroundExamples` map containing YAML string content keyed by artifact type identifier (e.g., `ControlCatalog`, `GuidanceCatalog`, `ThreatCatalog`, `RiskCatalog`, `Policy`)

#### Scenario: Example content matches upstream test data
- **WHEN** the generated example for `ControlCatalog` is inspected
- **THEN** the content matches the YAML from `gemaraproj/gemara/test/test-data/good-ccc.yaml`

### Requirement: Export schema field documentation in genwebdata
The `cmd/genwebdata` pipeline SHALL generate schema field documentation for each supported artifact type, including field names, types, required status, and descriptions.

#### Scenario: Schema docs are generated
- **WHEN** `make web-data` is run
- **THEN** the generated data includes a `playgroundSchemas` map containing schema documentation objects keyed by artifact type identifier

#### Scenario: Schema doc includes field details
- **WHEN** the schema documentation for `ControlCatalog` is inspected
- **THEN** it includes entries for fields such as `metadata`, `title`, `groups`, and `controls`, each with a name, type description, required flag, and human-readable description

### Requirement: Export lexicon terms in genwebdata
The `cmd/genwebdata` pipeline SHALL export Gemara lexicon terms as structured data in the generated TypeScript module, including each term's name, definition, and which artifact types it is relevant to.

#### Scenario: Lexicon terms are generated
- **WHEN** `make web-data` is run
- **THEN** the generated data includes a `playgroundLexicon` array containing lexicon term objects with `term`, `definition`, and `artifactTypes` fields

### Requirement: Artifact type to example file mapping
The `cmd/genwebdata` pipeline SHALL use the following mapping from artifact type identifiers to upstream test data files:
- `ControlCatalog` -> `good-ccc.yaml`
- `GuidanceCatalog` -> `good-aigf.yaml`
- `ThreatCatalog` -> `good-threat-catalog.yaml`
- `RiskCatalog` -> `good-risk-catalog.yaml`
- `Policy` -> `good-policy.yaml`

#### Scenario: All five artifact types have examples
- **WHEN** the generated `playgroundExamples` map is inspected
- **THEN** it contains exactly five entries, one for each supported artifact type

#### Scenario: Missing upstream file causes build error
- **WHEN** an upstream test data file cannot be fetched during `make web-data`
- **THEN** the build fails with a clear error message identifying which file could not be retrieved

### Requirement: Schema definition source
The schema field documentation SHALL be derived from the Gemara CUE schema definitions, either by parsing the `gemara://schema/definitions` MCP resource at build time or by maintaining a curated Go constant in `internal/consts/` that mirrors the schema structure.

#### Scenario: Schema docs cover required metadata fields
- **WHEN** the schema documentation for any artifact type is inspected
- **THEN** it includes the `metadata` field documentation with sub-fields `id`, `type`, `gemara-version`, `version`, `description`, and `author`
