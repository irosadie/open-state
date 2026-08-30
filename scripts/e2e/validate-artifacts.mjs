import { readdir, readFile, stat } from "node:fs/promises"
import { join, relative } from "node:path"

const root = new URL("../../apps/web/test-results/e2e/", import.meta.url)
const rootPath = root.pathname
const maxFileBytes = 10 * 1024 * 1024
const maxTotalBytes = 40 * 1024 * 1024
const forbidden = [
  /-----BEGIN/i,
  /\bsk-[a-z0-9]/i,
  /raw[_-]?(prompt|response)/i,
  /rag[_-]?(document|content)/i,
  /authorization\s*:/i,
  /openstate-e2e-fixture-password/i,
]

async function walk(directory) {
  const entries = await readdir(directory, { withFileTypes: true }).catch(() => [])
  const files = []
  for (const entry of entries) {
    const path = join(directory, entry.name)
    if (entry.isDirectory()) files.push(...(await walk(path)))
    else files.push(path)
  }
  return files
}

const files = await walk(rootPath)
let totalBytes = 0
for (const path of files) {
  const fileStat = await stat(path)
  totalBytes += fileStat.size
  if (fileStat.size > maxFileBytes) {
    throw new Error(`E2E artifact exceeds the per-file bound: ${relative(rootPath, path)}`)
  }
  const contents = await readFile(path)
  const text = contents.toString("utf8")
  if (forbidden.some((pattern) => pattern.test(text))) {
    throw new Error(`E2E artifact rejected by the synthetic-data policy: ${relative(rootPath, path)}`)
  }
}

if (totalBytes > maxTotalBytes) {
  throw new Error("E2E artifacts exceed the total-size bound")
}
