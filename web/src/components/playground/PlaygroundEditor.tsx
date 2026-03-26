// SPDX-License-Identifier: Apache-2.0

import { useRef, useEffect, useCallback } from 'react';
import { EditorState } from '@codemirror/state';
import { EditorView, keymap, lineNumbers, highlightActiveLine } from '@codemirror/view';
import { yaml } from '@codemirror/lang-yaml';
import { defaultKeymap, indentWithTab } from '@codemirror/commands';
import {
  syntaxHighlighting,
  defaultHighlightStyle,
  bracketMatching,
  indentOnInput,
} from '@codemirror/language';
import { oneDark } from '@codemirror/theme-one-dark';

interface PlaygroundEditorProps {
  value: string;
  onChange: (value: string) => void;
}

/**
 * useIsDarkMode returns true when the user's system
 * preference is dark mode.
 */
function useIsDarkMode(): boolean {
  const query =
    typeof window !== 'undefined'
      ? window.matchMedia('(prefers-color-scheme: dark)')
      : null;

  // Use a simple check; the editor is recreated when
  // the value prop changes so theme switches are handled
  // on next load.
  return query?.matches ?? false;
}

/**
 * PlaygroundEditor renders a CodeMirror 6 editor
 * configured for YAML editing with syntax highlighting,
 * line numbers, bracket matching, and auto-indentation.
 */
export function PlaygroundEditor({
  value,
  onChange,
}: PlaygroundEditorProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const viewRef = useRef<EditorView | null>(null);
  const onChangeRef = useRef(onChange);
  const isDark = useIsDarkMode();

  // Keep the callback ref up to date.
  onChangeRef.current = onChange;

  const createEditor = useCallback(() => {
    if (!containerRef.current) return;

    // Destroy previous instance if any.
    if (viewRef.current) {
      viewRef.current.destroy();
      viewRef.current = null;
    }

    const updateListener = EditorView.updateListener.of(
      (update) => {
        if (update.docChanged) {
          onChangeRef.current(
            update.state.doc.toString(),
          );
        }
      },
    );

    const extensions = [
      lineNumbers(),
      highlightActiveLine(),
      bracketMatching(),
      indentOnInput(),
      syntaxHighlighting(defaultHighlightStyle, {
        fallback: true,
      }),
      yaml(),
      keymap.of([...defaultKeymap, indentWithTab]),
      updateListener,
      EditorView.theme({
        '&': {
          height: '100%',
          fontSize: '14px',
        },
        '.cm-scroller': {
          overflow: 'auto',
          fontFamily: 'var(--mono)',
        },
        '.cm-content': {
          caretColor: 'var(--text-h)',
        },
        '.cm-gutters': {
          backgroundColor: 'var(--code-bg)',
          color: 'var(--text)',
          borderRight: '1px solid var(--border)',
        },
      }),
    ];

    if (isDark) {
      extensions.push(oneDark);
    }

    const state = EditorState.create({
      doc: value,
      extensions,
    });

    viewRef.current = new EditorView({
      state,
      parent: containerRef.current,
    });
  }, [value, isDark]);

  useEffect(() => {
    createEditor();
    return () => {
      if (viewRef.current) {
        viewRef.current.destroy();
        viewRef.current = null;
      }
    };
  }, [createEditor]);

  return (
    <div
      ref={containerRef}
      className="playground-editor-container"
    />
  );
}
