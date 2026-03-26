// SPDX-License-Identifier: Apache-2.0

import { useState, useCallback } from 'react';
import { journeyData } from '../../generated/journey-data';
import { Sidebar } from './sidebar/Sidebar';
import { PlaygroundEditor } from './PlaygroundEditor';
import { EditorToolbar } from './EditorToolbar';
import {
  ValidationResults,
  type ValidationResult,
} from './ValidationResults';

/** Schema definitions for cue vet, keyed by artifact type. */
const SCHEMA_DEFS: Record<string, string> = {
  ControlCatalog: '#ControlCatalog',
  GuidanceCatalog: '#GuidanceCatalog',
  ThreatCatalog: '#ThreatCatalog',
  RiskCatalog: '#RiskCatalog',
  Policy: '#Policy',
};

interface PlaygroundLayoutProps {
  initialType: string;
}

/**
 * PlaygroundLayout is the top-level playground component
 * with sidebar + editor layout and CUE validation.
 */
export function PlaygroundLayout({
  initialType,
}: PlaygroundLayoutProps) {
  const [activeType, setActiveType] = useState(initialType);
  const [editorContent, setEditorContent] = useState<string>(
    () =>
      (journeyData.playgroundExamples as Record<string, string>)[
        initialType
      ] ?? '',
  );
  const [hasEdited, setHasEdited] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] =
    useState(false);
  const [isValidating, setIsValidating] = useState(false);
  const [validationResult, setValidationResult] =
    useState<ValidationResult | null>(null);

  const handleEditorChange = useCallback((value: string) => {
    setEditorContent(value);
    setHasEdited(true);
  }, []);

  const handleSelectType = useCallback(
    (type: string) => {
      if (type === activeType) return;

      if (hasEdited) {
        const confirmed = window.confirm(
          'You have unsaved changes. Switch artifact ' +
            'type and load the example? Your current ' +
            'edits will be lost.',
        );
        if (!confirmed) return;
      }

      const example =
        (journeyData.playgroundExamples as Record<string, string>)[
          type
        ] ?? '';

      setActiveType(type);
      setEditorContent(example);
      setHasEdited(false);
      setValidationResult(null);
    },
    [activeType, hasEdited],
  );

  const handleValidate = useCallback(async () => {
    const definition = SCHEMA_DEFS[activeType];
    if (!definition) return;

    setIsValidating(true);
    setValidationResult(null);

    try {
      const resp = await fetch('/api/validate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          content: editorContent,
          definition,
        }),
      });

      if (resp.ok) {
        const data = await resp.json();
        setValidationResult({
          valid: data.valid,
          errors: data.errors ?? [],
        });
      } else {
        const data = await resp.json().catch(() => null);
        setValidationResult({
          valid: false,
          errors: [
            data?.error ??
              `Server returned HTTP ${resp.status}`,
          ],
        });
      }
    } catch (err) {
      setValidationResult({
        valid: false,
        errors: [
          err instanceof Error
            ? err.message
            : 'Validation request failed',
        ],
      });
    } finally {
      setIsValidating(false);
    }
  }, [activeType, editorContent]);

  return (
    <div className="pg-layout">
      <div className="pg-main">
        {/* Sidebar */}
        <Sidebar
          activeType={activeType}
          collapsed={sidebarCollapsed}
          onSelectType={handleSelectType}
          onToggleCollapse={() =>
            setSidebarCollapsed((c) => !c)
          }
        />

        {/* Editor area */}
        <div className="pg-editor-area">
          <EditorToolbar
            activeType={activeType}
            editorContent={editorContent}
            onValidate={handleValidate}
            isValidating={isValidating}
          />

          <div className="pg-editor-wrapper">
            <PlaygroundEditor
              value={editorContent}
              onChange={handleEditorChange}
            />
          </div>

          <ValidationResults
            result={validationResult}
            onDismiss={() => setValidationResult(null)}
          />
        </div>
      </div>
    </div>
  );
}
