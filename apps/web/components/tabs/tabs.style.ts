import { cva } from "class-variance-authority"

export const tabsList = cva(
  "inline-flex w-full max-w-fit items-center rounded-xl border p-1",
  {
    variants: {
      intent: {
        primary: "border-primary-200 bg-primary-50/40",
        secondary: "border-secondary-200 bg-secondary-50/40",
        danger: "border-danger-200 bg-danger-50/40",
        success: "border-success-200 bg-success-50/40",
        warning: "border-warning-200 bg-warning-50/40",
        info: "border-info-200 bg-info-50/40",
        clean: "border-gray-200 bg-white",
      },
    },
    defaultVariants: {
      intent: "clean",
    },
  },
)

export const tabsTrigger = cva(
  [
    "inline-flex items-center justify-center gap-2 rounded-lg font-semibold transition-colors",
    "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-500 focus-visible:ring-offset-1",
    "disabled:pointer-events-none disabled:opacity-50",
  ],
  {
    variants: {
      intent: {
        primary:
          "text-primary-500 hover:text-primary-600 data-[state=active]:bg-primary-100 data-[state=active]:text-primary-700",
        secondary:
          "text-secondary-500 hover:text-secondary-600 data-[state=active]:bg-secondary-100 data-[state=active]:text-secondary-700",
        danger:
          "text-danger-500 hover:text-danger-600 data-[state=active]:bg-danger-100 data-[state=active]:text-danger-700",
        success:
          "text-success-500 hover:text-success-600 data-[state=active]:bg-success-100 data-[state=active]:text-success-700",
        warning:
          "text-warning-600 hover:text-warning-700 data-[state=active]:bg-warning-100 data-[state=active]:text-warning-800",
        info: "text-info-500 hover:text-info-600 data-[state=active]:bg-info-100 data-[state=active]:text-info-700",
        clean:
          "text-gray-500 hover:text-gray-800 data-[state=active]:bg-primary-50 data-[state=active]:text-primary-700",
      },
      size: {
        small: "px-3 py-1.5 text-xs",
        medium: "px-4 py-2 text-sm",
        large: "px-5 py-2.5 text-base",
      },
    },
    defaultVariants: {
      intent: "clean",
      size: "medium",
    },
  },
)

export const tabsContent = cva(
  "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-500 focus-visible:ring-offset-1",
  {
    variants: {
      intent: {
        primary: undefined,
        secondary: undefined,
        danger: undefined,
        success: undefined,
        warning: undefined,
        info: undefined,
        clean: undefined,
      },
    },
    defaultVariants: {
      intent: "clean",
    },
  },
)
