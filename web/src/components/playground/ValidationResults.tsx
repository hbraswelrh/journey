// SPDX-License-Identifier: Apache-2.0

export interface ValidationResult {
  valid: boolean;
  errors: string[];
}

interface ValidationResultsProps {
  result: ValidationResult | null;
  onDismiss: () => void;
}

/**
 * ValidationResults displays CUE schema validation
 * success or error messages below the editor.
 */
export function ValidationResults({
  result,
  onDismiss,
}: ValidationResultsProps) {
  if (!result) return null;

  return (
    <div
      className={`pg-validation-results ${
        result.valid
          ? 'pg-validation-success'
          : 'pg-validation-error'
      }`}
    >
      <div className="pg-validation-header">
        <span className="pg-validation-status">
          {result.valid
            ? 'Validation Passed'
            : 'Validation Failed'}
        </span>
        <button
          className="pg-validation-dismiss"
          onClick={onDismiss}
          title="Dismiss"
        >
          &times;
        </button>
      </div>
      {result.errors.length > 0 && (
        <ul className="pg-validation-errors">
          {result.errors.map((err, i) => (
            <li key={i}>{err}</li>
          ))}
        </ul>
      )}
    </div>
  );
}
