"use client"

import {
  type CurrencyFormatProps,
  currencyFormat as currencyFormatFn,
} from "@vibecoding-starter/utils/currency-format"
import type React from "react"
import { type ChangeEvent, type FC, Fragment, useEffect, useState } from "react"
import { Input, type InputProps } from "../input"

const cleanCurrencyInput = (value: string) => {
  return Number.parseFloat(value.toString().replace(/[^\d]/g, "")) || 0
}

export type CurrencyInputProps = Omit<
  InputProps & React.RefAttributes<HTMLInputElement>,
  "type"
> & { currencyFormat?: CurrencyFormatProps }

const CurrencyInput: FC<CurrencyInputProps> = (props) => {
  const { onChange, value, currencyFormat, ...rest } = props

  const [tempValue, setTempValue] = useState<string>()

  const handleOnChange = (e: ChangeEvent<HTMLInputElement>) => {
    const cleanedValue = cleanCurrencyInput(e.target.value)

    setTempValue(currencyFormatFn(cleanedValue, currencyFormat))
    const newTarget = {
      ...e.target,
      value: String(cleanedValue),
      name: e.target.name,
      id: e.target.id,
    }

    const tempEvent = {
      ...e,
      target: newTarget,
      currentTarget: newTarget,
    }

    onChange?.(tempEvent)
  }

  useEffect(() => {
    if (value) {
      const cleanedValue = cleanCurrencyInput(value as string)

      setTempValue(currencyFormatFn(cleanedValue, currencyFormat))
    }
  }, [currencyFormat, value])

  return (
    <Fragment>
      <Input {...rest} value={tempValue} onChange={handleOnChange} />
    </Fragment>
  )
}

export default CurrencyInput
