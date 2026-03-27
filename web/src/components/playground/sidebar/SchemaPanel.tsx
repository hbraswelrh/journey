// SPDX-License-Identifier: Apache-2.0

import { journeyData } from '../../../generated/journey-data';

interface SchemaPanelProps {
  activeType: string;
}

/**
 * SchemaPanel renders schema field documentation for the
 * active artifact type from the generated data.
 */
export function SchemaPanel({ activeType }: SchemaPanelProps) {
  const schemas = journeyData.playgroundSchemas as Record<
    string,
    ReadonlyArray<{
      readonly name: string;
      readonly type: string;
      readonly required: boolean;
      readonly description: string;
    }>
  >;
  const fields = schemas[activeType] ?? [];

  if (fields.length === 0) {
    return (
      <div className="pg-schema-panel">
        <h4 className="pg-panel-heading">Schema</h4>
        <p className="pg-panel-empty">
          No schema documentation available for this
          artifact type.
        </p>
      </div>
    );
  }

  return (
    <div className="pg-schema-panel">
      <h4 className="pg-panel-heading">Schema Fields</h4>
      <div className="pg-schema-list">
        {fields.map((field) => (
          <div key={field.name} className="pg-schema-field">
            <div className="pg-schema-field-header">
              <code className="pg-schema-field-name">
                {field.name}
              </code>
              <span className="pg-schema-field-type">
                {field.type}
              </span>
              {field.required && (
                <span className="pg-schema-required-badge">
                  required
                </span>
              )}
            </div>
            <p className="pg-schema-field-desc">
              {field.description}
            </p>
          </div>
        ))}
      </div>
    </div>
  );
}
