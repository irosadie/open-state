"use client"

import type { VariantProps } from "class-variance-authority"
import type React from "react"
import type { FC, InputHTMLAttributes } from "react"
import { label as labelInput, radio } from "./radio.style"

export type RadioProps = Omit<
  InputHTMLAttributes<HTMLInputElement>,
  "disabled"
> &
  VariantProps<typeof radio> & {
    label?: string
    error?: string
    uppercase?: boolean
    name: string
  }

export const Radio: FC<RadioProps> = (props) => {
  const {
    className,
    intent,
    disabled,
    label,
    name,
    uppercase,
    id = Math.random().toString(36).substring(7),
    onChange,
    ...rest
  } = props

  const handleOnChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (disabled) {
      e.preventDefault()
      e.stopPropagation()
      return
    }
    onChange?.(e)
  }

  return (
    <div className="flex gap-2 items-center justify-start">
      <input
        id={id}
        name={name}
        type="radio"
        className={radio({ intent, disabled, className })}
        {...rest}
        onChange={handleOnChange}
      />
      {label && (
        <label
          htmlFor={id}
          className={labelInput({ uppercase, disabled, className })}
        >
          {label}
        </label>
      )}
    </div>
  )
}

export default Radio
