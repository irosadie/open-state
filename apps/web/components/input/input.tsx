"use client"

import type { VariantProps } from "class-variance-authority"
import { type FC, type InputHTMLAttributes, type ReactNode, useId } from "react"
import { input, label as labelInput } from "./input.style"

export type InputProps = Omit<
  InputHTMLAttributes<HTMLInputElement>,
  "disabled" | "size"
> &
  Omit<
    VariantProps<typeof input>,
    "hasError" | "hasLeftIcon" | "hasRightIcon"
  > & {
    label?: string
    error?: string
    hint?: string
    labelColor?: "main" | "primary" | "danger" | "success" | "warning" | "info"
    leftIcon?: ReactNode
    rightIcon?: ReactNode
    required?: boolean
  }

export const Input: FC<InputProps> = (props) => {
  const {
    className,
    intent,
    size,
    disabled,
    label,
    error,
    hint,
    labelColor,
    leftIcon,
    rightIcon,
    uppercase,
    required,
    rounded,
    ...rest
  } = props
  const inputId = rest.id ?? useId()

  return (
    <div className="space-y-1.5">
      {label && (
        <label htmlFor={inputId} className={labelInput({ intent: labelColor })}>
          {label}
          {required && <span className="text-danger-500">*</span>}
        </label>
      )}
      <div className="space-y-1">
        <div className="relative">
          <input
            id={inputId}
            className={input({
              intent,
              size,
              disabled,
              hasError: !!error,
              hasLeftIcon: !!leftIcon,
              hasRightIcon: !!rightIcon,
              uppercase,
              rounded,
              className,
            })}
            disabled={disabled || undefined}
            {...rest}
          />
          {leftIcon && (
            <div className="absolute inset-y-0 left-0 flex items-center pl-2.5 pointer-events-none w-full">
              {leftIcon}
            </div>
          )}
          {rightIcon && (
            <div className="absolute inset-y-0 right-0 flex items-center pr-2.5 pointer-events-none h-full">
              {rightIcon}
            </div>
          )}
        </div>
        {hint && !error && (
          <span className="text-xs text-main-300">{hint}</span>
        )}
        {error && <span className="text-xs text-danger-500">{error}</span>}
      </div>
    </div>
  )
}

export default Input
