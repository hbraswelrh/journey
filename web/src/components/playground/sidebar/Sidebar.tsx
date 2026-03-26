// SPDX-License-Identifier: Apache-2.0

import { useState } from 'react';
import { ArtifactSelector } from './ArtifactSelector';
import { SchemaPanel } from './SchemaPanel';
import { LexiconPanel } from './LexiconPanel';

type SidebarTab = 'artifacts' | 'schema' | 'lexicon';

interface SidebarProps {
  activeType: string;
  collapsed: boolean;
  onSelectType: (type: string) => void;
  onToggleCollapse: () => void;
}

/**
 * Sidebar provides tabbed navigation between Artifacts,
 * Schema, and Lexicon panels for the playground.
 */
export function Sidebar({
  activeType,
  collapsed,
  onSelectType,
  onToggleCollapse,
}: SidebarProps) {
  const [activeTab, setActiveTab] =
    useState<SidebarTab>('artifacts');

  if (collapsed) {
    return (
      <aside className="pg-sidebar collapsed">
        <button
          className="pg-sidebar-toggle"
          onClick={onToggleCollapse}
          title="Expand sidebar"
        >
          &raquo;
        </button>
        <div className="pg-sidebar-icons">
          <button
            className={`pg-sidebar-icon ${
              activeTab === 'artifacts' ? 'active' : ''
            }`}
            onClick={() => {
              setActiveTab('artifacts');
              onToggleCollapse();
            }}
            title="Artifacts"
          >
            A
          </button>
          <button
            className={`pg-sidebar-icon ${
              activeTab === 'schema' ? 'active' : ''
            }`}
            onClick={() => {
              setActiveTab('schema');
              onToggleCollapse();
            }}
            title="Schema"
          >
            S
          </button>
          <button
            className={`pg-sidebar-icon ${
              activeTab === 'lexicon' ? 'active' : ''
            }`}
            onClick={() => {
              setActiveTab('lexicon');
              onToggleCollapse();
            }}
            title="Lexicon"
          >
            L
          </button>
        </div>
      </aside>
    );
  }

  return (
    <aside className="pg-sidebar">
      <div className="pg-sidebar-header">
        <div className="pg-sidebar-tabs">
          <button
            className={`pg-sidebar-tab ${
              activeTab === 'artifacts' ? 'active' : ''
            }`}
            onClick={() => setActiveTab('artifacts')}
          >
            Artifacts
          </button>
          <button
            className={`pg-sidebar-tab ${
              activeTab === 'schema' ? 'active' : ''
            }`}
            onClick={() => setActiveTab('schema')}
          >
            Schema
          </button>
          <button
            className={`pg-sidebar-tab ${
              activeTab === 'lexicon' ? 'active' : ''
            }`}
            onClick={() => setActiveTab('lexicon')}
          >
            Lexicon
          </button>
        </div>
        <button
          className="pg-sidebar-toggle"
          onClick={onToggleCollapse}
          title="Collapse sidebar"
        >
          &laquo;
        </button>
      </div>

      <div className="pg-sidebar-content">
        {activeTab === 'artifacts' && (
          <ArtifactSelector
            activeType={activeType}
            onSelect={onSelectType}
          />
        )}
        {activeTab === 'schema' && (
          <SchemaPanel activeType={activeType} />
        )}
        {activeTab === 'lexicon' && (
          <LexiconPanel activeType={activeType} />
        )}
      </div>
    </aside>
  );
}
