"use client"

import type { VariantProps } from "class-variance-authority"
import { type FC, type TextareaHTMLAttributes, useId } from "react"
import { label as labelInput, textarea } from "./textarea.style"

export type TextareaProps = Omit<
  TextareaHTMLAttributes<HTMLTextAreaElement>,
  "disabled" | "size"
> &
  Omit<VariantProps<typeof textarea>, "hasError"> & {
    label?: string
    error?: string
    hint?: string
    labelColor?: "main" | "primary" | "danger" | "success" | "warning" | "info"
    required?: boolean
  }

export const Textarea: FC<TextareaProps> = (props) => {
  const {
    className,
    intent,
    size,
    disabled,
    label,
    error,
    hint,
    labelColor,
    uppercase,
    required,
    rounded,
    ...rest
  } = props
  const textareaId = rest.id ?? useId()

  return (
    <div className="space-y-1.5">
      {label && (
        <label
          htmlFor={textareaId}
          className={labelInput({ intent: labelColor })}
        >
          {label}
          {required && <span className="text-danger-500">*</span>}
        </label>
      )}
      <div className="space-y-1">
        <div className="relative">
          <textarea
            id={textareaId}
            className={textarea({
              intent,
              size,
              disabled,
              hasError: !!error,
              uppercase,
              rounded,
              className,
            })}
            disabled={disabled || undefined}
            {...rest}
          />
        </div>
        {hint && !error && (
          <span className="text-xs text-main-300">{hint}</span>
        )}
        {error && <span className="text-xs text-danger-500">{error}</span>}
      </div>
    </div>
  )
}

export default Textarea
