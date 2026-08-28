"use client"

import { Button } from "$/components/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "$/components/dialog"

type CapabilityConfirmDialogProps = {
  open: boolean
  isPending?: boolean
  title: string
  description: string
  confirmLabel: string
  onCancel: () => void
  onConfirm: () => void
}

export function CapabilityConfirmDialog({
  open,
  isPending = false,
  title,
  description,
  confirmLabel,
  onCancel,
  onConfirm,
}: CapabilityConfirmDialogProps) {
  return (
    <Dialog open={open} onOpenChange={(o) => !o && onCancel()}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
        </DialogHeader>

        <p className="text-sm text-gray-600">{description}</p>

        <DialogFooter>
          <Button
            type="button"
            intent="secondary"
            onClick={onCancel}
            disabled={isPending}
          >
            Cancel
          </Button>
          <Button
            type="button"
            intent="danger"
            loading={isPending}
            onClick={onConfirm}
          >
            {confirmLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
