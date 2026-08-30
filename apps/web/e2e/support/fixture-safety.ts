import { GOLDEN_JOURNEYS } from "../fixtures/golden-journeys"

const DISALLOWED_KEY =
  /(password|secret|access[_-]?token|refresh[_-]?token|authorization|cookie|raw[_-]?(prompt|response)|rag[_-]?(document|content)|credential)/i
const DISALLOWED_VALUE =
  /(-----BEGIN|sk-[a-z0-9]|bearer\s+[a-z0-9._-]{16,}|https?:\/\/(?!127\.0\.0\.1|localhost))/i

function inspectValue(value: unknown, path: string): void {
  if (Array.isArray(value)) {
    value.forEach((item, index) => inspectValue(item, `${path}[${index}]`))
    return
  }

  if (typeof value === "string") {
    if (DISALLOWED_VALUE.test(value)) {
      throw new Error(`unsafe fixture value at ${path}`)
    }
    return
  }

  if (!value || typeof value !== "object") return

  for (const [key, child] of Object.entries(value)) {
    if (DISALLOWED_KEY.test(key)) {
      throw new Error(`unsafe fixture field at ${path}.${key}`)
    }
    inspectValue(child, `${path}.${key}`)
  }
}

export function validateGoldenManifests(): void {
  inspectValue(GOLDEN_JOURNEYS, "GOLDEN_JOURNEYS")
}
