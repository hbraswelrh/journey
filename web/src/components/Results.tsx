// SPDX-License-Identifier: Apache-2.0

import { journeyData } from '../generated/journey-data';
import type { ActivityProfile } from '../lib/roles';

interface ResultsProps {
  profile: ActivityProfile;
  onBack: () => void;
  onGoToTutorials: () => void;
}

export function Results({ profile, onBack, onGoToTutorials }: ResultsProps) {
  const uniqueLayers = [
    ...new Set(profile.resolvedLayers.map((l) => l.layer)),
  ].sort((a, b) => a - b);

  return (
    <div className="card">
      <h2>Your Gemara Profile</h2>
      <p>
        Based on your role
        {profile.role ? ` (${profile.role.name})` : ''} and
        activities, here are your relevant Gemara layers and
        recommended artifacts.
      </p>

      {/* Layer Map */}
      <h3 style={{ marginTop: '24px', marginBottom: '8px' }}>
        Relevant Layers
      </h3>
      <div className="layer-map">
        {journeyData.layers.map((layer) => {
          const isActive = uniqueLayers.includes(layer.number);
          const mapping = profile.resolvedLayers.find(
            (l) => l.layer === layer.number,
          );
          return (
            <div
              key={layer.number}
              className={`layer-row ${isActive ? 'active' : ''}`}
            >
              <span className="layer-number">{layer.number}</span>
              <div className="layer-info">
                <h4>
                  {layer.name}
                  {mapping && (
                    <span
                      style={{
                        fontSize: '12px',
                        fontWeight: 400,
                        marginLeft: '8px',
                        color: 'var(--text)',
                      }}
                    >
                      ({mapping.confidence})
                    </span>
                  )}
                </h4>
                <p>{layer.purpose}</p>
              </div>
              {layer.artifactIds.length > 0 && isActive && (
                <div className="layer-artifacts">
                  {layer.artifactIds.map((id) => (
                    <span key={id} className="artifact-badge">
                      {id}
                    </span>
                  ))}
                </div>
              )}
            </div>
          );
        })}
      </div>

      {/* Layer Flows */}
      {uniqueLayers.length > 1 && (
        <>
          <h3 style={{ marginTop: '24px', marginBottom: '8px' }}>
            How Your Layers Connect
          </h3>
          <div className="flow-list">
            {journeyData.layerFlows
              .filter(
                (f) =>
                  uniqueLayers.includes(f.from) &&
                  uniqueLayers.includes(f.to),
              )
              .map((f, i) => {
                const fromName =
                  journeyData.layers.find((l) => l.number === f.from)
                    ?.name ?? '';
                const toName =
                  journeyData.layers.find((l) => l.number === f.to)
                    ?.name ?? '';
                return (
                  <div key={i} className="flow-item">
                    <strong>{fromName}</strong>
                    <span className="flow-arrow">&rarr;</span>
                    <strong>{toName}</strong>
                    <span style={{ color: 'var(--text)' }}>
                      &mdash; {f.description}
                    </span>
                  </div>
                );
              })}
          </div>
        </>
      )}

      {/* Artifact Recommendations */}
      {profile.recommendations.length > 0 && (
        <>
          <h3 style={{ marginTop: '24px', marginBottom: '8px' }}>
            Recommended Artifacts
          </h3>
          <div className="recommendations">
            {profile.recommendations.map((rec) => (
              <div key={rec.artifactType} className="rec-card">
                <div className="rec-header">
                  <h4>{rec.artifactType}</h4>
                  <span
                    className={`rec-approach ${rec.authoringApproach}`}
                  >
                    {rec.authoringApproach === 'wizard'
                      ? `Wizard: ${rec.mcpWizard}`
                      : 'Collaborative'}
                  </span>
                </div>
                <p>{rec.description}</p>
                <div className="rec-schema">
                  Schema: <code>{rec.schemaDef}</code>
                  &nbsp;&middot;&nbsp; Layer {rec.layer}
                  &nbsp;&middot;&nbsp; Confidence:{' '}
                  {rec.confidence}
                </div>
                {rec.checklist.length > 0 && (
                  <>
                    <h4
                      style={{
                        marginTop: '12px',
                        fontSize: '14px',
                      }}
                    >
                      Preparation Checklist
                    </h4>
                    <ul className="checklist">
                      {rec.checklist.map((item, i) => (
                        <li key={i}>{item}</li>
                      ))}
                    </ul>
                  </>
                )}
              </div>
            ))}
          </div>
        </>
      )}

      <div className="actions">
        <button className="btn btn-secondary" onClick={onBack}>
          Back
        </button>
        <button className="btn btn-primary" onClick={onGoToTutorials}>
          Recommended Tutorials &rarr;
        </button>
      </div>
    </div>
  );
}
