"use client"

import { type VariantProps, cx } from "class-variance-authority"
import { Fragment, type ReactNode, useCallback, useMemo } from "react"
import type {
  ActionMeta,
  ClassNamesConfig,
  GetOptionLabel,
  GroupBase,
  OptionsOrGroups,
  Props as ReactSelectProps,
} from "react-select"
import AsyncCreatableSelect from "react-select/async-creatable"
import CreatableSelect from "react-select/creatable"
import LoadingContent from "../content-loading/content-loading"
import {
  bgFocus,
  label as labelInput,
  optionHover,
  select as selectInput,
} from "./select.style"

// Global counter for generating unique instance IDs
let instanceCounter = 0

enum SelectType {
  ASYNC = "async",
  GENERAL = "general",
}

export type DefaultProps = Record<string, string>

export type SelectProps<T, M extends boolean = false> = Omit<
  ReactSelectProps,
  | "options"
  | "loadOptions"
  | "onChange"
  | "getOptionLabel"
  | "getOptionValue"
  | "value"
  | "isDisabled"
  | "isMulti"
> &
  Omit<VariantProps<typeof selectInput>, "hasError" | "hasLeftIcon"> & {
    isMulti?: M
    options?: T[]
    onChange?: (
      newValue: M extends true ? T[] : T,
      actionMeta: ActionMeta<M extends true ? T[] : T>,
    ) => void
    value?: M extends true ? Partial<T>[] : Partial<T>
    loadOptions?: (
      input: string,
      callback: (option: OptionsOrGroups<T, GroupBase<T>>) => void,
    ) => void
    getOptionLabel?: GetOptionLabel<T> | ((option: T) => ReactNode)
    getOptionValue?: GetOptionLabel<T> | ((option: T) => ReactNode)
    isAsync?: boolean
    leftIcon?: ReactNode
    disabled?: boolean
    label?: string
    error?: string
    hint?: string
    labelColor?: "main" | "primary" | "danger" | "success" | "warning" | "info"
    required?: boolean
    isLoading?: boolean
  }

const Select = <T = DefaultProps, M extends boolean = false>(
  props: SelectProps<T, M>,
) => {
  const {
    loadOptions,
    onChange,
    getOptionLabel,
    getOptionValue,
    options,
    value,
    isAsync,
    error,
    disabled,
    isMulti = false,
    intent,
    size,
    rounded,
    leftIcon,
    hint,
    required,
    label,
    labelColor,
    isLoading,
    menuIsOpen,
    onMenuClose,
    onMenuOpen,
    ...rest
  } = props

  // Generate a unique instance ID for this select component
  const instanceId = useMemo(() => {
    const id = `select-${instanceCounter++}`

    return id
  }, [])

  const selectType: SelectType = useMemo(() => {
    return isAsync ? SelectType.ASYNC : SelectType.GENERAL
  }, [isAsync])

  const handleChange = useCallback(
    (
      newValue: M extends true ? T[] : T,
      actionMeta: ActionMeta<M extends true ? T[] : T>,
    ) => {
      if (onChange) {
        onChange(newValue, actionMeta)
      }
    },
    [onChange],
  )

  const handleGetOptionLabel = useCallback(
    (option: unknown) => {
      if (getOptionLabel) {
        return getOptionLabel(option as T)
      }
      throw new Error("getOptionLabel is required")
    },
    [getOptionLabel],
  )

  const handleGetOptionValue = useCallback(
    (option: unknown) => {
      if (getOptionValue) {
        return getOptionValue(option as T)
      }
      return ""
    },
    [getOptionValue],
  )

  const selectClasses = useMemo(() => {
    return selectInput({
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
      return cx(selectClasses, additionalClasses)
    },
    menuList: () =>
      "bg-white shadow-lg text-sm overflow-y-auto max-h-[300px] pointer-events-auto",
    menu: () => "border border-gray-200 pointer-events-auto z-[99999]",
    option: () => optionHover({ intent }),
    multiValue: () => "bg-gray-200 rounded px-2 text-sm py-0.5",
    multiValueRemove: (props) => {
      if (props.isDisabled) {
        return "pointer-events-none opacity-50 cursor-not-allowed"
      }
      return "hover:bg-red-500 hover:text-white cursor-pointer"
    },
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
  }

  const select = (
    <CreatableSelect
      {...rest}
      value={value}
      options={options}
      onChange={
        handleChange as (
          newValue: unknown,
          actionMeta: ActionMeta<unknown>,
        ) => void
      }
      getOptionLabel={handleGetOptionLabel as GetOptionLabel<unknown>}
      getOptionValue={handleGetOptionValue as GetOptionLabel<unknown>}
      isDisabled={disabled}
      isMulti={isMulti}
      menuPosition="fixed"
      menuPlacement="auto"
      menuShouldScrollIntoView={false}
      menuShouldBlockScroll={false}
      menuPortalTarget={typeof window !== "undefined" ? document.body : null}
      inputId={instanceId}
      menuIsOpen={menuIsOpen}
      onMenuClose={onMenuClose}
      onMenuOpen={onMenuOpen}
      unstyled
      classNames={classNames}
      instanceId={instanceId}
      styles={{
        menuPortal: (base) => ({ ...base, zIndex: 9999 }),
        menu: (base) => ({ ...base, zIndex: 9999 }),
      }}
    />
  )

  const asyncCreatableSelect = (
    <AsyncCreatableSelect
      {...rest}
      value={value}
      loadOptions={loadOptions}
      defaultOptions
      onChange={
        handleChange as (
          newValue: unknown,
          actionMeta: ActionMeta<unknown>,
        ) => void
      }
      getOptionLabel={handleGetOptionLabel as GetOptionLabel<unknown>}
      getOptionValue={handleGetOptionValue as GetOptionLabel<unknown>}
      isDisabled={disabled}
      isMulti={isMulti}
      menuPosition="fixed"
      menuPlacement="auto"
      menuShouldScrollIntoView={false}
      menuShouldBlockScroll={false}
      menuPortalTarget={typeof window !== "undefined" ? document.body : null}
      inputId={instanceId}
      menuIsOpen={menuIsOpen}
      onMenuClose={onMenuClose}
      onMenuOpen={onMenuOpen}
      unstyled
      classNames={classNames}
      instanceId={instanceId}
      styles={{
        menuPortal: (base) => ({ ...base, zIndex: 9999 }),
        menu: (base) => ({ ...base, zIndex: 9999 }),
      }}
    />
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
          {
            {
              MOUNTED: (
                <Fragment>
                  {selectType === SelectType.ASYNC
                    ? asyncCreatableSelect
                    : select}
                </Fragment>
              ),
              LOADING: <LoadingContent lineNumber={1} />,
            }[isLoading ? "LOADING" : "MOUNTED"]
          }
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

export default Select
