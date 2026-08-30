"use client"

import { useState } from "react"

import { PermissionGate } from "$/components/auth-guard/permission-gate"
import { Button } from "$/components/button"
import { PanelCard } from "$/components/panel-card"
import { Textarea } from "$/components/textarea"
import { FlaskConicalIcon } from "lucide-react"

import { useCapabilitiesTest } from "$/hooks/transactions/use-capability"
import { extractClassifiedError, extractErrorMessage } from "$/utils/api-error"
import type { CapabilityInvocationResultResponse } from "@openstate/types"

type TestInvocationPanelProps = {
  capabilityId: string
  inputSchema?: Record<string, unknown> | null
}

export function TestInvocationPanel({
  capabilityId,
  inputSchema,
}: TestInvocationPanelProps) {
  const [payloadText, setPayloadText] = useState(
    JSON.stringify(buildSamplePayload(inputSchema), null, 2),
  )
  const [result, setResult] =
    useState<CapabilityInvocationResultResponse | null>(null)
  const [failure, setFailure] = useState<string | null>(null)
  const [classified, setClassified] = useState<{
    kind?: string
    code?: string
  } | null>(null)

  const { mutateAsync: testMutateAsync, isPending } = useCapabilitiesTest()

  const handleTest = async () => {
    let payload: Record<string, unknown>

    try {
      payload = JSON.parse(payloadText) as Record<string, unknown>
    } catch {
      setFailure("Payload must be valid JSON")

      return
    }

    setFailure(null)
    setResult(null)
    setClassified(null)

    await testMutateAsync(
      { capabilityId, payload: { payload } },
      {
        onSuccess: (data) => {
          setResult(data)
        },
        onError: (error) => {
          const classifiedErr = extractClassifiedError(error)

          if (classifiedErr) {
            setClassified({
              kind: classifiedErr.kind,
              code: classifiedErr.code,
            })
          }

          setFailure(extractErrorMessage(error) || "Test invocation failed")
        },
      },
    )
  }

  return (
    <PanelCard
      title="Test invocation"
      description="Run this capability through the mock/sandbox provider (PRD §2064). No live external call is made."
      action={
        <span className="inline-flex items-center gap-1 rounded-full bg-amber-100 px-2.5 py-0.5 text-xs font-medium text-amber-800">
          <FlaskConicalIcon size={12} /> Sandbox / mock
        </span>
      }
    >
      <div className="space-y-4 px-6 py-4">
        <Textarea
          label="Invocation payload (JSON)"
          rows={6}
          value={payloadText}
          onChange={(e) => setPayloadText(e.target.value)}
        />

        <div className="flex items-center gap-2">
          <PermissionGate action="capability:invoke">
            <Button
              intent="primary"
              loading={isPending}
              onClick={() => void handleTest()}
            >
              Run test
            </Button>
          </PermissionGate>
          <span className="text-xs text-gray-400">
            Executes via the mock provider only — no live side effects.
          </span>
        </div>

        {failure ? (
          <div className="rounded-md bg-red-50 px-4 py-3 text-sm text-red-700">
            {classified ? (
              <div className="mb-1 flex items-center gap-2">
                {classified.kind ? (
                  <span className="rounded bg-red-100 px-1.5 py-0.5 text-xs font-semibold uppercase text-red-800">
                    {classified.kind}
                  </span>
                ) : null}
                {classified.code ? (
                  <code className="text-xs text-red-800">
                    {classified.code}
                  </code>
                ) : null}
              </div>
            ) : null}
            {failure}
          </div>
        ) : null}

        {result ? (
          <div className="rounded-md border border-gray-200 bg-gray-50 p-4">
            <div className="mb-2 flex items-center gap-2">
              <span className="inline-flex items-center rounded-full bg-amber-100 px-2 py-0.5 text-xs font-medium text-amber-800">
                fromMock
              </span>
              {result.durationMs !== undefined ? (
                <span className="text-xs text-gray-500">
                  {result.durationMs} ms
                </span>
              ) : null}
              {result.event ? (
                <span className="font-mono text-xs text-gray-500">
                  {result.event}
                </span>
              ) : null}
            </div>
            <pre className="max-h-64 overflow-auto rounded bg-white p-3 text-xs text-gray-700">
              {JSON.stringify(result.data ?? {}, null, 2)}
            </pre>
          </div>
        ) : null}
      </div>
    </PanelCard>
  )
}

const buildSamplePayload = (schema?: Record<string, unknown> | null) => {
  const properties =
    (schema?.properties as Record<string, unknown> | undefined) ?? {}

  const sample: Record<string, unknown> = {}

  for (const [key] of Object.entries(properties)) {
    sample[key] = `__${key}__`
  }

  return sample
}
