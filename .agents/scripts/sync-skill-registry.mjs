import path from 'node:path';
import {
  agentsRoot,
  readManifest,
  readText,
  renderOpenCodeSkillList,
  renderSkillReferenceList,
  renderSkillRegistryTable,
  replaceMarkedSection,
  repoRoot,
  sortManifestSkills,
  writeManifest,
  writeText,
} from './skill-utils.mjs';

const manifest = readManifest();
manifest.skills = sortManifestSkills(manifest.skills ?? []);
writeManifest(manifest);

const agentsPath = path.join(agentsRoot, 'AGENTS.md');
const claudePath = path.join(repoRoot, 'CLAUDE.md');
const opencodePath = path.join(repoRoot, 'opencode.md');

const registryTable = renderSkillRegistryTable(manifest.skills);
const skillReferenceList = renderSkillReferenceList(manifest.skills);
const opencodeSkillList = renderOpenCodeSkillList(manifest.skills);

const updatedAgents = replaceMarkedSection(
  replaceMarkedSection(
    readText(agentsPath),
    '<!-- skill-registry:start -->',
    '<!-- skill-registry:end -->',
    registryTable,
  ),
  '<!-- skill-links:start -->',
  '<!-- skill-links:end -->',
  skillReferenceList,
);

const updatedClaude = replaceMarkedSection(
  readText(claudePath),
  '<!-- skill-registry:start -->',
  '<!-- skill-registry:end -->',
  registryTable,
);

const updatedOpenCode = replaceMarkedSection(
  readText(opencodePath),
  '<!-- skill-links:start -->',
  '<!-- skill-links:end -->',
  opencodeSkillList,
);

writeText(agentsPath, updatedAgents);
writeText(claudePath, updatedClaude);
writeText(opencodePath, updatedOpenCode);

console.log(`Synced skill registry sections in ${path.relative(repoRoot, agentsPath)}, ${path.relative(repoRoot, claudePath)}, and ${path.relative(repoRoot, opencodePath)}`);
