import path from 'node:path';
import { claudeSkillsRoot, createClaudeWrapperContent, listSkillRecords, writeText } from './skill-utils.mjs';

for (const skill of listSkillRecords()) {
  const wrapperPath = path.join(claudeSkillsRoot, skill.name, 'SKILL.md');
  writeText(wrapperPath, createClaudeWrapperContent(skill));
  console.log(`Synced ${path.relative(process.cwd(), wrapperPath)}`);
}
