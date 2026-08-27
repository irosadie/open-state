import { cva } from "class-variance-authority"

export const autocomplete = cva(
  [
    "autocomplete transition flex gap-1.5 justify-center items-center flex-nowrap outline-none box-ring ring-0 border-none",
    "w-full",
  ],
  {
    variants: {
      intent: {
        primary: "text-black ring-[1px] ring-inset",
        secondary: "text-black ring-[1px] ring-inset",
        danger: "text-black ring-[1px] ring-inset",
        success: "text-black ring-[1px] ring-inset",
        warning: "text-black ring-[1px] ring-inset",
        info: "text-black ring-[1px] ring-inset",
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
    ],
    defaultVariants: {
      disabled: false,
      intent: "primary",
      size: "medium",
      rounded: "medium",
      hasError: false,
    },
  },
)

export const optionHover = cva("px-3 py-2 cursor-pointer", {
  variants: {
    intent: {
      primary: "hover:bg-primary-100",
      secondary: "hover:bg-secondary-100",
      danger: "hover:bg-danger-100",
      success: "hover:bg-success-100",
      warning: "hover:bg-warning-100",
      info: "hover:bg-info-100",
      clean: "hover:bg-gray-100",
    },
  },
  defaultVariants: {
    intent: "primary",
  },
})

export const bgFocus = cva("", {
  variants: {
    intent: {
      primary: "bg-primary-50",
      secondary: "bg-secondary-50",
      danger: "bg-danger-50",
      success: "bg-success-50",
      warning: "bg-warning-50",
      info: "bg-info-50",
      clean: "bg-gray-50",
    },
  },
  defaultVariants: {
    intent: "primary",
  },
})

export const label = cva("block text-sm font-medium mb-1", {
  variants: {
    intent: {
      main: "text-gray-700",
      primary: "text-primary-700",
      danger: "text-danger-700",
      success: "text-success-700",
      warning: "text-warning-700",
      info: "text-info-700",
    },
  },
  defaultVariants: {
    intent: "main",
  },
})
