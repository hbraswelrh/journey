// SPDX-License-Identifier: Apache-2.0

// Client-side role discovery logic ported from the Go
// codebase (internal/roles/). Operates on the generated
// journey-data.ts constants.

import {
  journeyData,
  type Role,
  type ArtifactType,
} from '../generated/journey-data';

export type Confidence = 'inferred' | 'strong';

export interface LayerMapping {
  layer: number;
  confidence: Confidence;
  keywords: string[];
}

export interface ArtifactRecommendation {
  artifactType: string;
  schemaDef: string;
  description: string;
  layer: number;
  confidence: Confidence;
  mcpWizard: string;
  authoringApproach: string;
  checklist: string[];
}

export interface ActivityProfile {
  extractedKeywords: string[];
  matchedCategories: string[];
  resolvedLayers: LayerMapping[];
  userDescription: string;
  role: Role | null;
  recommendations: ArtifactRecommendation[];
}

/**
 * extractKeywords identifies domain keywords from free-text
 * using the LayerKeywords vocabulary. Longest match wins.
 */
export function extractKeywords(description: string): string[] {
  if (!description) return [];

  const lower = description.toLowerCase();
  const keywords = Object.keys(journeyData.layerKeywords);

  // Sort by length descending for longest-match-first.
  keywords.sort((a, b) => b.length - a.length);

  const found: string[] = [];
  const matched = new Set<string>();

  for (const kw of keywords) {
    if (lower.includes(kw) && !matched.has(kw)) {
      found.push(kw);
      matched.add(kw);
    }
  }

  return found;
}

/**
 * clarificationNeeded returns keywords that map to multiple
 * layers and require user clarification.
 */
export function clarificationNeeded(keywords: string[]): string[] {
  const ambiguous: string[] = [];
  for (const kw of keywords) {
    const lower = kw.toLowerCase();
    const layers = (journeyData.layerKeywords as Record<string, readonly number[]>)[lower];
    if (layers && layers.length > 1) {
      ambiguous.push(kw);
    }
  }
  return ambiguous;
}

/**
 * resolveLayerMappings combines role defaults with
 * keyword-extracted layers.
 */
export function resolveLayerMappings(
  role: Role | null,
  keywords: string[],
  description: string,
): ActivityProfile {
  const layerMap = new Map<number, LayerMapping>();

  // Add role defaults as inferred confidence.
  if (role) {
    for (const layer of role.defaultLayers) {
      layerMap.set(layer, {
        layer,
        confidence: 'inferred',
        keywords: [],
      });
    }
  }

  // Add keyword-resolved layers as strong confidence.
  for (const kw of keywords) {
    const lower = kw.toLowerCase();
    const layers = (journeyData.layerKeywords as Record<string, readonly number[]>)[lower];
    if (!layers) continue;

    for (const layer of layers) {
      const existing = layerMap.get(layer);
      if (existing) {
        existing.confidence = 'strong';
        existing.keywords.push(kw);
      } else {
        layerMap.set(layer, {
          layer,
          confidence: 'strong',
          keywords: [kw],
        });
      }
    }
  }

  // Sort: strong first, then by layer number.
  const layers = Array.from(layerMap.values()).sort((a, b) => {
    if (a.confidence !== b.confidence) {
      return a.confidence === 'strong' ? -1 : 1;
    }
    return a.layer - b.layer;
  });

  const categories = resolveCategories(keywords);
  const profile: ActivityProfile = {
    extractedKeywords: keywords,
    matchedCategories: categories,
    resolvedLayers: layers,
    userDescription: description,
    role,
    recommendations: [],
  };

  profile.recommendations = buildRecommendations(profile);
  return profile;
}

/**
 * resolveCategories maps keywords to activity category names.
 */
function resolveCategories(keywords: string[]): string[] {
  const matched = new Set<string>();
  const result: string[] = [];

  for (const kw of keywords) {
    const lower = kw.toLowerCase();
    for (const cat of journeyData.activityCategories) {
      for (const catKw of cat.keywords) {
        if (lower === catKw.toLowerCase() && !matched.has(cat.name)) {
          matched.add(cat.name);
          result.push(cat.name);
        }
      }
    }
  }

  return result;
}

/**
 * buildRecommendations produces artifact recommendations
 * from resolved layers.
 */
function buildRecommendations(
  profile: ActivityProfile,
): ArtifactRecommendation[] {
  if (profile.resolvedLayers.length === 0) return [];

  const seen = new Set<string>();
  const recs: ArtifactRecommendation[] = [];

  for (const lm of profile.resolvedLayers) {
    const layer = journeyData.layers.find((l) => l.number === lm.layer);
    if (!layer) continue;

    for (const artId of layer.artifactIds) {
      if (seen.has(artId)) continue;
      seen.add(artId);

      const art = journeyData.artifactTypes.find(
        (a: ArtifactType) => a.id === artId,
      );
      if (!art) continue;

      recs.push({
        artifactType: art.id,
        schemaDef: art.schemaDef,
        description: art.description,
        layer: lm.layer,
        confidence: lm.confidence,
        mcpWizard: art.mcpWizard ?? '',
        authoringApproach: art.authoringApproach,
        checklist: [...art.checklist],
      });
    }
  }

  return recs;
}

/**
 * matchRole finds the best predefined role match for
 * free-text input.
 */
export function matchRole(input: string): Role | null {
  const trimmed = input.trim();
  if (!trimmed) return null;

  const lower = trimmed.toLowerCase();

  // Exact match.
  for (const role of journeyData.roles) {
    if (lower === role.name.toLowerCase()) {
      return role;
    }
  }

  // Partial match.
  for (const role of journeyData.roles) {
    const roleLower = role.name.toLowerCase();
    if (lower.includes(roleLower) || roleLower.includes(lower)) {
      return role;
    }
  }

  return null;
}
