import { defineConfig, type Plugin } from 'vite'
import react from '@vitejs/plugin-react'
import { execFile } from 'node:child_process'
import { writeFile, unlink } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

// https://vite.dev/config/
// Vite's default appType is 'spa', which serves index.html
// for all unmatched routes (history API fallback).
// This enables direct navigation to /playground.

/** Only these CUE definitions may be passed to cue vet. */
const ALLOWED_DEFINITIONS = new Set([
  '#ControlCatalog',
  '#GuidanceCatalog',
  '#ThreatCatalog',
  '#RiskCatalog',
  '#Policy',
])

/** Maximum request body size in bytes (1 MB). */
const MAX_BODY_BYTES = 1024 * 1024

/** Minimum interval between requests in ms. */
const RATE_LIMIT_MS = 1000

/**
 * Vite plugin that exposes a POST /api/validate endpoint
 * during development. It writes the YAML content to a temp
 * file and runs `cue vet` against the Gemara CUE registry
 * module for schema validation.
 *
 * Security mitigations:
 * - Definition parameter validated against an allowlist
 * - Request body capped at 1 MB
 * - Basic rate limiting (1 request per second)
 */
function cueValidatePlugin(): Plugin {
  let lastRequestTime = 0

  return {
    name: 'cue-validate',
    configureServer(server) {
      server.middlewares.use('/api/validate', async (req, res) => {
        if (req.method !== 'POST') {
          res.statusCode = 405
          res.end(JSON.stringify({ error: 'Method not allowed' }))
          return
        }

        // Rate limiting.
        const now = Date.now()
        if (now - lastRequestTime < RATE_LIMIT_MS) {
          res.statusCode = 429
          res.end(
            JSON.stringify({ error: 'Too many requests. Try again shortly.' }),
          )
          return
        }
        lastRequestTime = now

        // Read request body with size limit.
        const chunks: Buffer[] = []
        let totalBytes = 0
        for await (const chunk of req) {
          totalBytes += (chunk as Buffer).length
          if (totalBytes > MAX_BODY_BYTES) {
            res.statusCode = 413
            res.end(
              JSON.stringify({ error: 'Request body too large (max 1 MB)' }),
            )
            return
          }
          chunks.push(chunk as Buffer)
        }

        let body: { content: string; definition: string }
        try {
          body = JSON.parse(Buffer.concat(chunks).toString())
        } catch {
          res.statusCode = 400
          res.end(JSON.stringify({ error: 'Invalid JSON' }))
          return
        }

        if (!body.content || !body.definition) {
          res.statusCode = 400
          res.end(
            JSON.stringify({
              error: 'Missing content or definition',
            }),
          )
          return
        }

        // Validate definition against allowlist.
        if (!ALLOWED_DEFINITIONS.has(body.definition)) {
          res.statusCode = 400
          res.end(
            JSON.stringify({
              error: `Invalid definition "${body.definition}". ` +
                `Allowed: ${[...ALLOWED_DEFINITIONS].join(', ')}`,
            }),
          )
          return
        }

        // Write YAML to a temp file for cue vet.
        const tmpFile = join(
          tmpdir(),
          `gemara-validate-${Date.now()}.yaml`,
        )

        try {
          await writeFile(tmpFile, body.content, 'utf-8')

          const result = await new Promise<{
            valid: boolean
            errors: string[]
          }>((resolve) => {
            execFile(
              'cue',
              [
                'vet',
                '-c',
                '-d',
                body.definition,
                'github.com/gemaraproj/gemara@latest',
                tmpFile,
              ],
              { timeout: 15000 },
              (error, _stdout, stderr) => {
                if (error) {
                  const errOutput = stderr.trim()
                  const errors = errOutput
                    ? errOutput.split('\n')
                    : ['Validation failed']
                  resolve({ valid: false, errors })
                } else {
                  resolve({ valid: true, errors: [] })
                }
              },
            )
          })

          res.setHeader('Content-Type', 'application/json')
          res.end(JSON.stringify(result))
        } catch (err) {
          res.statusCode = 500
          res.end(
            JSON.stringify({
              error:
                err instanceof Error
                  ? err.message
                  : 'Internal error',
            }),
          )
        } finally {
          unlink(tmpFile).catch(() => {})
        }
      })
    },
  }
}

export default defineConfig({
  plugins: [react(), cueValidatePlugin()],
  base: process.env.VITE_BASE_PATH || '/',
})
