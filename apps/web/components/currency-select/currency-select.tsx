import type { FC } from "react"
import { Select, type SelectProps } from "../select"
import currencies from "./currencies.json"

const currencySeparator = "\u00a0\u00a0\u00a0"

export type CurrencySelectProps = Omit<
  SelectProps<{ label: string; value: string }>,
  "options" | "value" | "getOptionLabel"
> & {
  value?: string
}

const options = Object.keys(currencies).map((code) => ({
  label: `${currencies[code as keyof typeof currencies]}${currencySeparator}${code}`,
  value: code,
}))

const CurrencySelect: FC<CurrencySelectProps> = (props) => {
  const { value, ...rest } = props

  const tempValue = value
    ? {
        value,
        label: `${currencies[value as keyof typeof currencies]}${currencySeparator}${value}`,
      }
    : undefined

  return (
    <Select<{ label: string; value: string }>
      {...rest}
      options={options}
      value={tempValue}
      getOptionLabel={(v) => v.label}
      formatOptionLabel={(v) => {
        const newValue = v as unknown as { label: string; value: string }

        return <span>{newValue.label}</span>
      }}
    />
  )
}

export default CurrencySelect
