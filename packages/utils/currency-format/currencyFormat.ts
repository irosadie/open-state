export type CurrencyFormatProps = {
  locale?: string
  currency?: string
  minimumFractionDigits?: number
  maximumFractionDigits?: number
}

const currencyFormat = (
  amount: number,
  {
    locale = "id-ID",
    currency = "IDR",
    minimumFractionDigits = 0,
    maximumFractionDigits = 0,
  }: CurrencyFormatProps = {},
) => {
  return new Intl.NumberFormat(locale, {
    style: "currency",
    currency,
    minimumFractionDigits,
    maximumFractionDigits,
  }).format(amount)
}

export default currencyFormat
