"use client"

import { type VariantProps, cx } from "class-variance-authority"
import { type ReactNode, useCallback, useMemo } from "react"
import type {
  ActionMeta,
  ClassNamesConfig,
  GetOptionLabel,
  GroupBase,
  OptionsOrGroups,
} from "react-select"
import AsyncSelect from "react-select/async"
import {
  autocomplete as autocompleteInput,
  bgFocus,
  label as labelInput,
  optionHover,
} from "./autocomplete.style"

// Global counter for generating unique instance IDs
let instanceCounter = 0

export type DefaultProps = Record<string, string>

export type AutocompleteProps<T> = Omit<
  VariantProps<typeof autocompleteInput>,
  "hasError"
> & {
  // Data from parent (fetched using hook)
  options: T[]
  isLoading?: boolean

  // Value & onChange
  value?: T | null
  onChange: (newValue: T | null) => void

  // Label & Value functions
  getOptionLabel: (option: T) => ReactNode
  getOptionValue: (option: T) => string

  // Search callback to parent
  onInputChange?: (search: string) => void

  // UI props
  leftIcon?: ReactNode
  disabled?: boolean
  label?: string
  error?: string
  hint?: string
  labelColor?: "main" | "primary" | "danger" | "success" | "warning" | "info"
  required?: boolean
  placeholder?: string
  isClearable?: boolean
  noOptionsMessage?: string
}

const Autocomplete = <T = DefaultProps>(props: AutocompleteProps<T>) => {
  const {
    options,
    isLoading,
    onChange,
    onInputChange,
    getOptionLabel,
    getOptionValue,
    value,
    error,
    disabled,
    intent,
    size,
    rounded,
    leftIcon,
    hint,
    required,
    label,
    labelColor,
    placeholder = "Search...",
    isClearable = true,
    noOptionsMessage = "No options",
    ...rest
  } = props

  // Generate a unique instance ID
  const instanceId = useMemo(() => {
    const id = `autocomplete-${instanceCounter++}`

    return id
  }, [])

  const handleChange = useCallback(
    (newValue: T | null, _actionMeta: ActionMeta<T>) => {
      onChange(newValue)
    },
    [onChange],
  )

  const handleInputChange = useCallback(
    (newValue: string) => {
      if (onInputChange) {
        onInputChange(newValue)
      }
      return newValue
    },
    [onInputChange],
  )

  const handleGetOptionLabel = useCallback(
    (option: unknown) => {
      return getOptionLabel(option as T)
    },
    [getOptionLabel],
  )

  const handleGetOptionValue = useCallback(
    (option: unknown) => {
      return getOptionValue(option as T)
    },
    [getOptionValue],
  )

  const autocompleteClasses = useMemo(() => {
    return autocompleteInput({
      intent,
      size,
      rounded,
      hasError: !!error,
      disabled,
    })
  }, [intent, size, rounded, error, disabled])

  const classNames: ClassNamesConfig<unknown, boolean, GroupBase<unknown>> = {
    control: (props) => {
      let additionalClasses = "flex items-start relative"

      if (props.isFocused) {
        additionalClasses = cx(additionalClasses, bgFocus({ intent }))
      }
      return cx(autocompleteClasses, additionalClasses)
    },
    menuList: () =>
      "bg-white shadow-lg text-sm overflow-y-auto max-h-[300px] pointer-events-auto",
    menu: () => "border border-gray-200 pointer-events-auto z-[99999]",
    option: () => optionHover({ intent }),
    valueContainer: (props) => {
      let mainClass = "gap-2 pr-16"

      if (leftIcon) {
        mainClass = "pl-6 gap-2 pr-16"
      }
      if (props.isDisabled) {
        return cx(mainClass, "pointer-events-auto cursor-not-allowed")
      }
      return mainClass
    },
    menuPortal: () => "z-[9999] pointer-events-auto",
    indicatorsContainer: () => "flex items-start",
    noOptionsMessage: () => "text-gray-500 text-sm py-2 px-3",
    loadingMessage: () => "text-gray-500 text-sm py-2 px-3",
    loadingIndicator: () => "text-gray-400",
  }

  const loadOptions = useCallback(
    async (_inputValue: string): Promise<OptionsOrGroups<T, GroupBase<T>>> => {
      // Just return current options, filtering is done via API
      return options as OptionsOrGroups<T, GroupBase<T>>
    },
    [options],
  )

  return (
    <div className="space-y-1.5">
      {label && (
        <label
          htmlFor={instanceId}
          className={labelInput({ intent: labelColor })}
        >
          {label}
          {required && <span className="text-danger-500">*</span>}
        </label>
      )}
      <div className="space-y-1">
        <div className="relative">
          <AsyncSelect
            {...rest}
            value={value}
            loadOptions={loadOptions}
            defaultOptions={options}
            cacheOptions={false}
            onInputChange={handleInputChange}
            onChange={
              handleChange as (
                newValue: unknown,
                actionMeta: ActionMeta<unknown>,
              ) => void
            }
            getOptionLabel={handleGetOptionLabel as GetOptionLabel<unknown>}
            getOptionValue={handleGetOptionValue as GetOptionLabel<unknown>}
            isDisabled={disabled}
            isLoading={isLoading}
            isClearable={isClearable}
            placeholder={placeholder}
            noOptionsMessage={() => noOptionsMessage}
            loadingMessage={() => "Loading..."}
            menuPosition="fixed"
            menuPlacement="auto"
            menuShouldScrollIntoView={false}
            menuShouldBlockScroll={false}
            menuPortalTarget={
              typeof window !== "undefined" ? document.body : null
            }
            inputId={instanceId}
            unstyled
            classNames={classNames}
            instanceId={instanceId}
            styles={{
              menuPortal: (base) => ({ ...base, zIndex: 9999 }),
              menu: (base) => ({ ...base, zIndex: 9999 }),
            }}
          />
          {leftIcon && (
            <div className="absolute inset-y-0 left-0 flex items-center pl-2.5 pointer-events-none w-full">
              {leftIcon}
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

export default Autocomplete
