// SPDX-License-Identifier: Apache-2.0

import { useState } from 'react';
import { journeyData, type Role } from '../generated/journey-data';
import { matchRole } from '../lib/roles';

interface RoleSelectionProps {
  selectedRole: Role | null;
  customRoleText: string;
  onSelect: (role: Role) => void;
  onCustomRole: (text: string, matchedRole: Role | null) => void;
  onNext: () => void;
}

export function RoleSelection({
  selectedRole,
  customRoleText,
  onSelect,
  onCustomRole,
  onNext,
}: RoleSelectionProps) {
  const [isCustomActive, setIsCustomActive] = useState(
    customRoleText.length > 0,
  );

  const handleCustomChange = (value: string) => {
    const matched = matchRole(value);
    onCustomRole(value, matched);
  };

  const handleCardSelect = (role: Role) => {
    setIsCustomActive(false);
    onCustomRole('', null);
    onSelect(role);
  };

  const handleCustomFocus = () => {
    setIsCustomActive(true);
  };

  const matchedRole = customRoleText ? matchRole(customRoleText) : null;
  const canContinue = selectedRole !== null || customRoleText.trim().length > 0;

  return (
    <div className="card">
      <h2>Select Your Role</h2>
      <p>
        Choose the role that best describes your job function,
        or type your own. This determines which Gemara layers
        and tutorials are most relevant to you.
      </p>

      <div className="role-grid">
        {journeyData.roles.map((role) => (
          <button
            key={role.name}
            className={`role-card ${
              !isCustomActive && selectedRole?.name === role.name
                ? 'selected'
                : ''
            }`}
            onClick={() => handleCardSelect(role)}
          >
            <h4>{role.name}</h4>
            <p>{role.description}</p>
            <div className="role-keywords">
              {role.defaultKeywords.map((kw) => (
                <span key={kw} className="keyword-tag">
                  {kw}
                </span>
              ))}
            </div>
          </button>
        ))}

        <div
          className={`role-card custom-role-card ${
            isCustomActive ? 'selected' : ''
          }`}
          onClick={handleCustomFocus}
        >
          <h4>Describe Your Own Role</h4>
          <p>
            Type your job title or role description and we
            will match it to the closest Gemara profile.
          </p>
          <input
            type="text"
            className="custom-role-input"
            placeholder="e.g., DevSecOps Lead, Risk Analyst..."
            value={customRoleText}
            onChange={(e) => handleCustomChange(e.target.value)}
            onFocus={handleCustomFocus}
          />
          {isCustomActive && customRoleText.trim() && (
            <div className="custom-role-match">
              {matchedRole ? (
                <span className="match-found">
                  Matched: <strong>{matchedRole.name}</strong>
                  — their defaults will be used as a starting
                  point. Refine further in the next step.
                </span>
              ) : (
                <span className="match-none">
                  No predefined role matched. Your activities
                  in the next step will determine the relevant
                  Gemara layers directly.
                </span>
              )}
            </div>
          )}
        </div>
      </div>

      <div className="actions">
        <button
          className="btn btn-primary"
          disabled={!canContinue}
          onClick={onNext}
        >
          Continue
        </button>
      </div>
    </div>
  );
}
