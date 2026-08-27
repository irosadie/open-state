"use client"
import "react-date-picker/dist/DatePicker.css"
import "react-calendar/dist/Calendar.css"
import "./custom-datepicker.css"
import type { VariantProps } from "class-variance-authority"
import { type FC, Fragment, type ReactNode, useId } from "react"
import DatePickerComponent, {
  type DatePickerProps as LibDatePickerProps,
} from "react-date-picker"
import { datePicker, label as labelInput } from "./datepicker.style"

export type DatePickerProps = Omit<
  LibDatePickerProps,
  "disabled" | "size" | "calendarIcon"
> &
  Omit<
    VariantProps<typeof datePicker>,
    "hasError" | "hasLeftIcon" | "hasRightIcon"
  > & {
    label?: string
    error?: string
    hint?: string
    labelColor?: "main" | "primary" | "danger" | "success" | "warning" | "info"
    leftIcon?: ReactNode
    rightIcon?: ReactNode
    required?: boolean
    openByIcon?: boolean
  }

const DatePicker: FC<DatePickerProps> = (props) => {
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
    openByIcon,
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
          <DatePickerComponent
            disabled={disabled || undefined}
            clearIcon={null}
            id={inputId}
            className={datePicker({
              intent,
              size,
              disabled,
              hasError: !!error,
              hasLeftIcon: !!leftIcon && !openByIcon,
              hasRightIcon: !!rightIcon,
              uppercase,
              rounded,
              className,
            })}
            calendarIcon={
              <Fragment>
                {
                  {
                    ICON: leftIcon,
                    NONE: null,
                  }[leftIcon && openByIcon ? "ICON" : "NONE"]
                }
              </Fragment>
            }
            {...rest}
          />
          {leftIcon && !openByIcon && (
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
      </div>
      {hint && !error && <span className="text-xs text-main-300">{hint}</span>}
      {error && <span className="text-xs text-danger-500">{error}</span>}
    </div>
  )
}

export default DatePicker
