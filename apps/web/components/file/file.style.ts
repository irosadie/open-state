import { cva } from "class-variance-authority"

export const label = cva("text-sm font-medium bg-white", {
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
