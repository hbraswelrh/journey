// SPDX-License-Identifier: Apache-2.0

interface Step {
  label: string;
  completed: boolean;
  active: boolean;
}

interface StepperProps {
  steps: Step[];
}

export function Stepper({ steps }: StepperProps) {
  return (
    <div className="stepper">
      {steps.map((step, i) => {
        const cls = step.completed
          ? 'completed'
          : step.active
            ? 'active'
            : '';
        return (
          <div key={i} className={`step-indicator ${cls}`}>
            <span className="step-number">
              {step.completed ? '\u2713' : i + 1}
            </span>
            {step.label}
          </div>
        );
      })}
    </div>
  );
}
