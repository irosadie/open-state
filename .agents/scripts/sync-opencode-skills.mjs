import path from 'node:path';
import { createSkillWrapperContent, listSkillRecords, opencodeSkillsRoot, writeText } from './skill-utils.mjs';

for (const skill of listSkillRecords()) {
  const wrapperPath = path.join(opencodeSkillsRoot, skill.name, 'SKILL.md');
  writeText(wrapperPath, createSkillWrapperContent(skill));
  console.log(`Synced ${path.relative(process.cwd(), wrapperPath)}`);
}
