// SPDX-License-Identifier: Apache-2.0

import { Link } from 'react-router-dom';
import { journeyData } from '../generated/journey-data';
import type { UpstreamTutorial } from '../generated/journey-data';
import type { ActivityProfile } from '../lib/roles';

interface TutorialSuggestionsProps {
  profile: ActivityProfile;
  onBack: () => void;
  onGoToMCP: () => void;
}

interface ScoredTutorial {
  tutorial: UpstreamTutorial;
  score: number;
  reasons: string[];
  isPrerequisite: boolean;
}

/**
 * rankTutorials scores each upstream tutorial based on
 * the user's resolved layers, role, and activity keywords.
 * Returns tutorials sorted by score descending, with
 * prerequisites promoted.
 */
function rankTutorials(
  profile: ActivityProfile,
): ScoredTutorial[] {
  const activeLayers = new Set(
    profile.resolvedLayers.map((l) => l.layer),
  );
  const strongLayers = new Set(
    profile.resolvedLayers
      .filter((l) => l.confidence === 'strong')
      .map((l) => l.layer),
  );
  const roleName = profile.role?.name ?? '';
  const keywords = new Set(
    profile.extractedKeywords.map((k) => k.toLowerCase()),
  );

  const scored: ScoredTutorial[] = [];

  for (const tutorial of journeyData.upstreamTutorials) {
    let score = 0;
    const reasons: string[] = [];

    // Layer match: +3 for strong match, +2 for inferred.
    if (activeLayers.has(tutorial.layer)) {
      if (strongLayers.has(tutorial.layer)) {
        score += 3;
        const layerInfo = journeyData.layers.find(
          (l) => l.number === tutorial.layer,
        );
        reasons.push(
          `Matches your Layer ${tutorial.layer} (${layerInfo?.name ?? ''}) activities`,
        );
      } else {
        score += 2;
        reasons.push(
          `Relevant to your role's Layer ${tutorial.layer} focus`,
        );
      }
    }

    // Role match: +2 if the user's role is in the
    // tutorial's target roles.
    if (
      roleName &&
      (tutorial.roles as readonly string[]).includes(roleName)
    ) {
      score += 2;
      reasons.push(`Recommended for ${roleName}`);
    }

    // Goal/keyword overlap: +1 per matching keyword
    // found in the tutorial's goals.
    for (const goal of tutorial.goals) {
      const goalLower = goal.toLowerCase();
      for (const kw of keywords) {
        if (goalLower.includes(kw)) {
          score += 1;
          break; // Only count each goal once.
        }
      }
    }

    // Artifact type overlap: +1 if the tutorial produces
    // an artifact type the user was recommended.
    for (const artType of tutorial.artifactTypes) {
      if (
        profile.recommendations.some(
          (r) => r.artifactType === artType,
        )
      ) {
        score += 1;
        reasons.push(
          `Teaches you to create ${artType} artifacts`,
        );
        break;
      }
    }

    if (score > 0) {
      scored.push({
        tutorial,
        score,
        reasons,
        isPrerequisite: false,
      });
    }
  }

  // Mark prerequisites: if a scored tutorial lists a
  // prerequisite, ensure the prerequisite is also
  // included and marked.
  const scoredIds = new Set(scored.map((s) => s.tutorial.id));
  for (const s of scored) {
    for (const preId of s.tutorial.prerequisites) {
      if (!scoredIds.has(preId)) {
        const preTutorial = journeyData.upstreamTutorials.find(
          (t) => t.id === preId,
        );
        if (preTutorial) {
          scored.push({
            tutorial: preTutorial,
            score: s.score - 0.5, // Just below the dependent.
            reasons: [
              `Prerequisite for ${s.tutorial.title}`,
            ],
            isPrerequisite: true,
          });
          scoredIds.add(preId);
        }
      }
    }
  }

  // Sort by score descending, then by layer ascending
  // for tie-breaking (lower layers first = learn
  // foundations first).
  scored.sort((a, b) => {
    if (b.score !== a.score) return b.score - a.score;
    return a.tutorial.layer - b.tutorial.layer;
  });

  return scored;
}

export function TutorialSuggestions({
  profile,
  onBack,
  onGoToMCP,
}: TutorialSuggestionsProps) {
  const scoredTutorials = rankTutorials(profile);

  // Split into primary (top matches) and additional.
  const primary = scoredTutorials.filter(
    (s) => s.score >= 3,
  );
  const additional = scoredTutorials.filter(
    (s) => s.score > 0 && s.score < 3,
  );

  return (
    <div className="card">
      <h2>Recommended Tutorials</h2>
      <p>
        Based on your role
        {profile.role ? ` (${profile.role.name})` : ''} and
        activities, complete these upstream Gemara tutorials
        before using the MCP server. The tutorials teach the
        concepts and artifact structures you will need.
      </p>

      {/* Learning path note */}
      <div className="tutorial-path-note">
        <strong>Suggested learning path:</strong> Work
        through the tutorials in order. Each tutorial
        builds on concepts from the previous one. After
        completing them, use the MCP server to create your
        own artifacts with guided wizards.
      </div>

      {/* Primary tutorials */}
      {primary.length > 0 && (
        <>
          <h3
            style={{ marginTop: '24px', marginBottom: '8px' }}
          >
            Best Fit for Your Journey
          </h3>
          <div className="tutorial-list">
            {primary.map((s, i) => (
              <TutorialCard
                key={s.tutorial.id}
                scored={s}
                index={i + 1}
                totalPrimary={primary.length}
              />
            ))}
          </div>
        </>
      )}

      {/* Additional tutorials */}
      {additional.length > 0 && (
        <>
          <h3
            style={{ marginTop: '24px', marginBottom: '8px' }}
          >
            Also Relevant
          </h3>
          <div className="tutorial-list">
            {additional.map((s) => (
              <TutorialCard
                key={s.tutorial.id}
                scored={s}
                index={0}
                totalPrimary={0}
              />
            ))}
          </div>
        </>
      )}

      {/* No tutorials matched */}
      {scoredTutorials.length === 0 && (
        <div
          className="tutorial-path-note"
          style={{ marginTop: '24px' }}
        >
          No upstream tutorials match your current profile.
          You can browse all available tutorials at{' '}
          <a
            href="https://gemara.openssf.org/tutorials/"
            target="_blank"
            rel="noopener noreferrer"
          >
            gemara.openssf.org/tutorials
          </a>
          .
        </div>
      )}

      {/* All tutorials link */}
      <div
        className="tutorial-all-link"
        style={{ marginTop: '20px' }}
      >
        Browse all tutorials at{' '}
        <a
          href="https://gemara.openssf.org/tutorials/"
          target="_blank"
          rel="noopener noreferrer"
        >
          gemara.openssf.org/tutorials
        </a>
      </div>

      <div className="actions">
        <button className="btn btn-secondary" onClick={onBack}>
          Back
        </button>
        <button className="btn btn-primary" onClick={onGoToMCP}>
          MCP Server Setup &rarr;
        </button>
      </div>
    </div>
  );
}

interface TutorialCardProps {
  scored: ScoredTutorial;
  index: number;
  totalPrimary: number;
}

function TutorialCard({
  scored,
  index,
  totalPrimary,
}: TutorialCardProps) {
  const { tutorial, reasons, isPrerequisite } = scored;
  const layerInfo = journeyData.layers.find(
    (l) => l.number === tutorial.layer,
  );

  return (
    <div
      className={`tutorial-card ${isPrerequisite ? 'prerequisite' : ''}`}
    >
      <div className="tutorial-card-header">
        <div className="tutorial-card-title">
          {index > 0 && totalPrimary > 1 && (
            <span className="tutorial-step-number">
              {index}
            </span>
          )}
          <h4>{tutorial.title}</h4>
        </div>
        <div className="tutorial-card-badges">
          <span className="tutorial-layer-badge">
            Layer {tutorial.layer}
            {layerInfo ? ` — ${layerInfo.name}` : ''}
          </span>
          {isPrerequisite && (
            <span className="tutorial-prereq-badge">
              Prerequisite
            </span>
          )}
        </div>
      </div>

      <p className="tutorial-card-desc">
        {tutorial.description}
      </p>

      {/* Why this tutorial */}
      {reasons.length > 0 && (
        <div className="tutorial-reasons">
          {reasons.map((r, i) => (
            <span key={i} className="tutorial-reason">
              {r}
            </span>
          ))}
        </div>
      )}

      {/* What you'll produce */}
      {tutorial.artifactTypes.length > 0 && (
        <div className="tutorial-produces">
          <span className="tutorial-produces-label">
            You will learn to create:
          </span>
          {tutorial.artifactTypes.map((at) => (
            <span key={at} className="artifact-badge">
              {at}
            </span>
          ))}
        </div>
      )}

      {/* Prerequisites */}
      {tutorial.prerequisites.length > 0 && (
        <div className="tutorial-prereqs">
          <span className="tutorial-prereqs-label">
            Complete first:
          </span>
          {tutorial.prerequisites.map((preId) => {
            const pre = journeyData.upstreamTutorials.find(
              (t) => t.id === preId,
            );
            return (
              <span key={preId} className="tutorial-prereq-ref">
                {pre?.title ?? preId}
              </span>
            );
          })}
        </div>
      )}

      {/* Open tutorial link + playground */}
      <div className="tutorial-card-actions">
        <a
          href={tutorial.url}
          target="_blank"
          rel="noopener noreferrer"
          className="btn btn-tutorial"
        >
          Open Tutorial
        </a>
        <Link
          to={`/playground?type=${
            tutorial.primaryArtifactType || ''
          }`}
          className="btn btn-playground"
        >
          Open Playground
        </Link>
      </div>
    </div>
  );
}
