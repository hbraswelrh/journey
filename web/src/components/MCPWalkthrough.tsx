// SPDX-License-Identifier: Apache-2.0

import { journeyData } from '../generated/journey-data';

interface MCPWalkthroughProps {
  onBack: () => void;
}

export function MCPWalkthrough({ onBack }: MCPWalkthroughProps) {
  const toolReqs = journeyData.mcpRequirements.filter(
    (r) => r.category === 'tools',
  );
  const serverReqs = journeyData.mcpRequirements.filter(
    (r) => r.category === 'server',
  );
  const configReqs = journeyData.mcpRequirements.filter(
    (r) => r.category === 'config',
  );

  const configExample = `{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "${journeyData.config.mcpBinaryName}": {
      "type": "local",
      "command": [
        "/path/to/${journeyData.config.mcpBinaryName}/bin/${journeyData.config.mcpBinaryName}",
        "serve", "--mode", "artifact"
      ],
      "enabled": true
    }
  }
}`;

  return (
    <>
      {/* Prerequisites */}
      <div className="card">
        <h2>MCP Server Requirements</h2>
        <p>
          The Gemara MCP server provides schema validation,
          terminology lookup, and guided artifact creation
          wizards. Here is everything you need to get it
          running.
        </p>

        <h3 style={{ marginTop: '24px', marginBottom: '8px' }}>
          1. Required Tools
        </h3>
        <div className="req-list">
          {toolReqs.map((req) => (
            <div key={req.id} className="req-item">
              <div
                className={`req-icon ${req.required ? 'required' : 'optional'}`}
              >
                {req.required ? '\u2757' : '\u2139\uFE0F'}
              </div>
              <div className="req-details">
                <h4>
                  {req.name}
                  <span
                    className={`req-badge ${req.required ? 'required' : 'optional'}`}
                  >
                    {req.required ? 'Required' : 'Optional'}
                  </span>
                </h4>
                <p>{req.description}</p>
                {req.installCmd && (
                  <div className="req-install">
                    <pre>{req.installCmd}</pre>
                  </div>
                )}
                {req.installUrl && (
                  <div className="req-install">
                    <a
                      href={req.installUrl}
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      {req.installUrl}
                    </a>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Server Installation */}
      <div className="card">
        <h3>2. Install the MCP Server</h3>
        <p>
          The gemara-mcp server must be built from source. It is
          a Go binary hosted at{' '}
          <a
            href={journeyData.config.gemaraMcpRepoUrl}
            target="_blank"
            rel="noopener noreferrer"
          >
            {journeyData.config.gemaraMcpRepoUrl}
          </a>
          .
        </p>

        <div className="req-list">
          {serverReqs.map((req) => (
            <div key={req.id} className="req-item">
              <div className="req-icon required">{'\u26A1'}</div>
              <div className="req-details">
                <h4>{req.name}</h4>
                <p>{req.description}</p>
                {req.installCmd && (
                  <div className="req-install">
                    <pre>{req.installCmd}</pre>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>

        <p style={{ marginTop: '12px', fontSize: '14px' }}>
          Alternatively, run{' '}
          <code>./journey</code> and select "Build from
          source" for automated installation with SHA-pinned
          commits.
        </p>
      </div>

      {/* Configuration */}
      <div className="card">
        <h3>3. Configure OpenCode</h3>
        <p>
          After building, create or update{' '}
          <code>{journeyData.config.openCodeConfigFile}</code> in
          your project directory:
        </p>

        {configReqs.map((req) => (
          <div key={req.id} style={{ marginTop: '12px' }}>
            <p style={{ fontSize: '14px' }}>{req.description}</p>
          </div>
        ))}

        <div className="config-preview">
          <h4 style={{ marginTop: '16px', marginBottom: '8px' }}>
            Example {journeyData.config.openCodeConfigFile}:
          </h4>
          <pre>{configExample}</pre>
        </div>
      </div>

      {/* Server Modes */}
      <div className="card">
        <h3>4. Choose a Server Mode</h3>
        <p>
          The MCP server supports two operating modes. Choose the
          one that fits your workflow.
        </p>

        <div className="mode-cards">
          {journeyData.mcpModes.map((mode) => (
            <div
              key={mode.id}
              className={`mode-card ${mode.recommended ? 'recommended' : ''}`}
            >
              <h4>
                {mode.name}
                {mode.recommended && (
                  <span className="recommended-badge">
                    Recommended
                  </span>
                )}
              </h4>
              <p>{mode.description}</p>
            </div>
          ))}
        </div>
      </div>

      {/* Capabilities */}
      <div className="card">
        <h3>5. Available MCP Capabilities</h3>
        <p>
          Once configured, OpenCode connects to the gemara-mcp
          server automatically. These capabilities become
          available in your session:
        </p>

        <table className="cap-table">
          <thead>
            <tr>
              <th>Category</th>
              <th>Name</th>
              <th>Description</th>
            </tr>
          </thead>
          <tbody>
            {journeyData.mcpCapabilities.map((cap) => (
              <tr key={cap.name}>
                <td>
                  <span className={`cap-category ${cap.category}`}>
                    {cap.category}
                  </span>
                </td>
                <td>
                  <code>{cap.name}</code>
                </td>
                <td>{cap.description}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Verification */}
      <div className="card">
        <h3>6. Verify Your Setup</h3>
        <p>
          Run the doctor command to verify everything is properly
          configured:
        </p>
        <pre style={{ marginTop: '12px' }}>./journey --doctor</pre>
        <p style={{ marginTop: '12px', fontSize: '14px' }}>
          This checks for all required tools, validates the
          opencode.json configuration, and confirms the server
          mode. All checks should pass before you start
          authoring.
        </p>

        <h4 style={{ marginTop: '20px', marginBottom: '8px' }}>
          Then start OpenCode:
        </h4>
        <pre>opencode</pre>
        <p style={{ marginTop: '12px', fontSize: '14px' }}>
          OpenCode reads the config and launches the gemara-mcp
          server as a background process. All MCP capabilities
          are then available in your session.
        </p>
      </div>

      {/* Community */}
      <div className="card">
        <h3>Share Your Journey</h3>
        <p>
          Completed the tutorials and set up the MCP server?
          Share your experience, ask questions, or show what
          you built in a GitHub Discussion.
        </p>
        <div style={{ marginTop: '12px' }}>
          <a
            href={journeyData.config.newDiscussionUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="btn btn-discussion"
          >
            Open a Discussion
          </a>
        </div>
      </div>

      <div className="actions">
        <button className="btn btn-secondary" onClick={onBack}>
          Back to Tutorials
        </button>
      </div>
    </>
  );
}
