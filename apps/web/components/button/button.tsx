"use client"

import type { VariantProps } from "class-variance-authority"
import { Loader2 } from "lucide-react"
import type React from "react"
import type { ButtonHTMLAttributes } from "react"
import { button } from "./button.style"

export type ButtonProps = Omit<
  ButtonHTMLAttributes<HTMLButtonElement>,
  "disabled"
> &
  VariantProps<typeof button> & {
    leftIcon?: React.ReactNode
    rightIcon?: React.ReactNode
    loading?: boolean
  }

export const Button: React.FC<ButtonProps> = (props) => {
  const {
    intent,
    size,
    disabled,
    className,
    leftIcon,
    children,
    textOnly,
    rounded,
    fullWidth,
    rightIcon,
    bordered,
    uppercase,
    loading = false,
    ...rest
  } = props

  return (
    <button
      className={button({
        intent,
        size,
        disabled: disabled || loading,
        fullWidth,
        textOnly,
        rounded,
        bordered,
        uppercase,
        className,
      })}
      disabled={disabled || loading}
      {...rest}
    >
      {loading ? (
        <Loader2 className="h-4 w-4 animate-spin" />
      ) : (
        <>
          {leftIcon ?? null}
          {children}
          {rightIcon ?? null}
        </>
      )}
    </button>
  )
}

export default Button
