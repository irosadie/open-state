import { cva } from "class-variance-authority"

export const radioGroup = cva("", {
  variants: {
    intent: {
      primary:
        "rounded-full border-2 border-white checked:bg-primary-500 ring-2 ring-primary-500",
    },
    disabled: {
      false: "",
      true: "opacity-50 cursor-not-allowed",
    },
  },
  compoundVariants: [],
  defaultVariants: {
    intent: "primary",
    disabled: false,
  },
})

export const label = cva("", {
  variants: {
    intent: {
      main: "text-main-500",
      primary: "text-primary-500",
      danger: "text-danger-500",
      success: "text-success-500",
      warning: "text-warning-500",
      info: "text-info-500",
    },
  },
  defaultVariants: {
    intent: "main",
  },
})

export const position = cva("", {
  variants: {
    intent: {
      horizontal: "flex flex-row gap-x-8 gap-y-2 flex-wrap",
      vertical: "flex flex-col gap-y-2",
      "col-2": "grid grid-cols-2 gap-y-2",
      "col-3": "grid grid-cols-3 gap-y-2",
      "col-4": "grid grid-cols-4 gap-y-2",
      "col-5": "grid grid-cols-5 gap-y-2",
    },
  },
  defaultVariants: {
    intent: "horizontal",
  },
})
