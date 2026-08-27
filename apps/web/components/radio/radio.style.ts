import { cva } from "class-variance-authority"

export const radio = cva(
  "appearance-none w-3 h-3 transition-all duration-300 ease-out",
  {
    variants: {
      intent: {
        primary:
          "rounded-full border-2 border-white checked:bg-primary-500 ring-2 ring-primary-500",
        secondary:
          "rounded-full border-2 border-white checked:bg-secondary-500 ring-2 ring-secondary-500",
        danger:
          "rounded-full border-2 border-white checked:bg-danger-500 ring-2 ring-danger-500",
        success:
          "rounded-full border-2 border-white checked:bg-success-500 ring-2 ring-success-500",
        warning:
          "rounded-full border-2 border-white checked:bg-warning-500 ring-2 ring-warning-500",
        info: "rounded-full border-2 border-white checked:bg-info-500 ring-2 ring-info-500",
      },
      disabled: {
        false: undefined,
        true: "opacity-50 cursor-not-allowed",
      },
    },
    compoundVariants: [],
    defaultVariants: {
      intent: "primary",
      disabled: false,
    },
  },
)

export const label = cva("text-sm text-main-500", {
  variants: {
    uppercase: {
      true: "uppercase",
      false: undefined,
    },
    disabled: {
      true: "opacity-90 cursor-not-allowed",
      false: undefined,
    },
  },
  defaultVariants: {
    uppercase: false,
    disabled: false,
  },
})
