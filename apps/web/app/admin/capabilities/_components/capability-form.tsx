"use client"

import { useState } from "react"

import { PermissionGate } from "$/components/auth-guard/permission-gate"
import { Button } from "$/components/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "$/components/dialog"
import { Input } from "$/components/input"
import { Select } from "$/components/select"
import { Textarea } from "$/components/textarea"
import {
  type CreateCapabilitySchemaProps,
  createCapabilitySchema,
  providerTypeLabels,
} from "@openstate/schemas"
import type { ZodError } from "zod"

type CapabilityFormDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  isPending: boolean
  onSubmit: (payload: CreateCapabilitySchemaProps) => Promise<void>
}

type FormState = {
  name: string
  description: string
  providerType?: CreateCapabilitySchemaProps["providerType"]
  providerId: string
  inputSchema: string
  outputSchema: string
  credentialReference: string
  version: string
}

const emptyForm: FormState = {
  name: "",
  description: "",
  providerType: undefined,
  providerId: "",
  inputSchema: "",
  outputSchema: "",
  credentialReference: "",
  version: "1",
}

type FieldErrors = Partial<Record<keyof CreateCapabilitySchemaProps, string>>

const parseJsonField = (raw: string): Record<string, unknown> | undefined => {
  if (!raw.trim()) return undefined

  try {
    return JSON.parse(raw) as Record<string, unknown>
  } catch {
    return undefined
  }
}

export function CapabilityFormDialog({
  open,
  onOpenChange,
  isPending,
  onSubmit,
}: CapabilityFormDialogProps) {
  const [form, setForm] = useState<FormState>(emptyForm)
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})

  const handleClose = () => {
    setForm(emptyForm)
    setFieldErrors({})
    onOpenChange(false)
  }

  const update = <K extends keyof FormState>(key: K, value: FormState[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }))
  }

  const handleSubmit = async () => {
    const inputSchema = parseJsonField(form.inputSchema)
    const outputSchema = parseJsonField(form.outputSchema)

    if (form.inputSchema.trim() && inputSchema === undefined) {
      setFieldErrors({ inputSchema: "Input schema must be valid JSON" })

      return
    }

    if (form.outputSchema.trim() && outputSchema === undefined) {
      setFieldErrors({ outputSchema: "Output schema must be valid JSON" })

      return
    }

    const parsed = createCapabilitySchema.safeParse({
      name: form.name,
      description: form.description || undefined,
      providerType: form.providerType,
      providerId: form.providerId || undefined,
      inputSchema,
      outputSchema,
      version: form.version ? Number(form.version) : undefined,
      credentialReference: form.credentialReference || undefined,
    })

    if (!parsed.success) {
      const errors: FieldErrors = {}

      for (const issue of (parsed.error as ZodError).issues) {
        const key = issue.path[0] as keyof CreateCapabilitySchemaProps

        if (key) {
          errors[key] = issue.message
        }
      }

      setFieldErrors(errors)

      return
    }

    await onSubmit(parsed.data)
    handleClose()
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Register Capability</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          <Input
            label="Name"
            required
            placeholder="payment.create"
            value={form.name}
            error={fieldErrors.name}
            onChange={(e) => update("name", e.target.value)}
          />

          <Input
            label="Description"
            placeholder="Optional description"
            value={form.description}
            onChange={(e) => update("description", e.target.value)}
          />

          <Select<(typeof providerTypeLabels)[number]>
            label="Provider type"
            required
            options={providerTypeLabels}
            getOptionLabel={(o) => o.label}
            getOptionValue={(o) => o.value}
            onChange={(option) =>
              update(
                "providerType",
                (option as (typeof providerTypeLabels)[number])
                  .value as CreateCapabilitySchemaProps["providerType"],
              )
            }
            error={fieldErrors.providerType}
          />

          <Input
            label="Provider ID"
            placeholder="Optional provider id"
            value={form.providerId}
            onChange={(e) => update("providerId", e.target.value)}
          />

          <Textarea
            label="Input schema (JSON)"
            placeholder='{"type":"object"}'
            rows={3}
            value={form.inputSchema}
            error={fieldErrors.inputSchema}
            onChange={(e) => update("inputSchema", e.target.value)}
          />

          <Textarea
            label="Output schema (JSON)"
            placeholder='{"type":"object"}'
            rows={3}
            value={form.outputSchema}
            error={fieldErrors.outputSchema}
            onChange={(e) => update("outputSchema", e.target.value)}
          />

          <Input
            label="Credential reference"
            placeholder="ref:vault/my-secret (never the secret value)"
            value={form.credentialReference}
            onChange={(e) => update("credentialReference", e.target.value)}
          />

          <Input
            label="Version"
            type="number"
            value={form.version}
            error={fieldErrors.version}
            onChange={(e) => update("version", e.target.value)}
          />

          <DialogFooter>
            <Button
              type="button"
              intent="secondary"
              onClick={handleClose}
              disabled={isPending}
            >
              Cancel
            </Button>
            <PermissionGate action="capability:create">
              <Button
                type="button"
                intent="primary"
                loading={isPending}
                onClick={() => void handleSubmit()}
              >
                Create
              </Button>
            </PermissionGate>
          </DialogFooter>
        </div>
      </DialogContent>
    </Dialog>
  )
}
