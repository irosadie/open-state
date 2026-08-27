import { cva } from "class-variance-authority"

export const datePicker = cva(
  [
    "input transition flex gap-1.5 justify-center items-center flex-nowrap outline-none box-ring",
    "w-full",
  ],
  {
    variants: {
      intent: {
        primary: "text-black ring-[1px] ring-inset focus:bg-primary-50 ",
        secondary: "text-black ring-[1px] ring-inset focus:bg-secondary-50 ",
        danger: "text-black ring-[1px] ring-inset focus:bg-danger-50 ",
        success: "text-black ring-[1px] ring-inset focus:bg-success-50 ",
        warning: "text-black ring-[1px] ring-inset focus:bg-warning-50 ",
        info: "text-black ring-[1px] ring-inset focus:bg-info-50 ",
        clean:
          "text-black ring-[1px] ring-inset ring-gray-300 focus:bg-gray-50 bg-white",
      },
      size: {
        tiny: "text-xs py-1 px-2",
        small: "text-sm py-1.5 px-3",
        medium: "text-base py-2 px-4",
        large: "text-lg py-3 px-6",
        giant: "text-xl py-4 px-8",
      },
      disabled: {
        false: undefined,
        true: "opacity-50 cursor-not-allowed",
      },
      uppercase: {
        false: undefined,
        true: "uppercase",
      },
      rounded: {
        none: undefined,
        small: "rounded-sm",
        medium: "rounded-md",
        large: "rounded-lg",
        full: "rounded-full",
      },
      hasError: {
        false: undefined,
        true: "ring-danger-500",
      },
      hasLeftIcon: {
        true: "pl-10",
        false: undefined,
      },
      hasRightIcon: {
        true: "pr-10",
        false: undefined,
      },
    },
    compoundVariants: [
      {
        hasError: false,
        intent: "primary",
        class: "ring-primary-500",
      },
      {
        hasError: false,
        intent: "secondary",
        class: "ring-secondary-500",
      },
      {
        hasError: false,
        intent: "danger",
        class: "ring-danger-500",
      },
      {
        hasError: false,
        intent: "success",
        class: "ring-success-500",
      },
      {
        hasError: false,
        intent: "warning",
        class: "ring-warning-500",
      },
      {
        hasError: false,
        intent: "info",
        class: "ring-info-500",
      },
      {
        hasError: false,
        intent: "clean",
        class: "ring-gray-300",
      },
    ],
    defaultVariants: {
      disabled: false,
      intent: "primary",
      size: "medium",
      uppercase: false,
      rounded: "medium",
    },
  },
)

export const label = cva("text-sm font-medium", {
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
