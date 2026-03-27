// SPDX-License-Identifier: Apache-2.0

import { journeyData } from '../../../generated/journey-data';

interface LexiconPanelProps {
  activeType: string;
}

/**
 * LexiconPanel renders Gemara lexicon terms, sorting
 * relevant terms first based on the active artifact type.
 */
export function LexiconPanel({
  activeType,
}: LexiconPanelProps) {
  // Sort terms: relevant to active type first.
  const sorted = [...journeyData.playgroundLexicon].sort(
    (a, b) => {
      const aRelevant = (a.artifactTypes as readonly string[]).includes(activeType);
      const bRelevant = (b.artifactTypes as readonly string[]).includes(activeType);
      if (aRelevant && !bRelevant) return -1;
      if (!aRelevant && bRelevant) return 1;
      return 0;
    },
  );

  return (
    <div className="pg-lexicon-panel">
      <h4 className="pg-panel-heading">Lexicon</h4>
      <div className="pg-lexicon-list">
        {sorted.map((term) => {
          const isRelevant =
            (term.artifactTypes as readonly string[]).includes(activeType);
          return (
            <div
              key={term.term}
              className={`pg-lexicon-term ${
                isRelevant ? 'relevant' : ''
              }`}
            >
              <dt className="pg-lexicon-name">
                {term.term}
              </dt>
              <dd className="pg-lexicon-def">
                {term.definition}
              </dd>
            </div>
          );
        })}
      </div>
    </div>
  );
}
