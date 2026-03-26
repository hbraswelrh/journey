// SPDX-License-Identifier: Apache-2.0

import { useSearchParams, Link } from 'react-router-dom';
import { journeyData } from '../../generated/journey-data';
import { PlaygroundLayout } from './PlaygroundLayout';
import '../../Playground.css';

/** Valid artifact type identifiers for the playground. */
const VALID_TYPES = new Set(
  Object.keys(
    journeyData.playgroundExamples as Record<string, string>,
  ),
);

/** Default artifact type when none or invalid is specified. */
const DEFAULT_TYPE = 'ControlCatalog';

/**
 * Playground reads the `type` query parameter and
 * pre-selects the corresponding artifact type. Invalid
 * or missing values fall back to ControlCatalog.
 */
function Playground() {
  const [searchParams] = useSearchParams();
  const typeParam = searchParams.get('type') ?? '';

  const initialType = VALID_TYPES.has(typeParam)
    ? typeParam
    : DEFAULT_TYPE;

  return (
    <div className="pg-page">
      <header className="pg-header">
        <h1>Gemara Playground</h1>
        <Link to="/" className="pg-back-link">
          &larr; Back to Journey
        </Link>
      </header>
      <PlaygroundLayout initialType={initialType} />
    </div>
  );
}

export default Playground;
