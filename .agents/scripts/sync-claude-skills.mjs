import path from 'node:path';
import { claudeSkillsRoot, createSkillWrapperContent, listSkillRecords, writeText } from './skill-utils.mjs';

for (const skill of listSkillRecords()) {
  const wrapperPath = path.join(claudeSkillsRoot, skill.name, 'SKILL.md');
  writeText(wrapperPath, createSkillWrapperContent(skill));
  console.log(`Synced ${path.relative(process.cwd(), wrapperPath)}`);
}
