import currencies from "./currencies.json"

const currencySeparator = "\u00a0\u00a0"

const options = Object.keys(currencies).map((code) => ({
  label: `${currencies[code as keyof typeof currencies]}${currencySeparator}${code}`,
  value: code,
}))

export type CurrencyDisplayProps = {
  value: string
}

const CurrencyDisplay = (props: CurrencyDisplayProps) => {
  const { value } = props

  const currencyDisplay = options.find((v) => v.value === value)?.label || ""

  return <span>{currencyDisplay}</span>
}

export default CurrencyDisplay
