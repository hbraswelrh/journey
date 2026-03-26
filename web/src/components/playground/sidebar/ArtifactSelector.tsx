// SPDX-License-Identifier: Apache-2.0

import { journeyData } from '../../../generated/journey-data';

/** Ordered list of playground artifact type identifiers. */
const ARTIFACT_TYPES = [
  'ControlCatalog',
  'GuidanceCatalog',
  'ThreatCatalog',
  'RiskCatalog',
  'Policy',
] as const;

/** Human-readable labels for artifact types. */
const ARTIFACT_LABELS: Record<string, string> = {
  ControlCatalog: 'Control Catalog',
  GuidanceCatalog: 'Guidance Catalog',
  ThreatCatalog: 'Threat Catalog',
  RiskCatalog: 'Risk Catalog',
  Policy: 'Policy',
};

interface ArtifactSelectorProps {
  activeType: string;
  onSelect: (type: string) => void;
}

/**
 * ArtifactSelector lists the five artifact types with
 * click-to-select behavior and active state styling.
 */
export function ArtifactSelector({
  activeType,
  onSelect,
}: ArtifactSelectorProps) {
  return (
    <div className="pg-artifact-selector">
      <h4 className="pg-panel-heading">Artifact Types</h4>
      <div className="pg-artifact-list">
        {ARTIFACT_TYPES.map((type) => {
          const hasExample =
            type in (journeyData.playgroundExamples as Record<string, string>);
          return (
            <button
              key={type}
              className={`pg-artifact-item ${
                activeType === type ? 'active' : ''
              }`}
              onClick={() => onSelect(type)}
            >
              <span className="pg-artifact-label">
                {ARTIFACT_LABELS[type] ?? type}
              </span>
              {hasExample && (
                <span className="pg-artifact-example-badge">
                  example
                </span>
              )}
            </button>
          );
        })}
      </div>
    </div>
  );
}
