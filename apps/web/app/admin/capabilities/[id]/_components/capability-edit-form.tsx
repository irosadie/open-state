"use client"

import { useState } from "react"

import { PermissionGate } from "$/components/auth-guard/permission-gate"
import { Button } from "$/components/button"
import { Input } from "$/components/input"
import { Select } from "$/components/select"
import { Textarea } from "$/components/textarea"
import {
  type UpdateCapabilitySchemaProps,
  capabilityStatusLabels,
  providerTypeLabels,
  updateCapabilitySchema,
} from "@openstate/schemas"
import type { CapabilityResponse } from "@openstate/types"
import type { ZodError } from "zod"

type CapabilityEditFormProps = {
  capability: CapabilityResponse
  isPending: boolean
  onCancel: () => void
  onSave: (payload: UpdateCapabilitySchemaProps) => Promise<void>
}

type FieldErrors = Partial<Record<keyof UpdateCapabilitySchemaProps, string>>

const parseJsonField = (raw: string): Record<string, unknown> | undefined => {
  if (!raw.trim()) return undefined

  try {
    return JSON.parse(raw) as Record<string, unknown>
  } catch {
    return undefined
  }
}

export function CapabilityEditForm({
  capability,
  isPending,
  onCancel,
  onSave,
}: CapabilityEditFormProps) {
  const [description, setDescription] = useState(capability.description || "")
  const [providerType, setProviderType] = useState(capability.providerType)
  const [providerId, setProviderId] = useState(capability.providerId || "")
  const [inputSchema, setInputSchema] = useState(
    JSON.stringify(capability.inputSchema || {}, null, 2),
  )
  const [outputSchema, setOutputSchema] = useState(
    JSON.stringify(capability.outputSchema || {}, null, 2),
  )
  const [status, setStatus] = useState(capability.status)
  const [version, setVersion] = useState(String(capability.version))
  const [credentialReference, setCredentialReference] = useState(
    capability.credentialReference || "",
  )
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})

  const handleSave = async () => {
    const parsedInput = parseJsonField(inputSchema)
    const parsedOutput = parseJsonField(outputSchema)

    if (parsedInput === undefined) {
      setFieldErrors({ inputSchema: "Input schema must be valid JSON" })

      return
    }

    if (parsedOutput === undefined) {
      setFieldErrors({ outputSchema: "Output schema must be valid JSON" })

      return
    }

    const parsed = updateCapabilitySchema.safeParse({
      description: description || undefined,
      providerType,
      providerId: providerId || undefined,
      inputSchema: parsedInput,
      outputSchema: parsedOutput,
      status,
      version: version ? Number(version) : undefined,
      credentialReference: credentialReference || undefined,
    })

    if (!parsed.success) {
      const errors: FieldErrors = {}

      for (const issue of (parsed.error as ZodError).issues) {
        const key = issue.path[0] as keyof UpdateCapabilitySchemaProps

        if (key) {
          errors[key] = issue.message
        }
      }

      setFieldErrors(errors)

      return
    }

    await onSave(parsed.data)
  }

  return (
    <div className="space-y-4 px-1">
      <Input
        label="Description"
        value={description}
        onChange={(e) => setDescription(e.target.value)}
      />

      <Select<(typeof providerTypeLabels)[number]>
        label="Provider type"
        options={providerTypeLabels}
        value={
          providerTypeLabels.find((l) => l.value === providerType) ?? undefined
        }
        getOptionLabel={(o) => o.label}
        getOptionValue={(o) => o.value}
        onChange={(option) =>
          setProviderType(
            (option as (typeof providerTypeLabels)[number])
              .value as CapabilityResponse["providerType"],
          )
        }
        error={fieldErrors.providerType}
      />

      <Input
        label="Provider ID"
        value={providerId}
        onChange={(e) => setProviderId(e.target.value)}
      />

      <Select<(typeof capabilityStatusLabels)[number]>
        label="Status"
        options={capabilityStatusLabels}
        value={
          capabilityStatusLabels.find((l) => l.value === status) ?? undefined
        }
        getOptionLabel={(o) => o.label}
        getOptionValue={(o) => o.value}
        onChange={(option) =>
          setStatus(
            (option as (typeof capabilityStatusLabels)[number])
              .value as CapabilityResponse["status"],
          )
        }
        error={fieldErrors.status}
      />

      <Input
        label="Version"
        type="number"
        value={version}
        onChange={(e) => setVersion(e.target.value)}
        error={fieldErrors.version}
      />

      <Textarea
        label="Input schema (JSON)"
        rows={4}
        value={inputSchema}
        error={fieldErrors.inputSchema}
        onChange={(e) => setInputSchema(e.target.value)}
      />

      <Textarea
        label="Output schema (JSON)"
        rows={4}
        value={outputSchema}
        error={fieldErrors.outputSchema}
        onChange={(e) => setOutputSchema(e.target.value)}
      />

      <Input
        label="Credential reference"
        value={credentialReference}
        placeholder="ref:vault/my-secret (never the secret value)"
        onChange={(e) => setCredentialReference(e.target.value)}
      />

      <div className="flex justify-end gap-2 pt-2">
        <Button
          type="button"
          intent="secondary"
          onClick={onCancel}
          disabled={isPending}
        >
          Cancel
        </Button>
        <PermissionGate action="capability:update">
          <Button
            type="button"
            intent="primary"
            loading={isPending}
            onClick={() => void handleSave()}
          >
            Save
          </Button>
        </PermissionGate>
      </div>
    </div>
  )
}
