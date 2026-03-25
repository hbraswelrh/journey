// SPDX-License-Identifier: Apache-2.0

import { useState } from 'react';
import { journeyData, type Role } from '../generated/journey-data';
import {
  extractKeywords,
  resolveLayerMappings,
  type ActivityProfile,
} from '../lib/roles';

interface ActivityProbeProps {
  role: Role | null;
  customRoleText?: string;
  onComplete: (profile: ActivityProfile) => void;
  onBack: () => void;
}

export function ActivityProbe({
  role,
  customRoleText,
  onComplete,
  onBack,
}: ActivityProbeProps) {
  const [description, setDescription] = useState('');
  const [selectedCategories, setSelectedCategories] = useState<
    Set<string>
  >(new Set());

  const keywords = extractKeywords(description);
  const categoryKeywords = Array.from(selectedCategories).flatMap(
    (catName) => {
      const cat = journeyData.activityCategories.find(
        (c) => c.name === catName,
      );
      return cat ? [...cat.keywords] : [];
    },
  );
  const allKeywords = [
    ...new Set([...keywords, ...categoryKeywords]),
  ];

  const toggleCategory = (name: string) => {
    setSelectedCategories((prev) => {
      const next = new Set(prev);
      if (next.has(name)) {
        next.delete(name);
      } else {
        next.add(name);
      }
      return next;
    });
  };

  const handleComplete = () => {
    // Include custom role text in keyword extraction if
    // no predefined role was matched.
    const customKeywords = !role && customRoleText
      ? extractKeywords(customRoleText)
      : [];
    const combined = [
      ...new Set([...allKeywords, ...customKeywords]),
    ];

    const profile = resolveLayerMappings(
      role,
      combined,
      description || customRoleText || '',
    );
    onComplete(profile);
  };

  const roleName = role?.name ?? customRoleText ?? 'your role';

  return (
    <div className="card">
      <h2>Describe Your Activities</h2>
      <p>
        Tell us what you work on day-to-day as a{' '}
        <strong>{roleName}</strong>. We will map your activities
        to the relevant Gemara layers and recommend artifacts.
      </p>

      <div className="activity-input">
        <label htmlFor="activity-desc">
          What are your main security-related activities?
        </label>
        <textarea
          id="activity-desc"
          placeholder={`e.g., "I manage CI/CD pipeline security and conduct threat modeling for our deployment infrastructure"`}
          value={description}
          onChange={(e) => setDescription(e.target.value)}
        />
      </div>

      <div style={{ marginTop: '16px' }}>
        <label style={{ fontWeight: 500, color: 'var(--text-h)' }}>
          Or select activity categories:
        </label>
        <div className="category-chips">
          {journeyData.activityCategories.map((cat) => (
            <button
              key={cat.name}
              className={`category-chip ${
                selectedCategories.has(cat.name) ? 'selected' : ''
              }`}
              onClick={() => toggleCategory(cat.name)}
              title={cat.description}
            >
              {cat.name}
            </button>
          ))}
        </div>
      </div>

      {allKeywords.length > 0 && (
        <div style={{ marginTop: '16px' }}>
          <label
            style={{
              fontWeight: 500,
              color: 'var(--text-h)',
              display: 'block',
              marginBottom: '8px',
            }}
          >
            Detected keywords:
          </label>
          <div className="role-keywords">
            {allKeywords.map((kw) => (
              <span key={kw} className="keyword-tag matched">
                {kw}
              </span>
            ))}
          </div>
        </div>
      )}

      <div className="actions">
        <button className="btn btn-secondary" onClick={onBack}>
          Back
        </button>
        <button className="btn btn-primary" onClick={handleComplete}>
          See Results
        </button>
      </div>
    </div>
  );
}
