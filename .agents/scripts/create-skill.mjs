import { execFileSync } from 'node:child_process';
import { existsSync } from 'node:fs';
import path from 'node:path';
import {
  createClaudeWrapperContent,
  createOpenAiYamlContent,
  manifestPath,
  opencodeSkillsRoot,
  parseSkillFrontmatter,
  readManifest,
  readText,
  repoRoot,
  sortManifestSkills,
  titleFromSkillName,
  writeManifest,
  writeText,
} from './skill-utils.mjs';

const validScopes = new Set(['frontend', 'backend', 'docs', 'ops', 'flow', 'meta']);

function printHelp() {
  console.log(`Usage:\n  bun run skills:create -- --name <scope-capability> --scope <scope> --description "..." [options]\n\nOptions:\n  --name <value>                Skill name in kebab-case\n  --scope <value>               One of: frontend, backend, docs, ops, flow, meta\n  --description <value>         Canonical description used in SKILL.md and manifest\n  --title <value>               H1 title for SKILL.md and default display name\n  --display-name <value>        display_name for agents/openai.yaml\n  --short-description <value>   short_description for agents/openai.yaml\n  --default-prompt <value>      default_prompt for agents/openai.yaml\n  --when <value>                Skill Registry text for \"Kapan Dipakai\"\n  --dry-run                     Print planned changes without writing files\n  --help                        Show this help\n`);
}

function parseArgs(argv) {
  const result = { dryRun: false };

  for (let index = 0; index < argv.length; index += 1) {
    const current = argv[index];
    if (!current.startsWith('--')) {
      throw new Error(`Unexpected argument: ${current}`);
    }

    const key = current.slice(2);
    if (key === 'dry-run' || key === 'help') {
      result[key === 'dry-run' ? 'dryRun' : 'help'] = true;
      continue;
    }

    const next = argv[index + 1];
    if (!next || next.startsWith('--')) {
      throw new Error(`Missing value for --${key}`);
    }

    result[key] = next;
    index += 1;
  }

  return result;
}

function ensureValidName(name) {
  if (!/^[a-z0-9]+(?:-[a-z0-9]+)+$/.test(name)) {
    throw new Error('Skill name must be kebab-case and contain at least one dash');
  }
}

function buildSkillMarkdown({ name, description, title }) {
  return `---\nname: ${name}\ndescription: ${description}\n---\n\n# Skill: ${title}\n\n## Quick Context (Required)\n- Folder scope + conventions: \`references/context.md\`\n- Execution checklist: \`templates/checklist.md\`\n\nUse this skill when the task requires a consistent ${title.toLowerCase()} workflow for this repo. Fill in context details, examples, and checklist matching the skill scope before broad use.\n\n## Workflow\n\n1. Read the requirement and constraints of the task this skill will handle.\n2. Update \`references/context.md\` with the target folder, examples, and relevant key patterns.\n3. Execute the main changes following this skill's scope.\n4. Review the final result using \`templates/checklist.md\` before finishing.\n\n## Prohibitions\n\n- **NEVER** modify files outside the skill scope without a clear task reason.\n- **NEVER** leave placeholders, wrong paths, or instructions that conflict with repo rules.\n- **NEVER** move the main skill instructions to \`agents/openai.yaml\` — the source of truth stays in \`SKILL.md\`.\n\n## Pre-Completion Checklist\n\n- [ ] Task scope validated\n- [ ] \`references/context.md\` explains the target folder and key patterns\n- [ ] \`templates/checklist.md\` covers the main verification steps\n- [ ] Skill metadata is consistent across all files\n- [ ] All files end with a newline (EOF)\n`;
}

function buildContextMarkdown(title) {
  return `# Context: ${title}\n\n## Target Folder\n\n- Add the main folders this skill typically touches.\n- Add important files or subtrees that need to be read before execution.\n\n## Real Code Examples\n\n- Add the most relevant example reference from \`.agents/examples/\` if available.\n- If none exists, write the existing code pattern that should be followed.\n\n## Key Patterns\n\n- Summarize architecture decisions, naming conventions, and anti-patterns to avoid.\n- Add commands or reference files that must always be checked when using this skill.\n`;
}

function buildChecklistMarkdown(title) {
  return `# Checklist: ${title}\n\n## Preparation\n\n- [ ] Read \`.agents/settings.json\`\n- [ ] Read \`references/context.md\`\n- [ ] Validate that the task scope matches this skill\n\n## Execution\n\n- [ ] Main changes follow the workflow in \`SKILL.md\`\n- [ ] No files outside scope changed without reason\n- [ ] Relevant examples or guides referenced\n\n## Finalization\n\n- [ ] Skill metadata still consistent\n- [ ] No placeholders or TODOs remaining\n- [ ] All files end with a newline (EOF)\n`;
}

function planFiles({ name, description, title, displayName, shortDescription, defaultPrompt }) {
  return [
    {
      path: path.join(repoRoot, '.agents', 'skills', name, 'SKILL.md'),
      content: buildSkillMarkdown({ name, description, title }),
    },
    {
      path: path.join(repoRoot, '.agents', 'skills', name, 'references', 'context.md'),
      content: buildContextMarkdown(title),
    },
    {
      path: path.join(repoRoot, '.agents', 'skills', name, 'templates', 'checklist.md'),
      content: buildChecklistMarkdown(title),
    },
    {
      path: path.join(repoRoot, '.agents', 'skills', name, 'agents', 'openai.yaml'),
      content: createOpenAiYamlContent({ displayName, shortDescription, defaultPrompt }),
    },
    {
      path: path.join(repoRoot, '.claude', 'skills', name, 'SKILL.md'),
      content: createClaudeWrapperContent({ name, description }),
    },
    {
      path: path.join(opencodeSkillsRoot, name, 'SKILL.md'),
      content: createClaudeWrapperContent({ name, description }),
    },
  ];
}

try {
  const args = parseArgs(process.argv.slice(2));
  if (args.help) {
    printHelp();
    process.exit(0);
  }

  const name = args.name;
  const scope = args.scope;
  const description = args.description;

  if (!name || !scope || !description) {
    printHelp();
    throw new Error('Arguments --name, --scope, and --description are required');
  }

  ensureValidName(name);
  if (!validScopes.has(scope)) {
    throw new Error(`Invalid scope: ${scope}`);
  }

  const manifest = readManifest();
  if (manifest.skills.some((entry) => entry.name === name)) {
    throw new Error(`Skill already exists in manifest: ${name}`);
  }

  const skillDir = path.join(repoRoot, '.agents', 'skills', name);
  if (existsSync(skillDir)) {
    throw new Error(`Skill directory already exists: ${path.relative(repoRoot, skillDir)}`);
  }

  const title = args.title ?? titleFromSkillName(name);
  const displayName = args['display-name'] ?? title;
  const shortDescription = args['short-description'] ?? description;
  const defaultPrompt = args['default-prompt'] ?? `Use $${name} to execute the ${title} workflow for this repository.`;
  const whenToUse = args.when ?? description;

  const files = planFiles({
    name,
    description,
    title,
    displayName,
    shortDescription,
    defaultPrompt,
  });

  const updatedManifest = {
    ...manifest,
    skills: sortManifestSkills([
      ...manifest.skills,
      {
        name,
        description,
        scope,
        whenToUse,
        path: `.agents/skills/${name}`,
      },
    ]),
  };

  const agentsPath = path.join(repoRoot, '.agents', 'AGENTS.md');
  const claudePath = path.join(repoRoot, 'CLAUDE.md');

  if (args.dryRun) {
    console.log(`Dry run: would create skill ${name}`);
    for (const file of files) {
      console.log(`- write ${path.relative(repoRoot, file.path)}`);
    }
    console.log(`- update ${path.relative(repoRoot, manifestPath)}`);
    console.log(`- update ${path.relative(repoRoot, agentsPath)}`);
    console.log(`- update ${path.relative(repoRoot, claudePath)}`);
    process.exit(0);
  }

  for (const file of files) {
    writeText(file.path, file.content);
  }
  writeManifest(updatedManifest);

  execFileSync(process.execPath, [path.join(repoRoot, '.agents', 'scripts', 'sync-skill-registry.mjs')], {
    cwd: repoRoot,
    stdio: 'inherit',
  });
  execFileSync(process.execPath, [path.join(repoRoot, '.agents', 'scripts', 'sync-claude-skills.mjs')], {
    cwd: repoRoot,
    stdio: 'inherit',
  });
  execFileSync(process.execPath, [path.join(repoRoot, '.agents', 'scripts', 'sync-opencode-skills.mjs')], {
    cwd: repoRoot,
    stdio: 'inherit',
  });
  execFileSync(process.execPath, [path.join(repoRoot, '.agents', 'scripts', 'validate-skills.mjs')], {
    cwd: repoRoot,
    stdio: 'inherit',
  });

  const createdSkill = parseSkillFrontmatter(readText(path.join(skillDir, 'SKILL.md')));
  console.log(`Created skill ${createdSkill.name} at .agents/skills/${createdSkill.name}`);
} catch (error) {
  console.error(error.message);
  process.exit(1);
}
