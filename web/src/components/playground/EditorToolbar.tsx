// SPDX-License-Identifier: Apache-2.0

import { useState } from 'react';

/** Human-readable labels for artifact types. */
const ARTIFACT_LABELS: Record<string, string> = {
  ControlCatalog: 'Control Catalog',
  GuidanceCatalog: 'Guidance Catalog',
  ThreatCatalog: 'Threat Catalog',
  RiskCatalog: 'Risk Catalog',
  Policy: 'Policy',
};

interface EditorToolbarProps {
  activeType: string;
  editorContent: string;
  onValidate: () => void;
  isValidating: boolean;
}

/**
 * EditorToolbar renders the artifact type label,
 * validate button, and copy-to-clipboard button.
 */
export function EditorToolbar({
  activeType,
  editorContent,
  onValidate,
  isValidating,
}: EditorToolbarProps) {
  const [copyStatus, setCopyStatus] = useState<
    'idle' | 'copied'
  >('idle');

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(editorContent);
      setCopyStatus('copied');
      setTimeout(() => setCopyStatus('idle'), 2000);
    } catch {
      // Fallback for older browsers.
      const textarea = document.createElement('textarea');
      textarea.value = editorContent;
      document.body.appendChild(textarea);
      textarea.select();
      document.execCommand('copy');
      document.body.removeChild(textarea);
      setCopyStatus('copied');
      setTimeout(() => setCopyStatus('idle'), 2000);
    }
  };

  return (
    <div className="pg-toolbar">
      <div className="pg-toolbar-left">
        <span className="pg-toolbar-type">
          {ARTIFACT_LABELS[activeType] ?? activeType}
        </span>
      </div>
      <div className="pg-toolbar-right">
        <button
          className="pg-toolbar-btn pg-validate-btn"
          onClick={onValidate}
          disabled={isValidating || !editorContent.trim()}
          title="Validate against Gemara CUE schema"
        >
          {isValidating ? 'Validating...' : 'Validate'}
        </button>
        <button
          className="pg-toolbar-btn pg-copy-btn"
          onClick={handleCopy}
        >
          {copyStatus === 'copied'
            ? 'Copied!'
            : 'Copy to Clipboard'}
        </button>
      </div>
    </div>
  );
}
