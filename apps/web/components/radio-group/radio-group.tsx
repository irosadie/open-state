"use client"

import type { VariantProps } from "class-variance-authority"
import type React from "react"
import { useCallback, useState } from "react"
import {
  Radio as RadioComponent,
  type RadioProps as RadioComponentProps,
} from "./../radio"
import {
  label as labelRadio,
  position as positionRadio,
  type radioGroup,
} from "./radio-group.style"

export type DefaultProps = Record<string, unknown>

export type RadioGroupProps<T> = VariantProps<typeof radioGroup> &
  Omit<
    RadioComponentProps & React.RefAttributes<HTMLInputElement>,
    "error" | "value" | "label" | "checked" | "defaultChecked"
  > & {
    data: T[]
    error?: string
    checkedValue?: unknown
    getDataLabel: (option: T) => unknown
    getDataValue: (option: T) => unknown
    label?: string
    hint?: string
    labelColor?: "main" | "primary" | "danger" | "success" | "warning" | "info"
    required?: boolean
    uppercase?: boolean
    position?: "horizontal" | "vertical" | "col-2" | "col-3" | "col-4" | "col-5"
  }

const RadioGroup = <T = DefaultProps>(props: RadioGroupProps<T>) => {
  const {
    error,
    data,
    checkedValue,
    label,
    hint,
    labelColor,
    required,
    position,
    getDataLabel,
    getDataValue,
    onChange,
    ...rest
  } = props

  const handleGetDataLabel = useCallback(
    (option: unknown) => {
      if (getDataLabel) {
        return getDataLabel(option as T)
      }
      throw new Error("getDataLabel is required")
    },
    [getDataLabel],
  )

  const handleGetDataValue = useCallback(
    (option: unknown) => {
      if (getDataValue) {
        return getDataValue(option as T)
      }
      throw new Error("getDataValue is required")
    },
    [getDataValue],
  ) as (option: T) => string

  const [tempChecked, setTempChecked] = useState(
    checkedValue ?? handleGetDataValue(data[0] as T),
  )

  const handleOnChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setTempChecked(e.target.value)
    onChange?.(e)
  }

  const renderRadio = (data || []).map((option, index) => {
    const label = handleGetDataLabel(option) as string | undefined
    const value = handleGetDataValue(option) as string | undefined

    return (
      <div key={value ?? `option-${index + 1}`}>
        <RadioComponent
          value={value}
          label={label}
          checked={value === tempChecked}
          onChange={handleOnChange}
          {...rest}
        />
      </div>
    )
  })

  return (
    <fieldset className="space-y-1.5 text-sm font-medium flex flex-col">
      {label && (
        <legend className={labelRadio({ intent: labelColor })}>
          {label}
          {required && <span className="text-danger-500">*</span>}
        </legend>
      )}
      <div className="space-y-1">
        <ul className={positionRadio({ intent: position })}>{renderRadio}</ul>
        {hint && !error && (
          <span className="text-xs text-main-300">{hint}</span>
        )}
        {error && <span className="text-xs text-danger-500">{error}</span>}
      </div>
    </fieldset>
  )
}

export default RadioGroup
