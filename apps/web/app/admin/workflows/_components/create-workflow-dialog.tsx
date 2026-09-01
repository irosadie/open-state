"use client"

import { Button } from "$/components/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "$/components/dialog"
import { Input } from "$/components/input"
import { useWorkflowsCreate } from "$/hooks/transactions/use-workflow"
import { extractErrorMessage } from "$/utils/api-error"
import {
  type CreateWorkflowSchemaProps,
  createWorkflowSchema,
} from "@openstate/schemas"
import { useRouter } from "next/navigation"
import { type FormEvent, useState } from "react"

type CreateWorkflowDialogProps = {
  open: boolean
  onCancel: () => void
  projectId?: string
}

type FieldErrors = Partial<Record<keyof CreateWorkflowSchemaProps, string>>

const emptyForm: Pick<
  CreateWorkflowSchemaProps,
  "slug" | "name" | "description"
> = {
  slug: "",
  name: "",
  description: "",
}

export function CreateWorkflowDialog({
  open,
  onCancel,
  projectId,
}: CreateWorkflowDialogProps) {
  const router = useRouter()
  const { mutateAsync, isPending } = useWorkflowsCreate()
  const [form, setForm] = useState(emptyForm)
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [apiError, setApiError] = useState("")

  const handleChange = (field: keyof typeof emptyForm, value: string) => {
    setForm((prev) => ({ ...prev, [field]: value }))
    setFieldErrors((prev) => ({ ...prev, [field]: undefined }))
    setApiError("")
  }

  const handleCancel = () => {
    setForm(emptyForm)
    setFieldErrors({})
    setApiError("")
    onCancel()
  }

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()

    const parsed = createWorkflowSchema.safeParse({
      ...form,
      projectId,
      definition: {},
    })

    if (!parsed.success) {
      const errors: FieldErrors = {}

      for (const issue of parsed.error.issues) {
        const field = issue.path[0] as keyof CreateWorkflowSchemaProps

        if (!errors[field]) {
          errors[field] = issue.message
        }
      }

      setFieldErrors(errors)
      return
    }

    try {
      const data = await mutateAsync(parsed.data)
      router.push(
        `/state-builder/${data.id}${
          projectId ? `?projectId=${encodeURIComponent(projectId)}` : ""
        }`,
      )
    } catch (err) {
      setApiError(extractErrorMessage(err) ?? "Failed to create workflow.")
    }
  }

  return (
    <Dialog open={open} onOpenChange={(o) => !o && handleCancel()}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>New Workflow</DialogTitle>
        </DialogHeader>

        <form
          id="create-workflow-form"
          className="space-y-4"
          onSubmit={(e) => void handleSubmit(e)}
        >
          <Input
            label="Slug"
            name="slug"
            value={form.slug}
            error={fieldErrors.slug}
            onChange={(e) => handleChange("slug", e.target.value)}
            placeholder="my-workflow"
            required
          />
          <Input
            label="Name"
            name="name"
            value={form.name}
            error={fieldErrors.name}
            onChange={(e) => handleChange("name", e.target.value)}
            placeholder="My Workflow"
            required
          />
          <Input
            label="Description"
            name="description"
            value={form.description ?? ""}
            onChange={(e) => handleChange("description", e.target.value)}
            placeholder="Optional description"
          />

          {apiError ? (
            <p className="text-sm text-danger-500">{apiError}</p>
          ) : null}
        </form>

        <DialogFooter>
          <Button
            type="button"
            intent="secondary"
            onClick={handleCancel}
            disabled={isPending}
          >
            Cancel
          </Button>
          <Button type="submit" form="create-workflow-form" loading={isPending}>
            Create
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
