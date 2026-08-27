import { cva } from "class-variance-authority"

export const theadBg = cva("", {
  variants: {
    intent: {
      primary: "bg-primary-100",
      secondary: "bg-secondary-100",
      danger: "bg-danger-100",
      success: "bg-success-100",
      warning: "bg-warning-100",
      info: "bg-info-100",
      clean: "bg-gray-50",
    },
  },
  defaultVariants: {
    intent: "primary",
  },
})

export const theadRowBorder = cva("p-2 text-left text-white", {
  variants: {
    intent: {
      primary: "border-b border-primary-300",
      secondary: "border-b border-secondary-300",
      danger: "border-b border-danger-300",
      success: "border-b border-success-300",
      warning: "border-b border-warning-300",
      info: "border-b border-info-300",
      clean: "border-b border-gray-200",
    },
  },
  defaultVariants: {
    intent: "primary",
  },
})

export const thBg = cva("p-2 text-left", {
  variants: {
    intent: {
      primary: "bg-primary-500 text-white",
      secondary: "bg-secondary-500 text-white",
      danger: "bg-danger-500 text-white",
      success: "bg-success-500 text-white",
      warning: "bg-warning-500 text-white",
      info: "bg-info-500 text-white",
      clean:
        "bg-gray-50 text-gray-500 text-xs font-medium uppercase tracking-wider px-6 py-3",
    },
  },
  defaultVariants: {
    intent: "primary",
  },
})

export const trow = cva(
  "transition duration-300 even:bg-white hover:even:bg-neutral",
  {
    variants: {
      intent: {
        primary: "odd:bg-primary-50 hover:odd:bg-primary-100",
        secondary: "odd:bg-secondary-50 hover:odd:bg-secondary-100",
        danger: "odd:bg-danger-50 hover:odd:bg-danger-100",
        success: "odd:bg-success-50 hover:odd:bg-success-100",
        warning: "odd:bg-warning-50 hover:odd:bg-warning-100",
        info: "odd:bg-info-50 hover:odd:bg-info-100",
        clean:
          "bg-white hover:bg-gray-50 transition-colors odd:bg-white even:bg-gray-100 hover:odd:bg-gray-50 hover:even:bg-gray-100/50",
      },
      hasExpandedRow: {
        true: "!bg-main-50 cursor-pointer",
        false: undefined,
      },
    },
    defaultVariants: {
      intent: "primary",
      hasExpandedRow: false,
    },
  },
)

export const expandedRowBorder = cva("", {
  variants: {
    intent: {
      primary: "border-b border-primary-300",
      secondary: "border-b border-secondary-300",
      danger: "border-b border-danger-300",
      success: "border-b border-success-300",
      warning: "border-b border-warning-300",
      info: "border-b border-info-300",
      clean: "border-b border-gray-200",
    },
  },
  defaultVariants: {
    intent: "primary",
  },
})

export const tbodyBg = cva("", {
  variants: {
    intent: {
      primary: "bg-white",
      secondary: "bg-white",
      danger: "bg-white",
      success: "bg-white",
      warning: "bg-white",
      info: "bg-white",
      clean: "bg-white divide-y divide-gray-200",
    },
  },
  defaultVariants: {
    intent: "primary",
  },
})

export const tdStyle = cva("", {
  variants: {
    intent: {
      primary: "p-2",
      secondary: "p-2",
      danger: "p-2",
      success: "p-2",
      warning: "p-2",
      info: "p-2",
      clean: "px-6 py-4 whitespace-nowrap",
    },
  },
  defaultVariants: {
    intent: "primary",
  },
})
