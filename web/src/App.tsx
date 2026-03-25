// SPDX-License-Identifier: Apache-2.0

import { useState } from 'react';
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

function isAfter(current: Step, target: Step): boolean {
  return stepOrder.indexOf(current) > stepOrder.indexOf(target);
}

function App() {
  const [currentStep, setCurrentStep] = useState<Step>('role');
  const [selectedRole, setSelectedRole] = useState<Role | null>(
    null,
  );
  const [customRoleText, setCustomRoleText] = useState('');
  const [profile, setProfile] = useState<ActivityProfile | null>(
    null,
  );

  const steps = [
    {
      label: 'Role',
      completed: isAfter(currentStep, 'role') && selectedRole !== null,
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
          onGoToTutorials={() => setCurrentStep('tutorials')}
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
