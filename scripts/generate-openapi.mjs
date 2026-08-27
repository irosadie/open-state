#!/usr/bin/env node
// Generate merged OpenAPI spec from split files in docs/openapi/

import { readFileSync, writeFileSync, readdirSync } from "node:fs"
import { resolve, join } from "node:path"

const ROOT = resolve(process.cwd())
const OPENAPI_DIR = join(ROOT, "docs/openapi")
const OUTPUT = join(ROOT, "docs/openapi.json")

function readJSON(filePath) {
  return JSON.parse(readFileSync(filePath, "utf-8"))
}

function listJSON(dir) {
  try {
    return readdirSync(dir)
      .filter((f) => f.endsWith(".json"))
      .map((f) => join(dir, f))
  } catch {
    return []
  }
}

// Start with base
const spec = readJSON(join(OPENAPI_DIR, "base.json"))

// Fix server URL to Go backend port
if (!spec.servers || spec.servers.length === 0) {
  spec.servers = [{ url: "http://localhost:8020", description: "Local development" }]
} else {
  spec.servers = spec.servers.map((s) =>
    s.url.includes("8020") || s.url.includes("3001") ? { ...s, url: "http://localhost:8020" } : s
  )
}

// Merge paths
spec.paths = spec.paths || {}
for (const file of listJSON(join(OPENAPI_DIR, "paths"))) {
  const paths = readJSON(file)
  Object.assign(spec.paths, paths)
}

// Merge schemas
spec.components = spec.components || {}
spec.components.schemas = spec.components.schemas || {}
for (const file of listJSON(join(OPENAPI_DIR, "schemas"))) {
  const schemas = readJSON(file)
  Object.assign(spec.components.schemas, schemas)
}

writeFileSync(OUTPUT, JSON.stringify(spec, null, 2) + "\n")
process.stdout.write(`[openapi] merged spec written to ${OUTPUT}\n`)
