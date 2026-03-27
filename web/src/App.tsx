// SPDX-License-Identifier: Apache-2.0

import { useState, useEffect, useRef } from 'react';
import { journeyData } from './generated/journey-data';
import type { Role } from './generated/journey-data';
import type { ActivityProfile } from './lib/roles';
import { Stepper } from './components/Stepper';
import { RoleSelection } from './components/RoleSelection';
import { ActivityProbe } from './components/ActivityProbe';
import { Results } from './components/Results';
import { TutorialSuggestions } from './components/TutorialSuggestions';
import { MCPWalkthrough } from './components/MCPWalkthrough';
import './App.css';

type Step =
  | 'role'
  | 'activity'
  | 'results'
  | 'tutorials'
  | 'mcp';

const stepOrder: Step[] = [
  'role',
  'activity',
  'results',
  'tutorials',
  'mcp',
];

const STORAGE_KEY = 'gemara-journey-state';

interface SavedState {
  currentStep: Step;
  roleName: string | null;
  customRoleText: string;
  profile: ActivityProfile | null;
}

function saveToStorage(state: SavedState): void {
  try {
    localStorage.setItem(
      STORAGE_KEY,
      JSON.stringify(state),
    );
  } catch {
    // Ignore storage errors.
  }
}

function loadFromStorage(): SavedState | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as SavedState;
    if (!stepOrder.includes(parsed.currentStep)) return null;
    return parsed;
  } catch {
    return null;
  }
}

function clearStorage(): void {
  try {
    localStorage.removeItem(STORAGE_KEY);
  } catch {
    // Ignore.
  }
}

/** Look up a Role object by name from the generated data. */
function findRole(name: string | null): Role | null {
  if (!name) return null;
  return (
    journeyData.roles.find((r) => r.name === name) ?? null
  );
}

function isAfter(current: Step, target: Step): boolean {
  return stepOrder.indexOf(current) > stepOrder.indexOf(target);
}

function App() {
  // Load saved state once on mount.
  const [savedSnapshot] = useState(() => loadFromStorage());
  const hasSavedProgress =
    savedSnapshot !== null &&
    savedSnapshot.currentStep !== 'role';

  const [showResumePrompt, setShowResumePrompt] =
    useState(hasSavedProgress);
  const [currentStep, setCurrentStep] =
    useState<Step>('role');
  const [selectedRole, setSelectedRole] =
    useState<Role | null>(null);
  const [customRoleText, setCustomRoleText] = useState('');
  const [profile, setProfile] =
    useState<ActivityProfile | null>(null);

  // Use refs to always have current values for saving,
  // avoiding stale closure issues.
  const stateRef = useRef({
    currentStep: 'role' as Step,
    selectedRole: null as Role | null,
    customRoleText: '',
    profile: null as ActivityProfile | null,
  });

  // Keep ref in sync.
  stateRef.current = {
    currentStep,
    selectedRole,
    customRoleText,
    profile,
  };

  // Save to sessionStorage whenever state changes.
  useEffect(() => {
    // Don't save while showing the resume prompt.
    if (showResumePrompt) return;
    saveToStorage({
      currentStep,
      roleName: selectedRole?.name ?? null,
      customRoleText,
      profile,
    });
  }, [currentStep, selectedRole, customRoleText, profile, showResumePrompt]);

  const handleResume = () => {
    if (!savedSnapshot) return;
    const role = findRole(savedSnapshot.roleName);
    const restoredProfile = savedSnapshot.profile
      ? {
          ...savedSnapshot.profile,
          role: savedSnapshot.profile.role
            ? findRole(savedSnapshot.profile.role.name)
            : null,
        }
      : null;

    setSelectedRole(role);
    setCustomRoleText(savedSnapshot.customRoleText ?? '');
    setProfile(restoredProfile);
    setCurrentStep(savedSnapshot.currentStep);
    setShowResumePrompt(false);
  };

  const handleRestart = () => {
    clearStorage();
    setCurrentStep('role');
    setSelectedRole(null);
    setCustomRoleText('');
    setProfile(null);
    setShowResumePrompt(false);
  };

  const steps = [
    {
      label: 'Role',
      completed:
        isAfter(currentStep, 'role') &&
        selectedRole !== null,
      active: currentStep === 'role',
    },
    {
      label: 'Activities',
      completed: isAfter(currentStep, 'activity'),
      active: currentStep === 'activity',
    },
    {
      label: 'Results',
      completed: isAfter(currentStep, 'results'),
      active: currentStep === 'results',
    },
    {
      label: 'Tutorials',
      completed: isAfter(currentStep, 'tutorials'),
      active: currentStep === 'tutorials',
    },
    {
      label: 'MCP Setup',
      completed: false,
      active: currentStep === 'mcp',
    },
  ];

  // Show resume/restart prompt if saved progress exists.
  if (showResumePrompt && savedSnapshot) {
    const stepLabels: Record<string, string> = {
      role: 'Role',
      activity: 'Activities',
      results: 'Results',
      tutorials: 'Tutorials',
      mcp: 'MCP Setup',
    };
    const stepLabel =
      stepLabels[savedSnapshot.currentStep] ??
      savedSnapshot.currentStep;

    return (
      <div className="app">
        <header className="header">
          <h1>Gemara User Journey</h1>
          <p>Gemara Role Discovery &amp; MCP Setup</p>
        </header>

        <div className="card">
          <h2>Welcome Back</h2>
          <p>
            You have a session in progress
            {savedSnapshot.roleName
              ? ` as a ${savedSnapshot.roleName}`
              : ''}
            . You were on the{' '}
            <strong>{stepLabel}</strong> step.
          </p>
          <div
            className="actions"
            style={{ marginTop: '24px' }}
          >
            <button
              className="btn btn-secondary"
              onClick={handleRestart}
            >
              Start Over
            </button>
            <button
              className="btn btn-primary"
              onClick={handleResume}
            >
              Continue Where I Left Off
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="app">
      <header className="header">
        <h1>Gemara User Journey</h1>
        <p>Gemara Role Discovery &amp; MCP Setup</p>
        <a
          href={journeyData.config.newDiscussionUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="btn btn-discussion"
          style={{ marginTop: '12px' }}
        >
          Open a Discussion
        </a>
      </header>

      <Stepper steps={steps} />

      {currentStep === 'role' && (
        <RoleSelection
          selectedRole={selectedRole}
          customRoleText={customRoleText}
          onSelect={(role) => {
            setSelectedRole(role);
            setCustomRoleText('');
          }}
          onCustomRole={(text, matched) => {
            setCustomRoleText(text);
            setSelectedRole(matched);
          }}
          onNext={() => setCurrentStep('activity')}
        />
      )}

      {currentStep === 'activity' && (
        <ActivityProbe
          role={selectedRole}
          customRoleText={customRoleText}
          onComplete={(p) => {
            setProfile(p);
            setCurrentStep('results');
          }}
          onBack={() => setCurrentStep('role')}
        />
      )}

      {currentStep === 'results' && profile && (
        <Results
          profile={profile}
          onBack={() => setCurrentStep('activity')}
          onGoToTutorials={() =>
            setCurrentStep('tutorials')
          }
        />
      )}

      {currentStep === 'tutorials' && profile && (
        <TutorialSuggestions
          profile={profile}
          onBack={() => setCurrentStep('results')}
          onGoToMCP={() => setCurrentStep('mcp')}
        />
      )}

      {currentStep === 'mcp' && (
        <MCPWalkthrough
          onBack={() => setCurrentStep('tutorials')}
        />
      )}
    </div>
  );
}

export default App;
