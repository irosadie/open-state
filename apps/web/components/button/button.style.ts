import { cn } from "$/utils/cn"
import { cva } from "class-variance-authority"

const rawButton = cva(
  "button flex gap-1.5 justify-center items-center flex-nowrap h-fit hover:opacity-80 hover:cursor-pointer disabled:cursor-not-allowed disabled:hover:opacity-100 transition-colors",
  {
    variants: {
      intent: {
        primary: "bg-primary-500",
        secondary: "bg-secondary-500",
        danger: "bg-danger-500",
        success: "bg-success-500",
        warning: "bg-warning-500",
        info: "bg-info-500",
        clean: "ring-gray-300",
      },
      size: {
        tiny: "text-xs py-1 px-2 font-normal",
        small: "text-sm py-1.5 px-3 font-normal",
        medium: "text-sm py-2.5 px-4 font-normal",
        large: "text-lg py-3 px-6 font-semibold",
        giant: "text-xl py-4 px-8 font-bold",
      },
      disabled: {
        false: undefined,
        true: "opacity-50 hover:opacity-50 cursor-not-allowed",
      },
      uppercase: {
        false: undefined,
        true: "uppercase",
      },
      rounded: {
        none: undefined,
        tiny: "rounded",
        small: "rounded-sm",
        medium: "rounded-md",
        large: "rounded-lg",
        full: "rounded-full",
      },
      bordered: {
        false: undefined,
        true: "ring-[1px] ring-inset",
      },
      textOnly: {
        false: undefined,
        true: "!bg-transparent ring-[0px]",
      },
      fullWidth: {
        false: "w-fit",
        true: "w-full",
      },
    },
    compoundVariants: [
      {
        bordered: true,
        intent: "primary",
        class: "text-primary-500 ring-primary-500 !bg-transparent",
      },
      {
        bordered: true,
        intent: "secondary",
        class: "text-secondary-500 ring-secondary-500 !bg-transparent",
      },
      {
        bordered: true,
        intent: "success",
        class: "text-success-500 ring-success-500 !bg-transparent",
      },
      {
        bordered: true,
        intent: "danger",
        class: "text-danger-500 ring-danger-500 !bg-transparent",
      },
      {
        bordered: true,
        intent: "warning",
        class: "text-warning-500 ring-warning-500 !bg-transparent",
      },
      {
        bordered: true,
        intent: "info",
        class: "text-info-500 ring-info-500 !bg-transparent",
      },
      {
        textOnly: true,
        intent: "primary",
        class: "text-primary-500",
      },
      {
        textOnly: true,
        intent: "secondary",
        class: "text-secondary-500",
      },
      {
        textOnly: true,
        intent: "success",
        class: "text-success-500",
      },
      {
        textOnly: true,
        intent: "danger",
        class: "text-danger-500",
      },
      {
        textOnly: true,
        intent: "warning",
        class: "text-warning-500",
      },
      {
        textOnly: true,
        intent: "info",
        class: "text-info-500",
      },
      {
        textOnly: true,
        intent: "clean",
        class: "text-gray-700",
      },
      {
        textOnly: false,
        intent: "primary",
        bordered: false,
        class: "text-white",
      },
      {
        textOnly: false,
        intent: "secondary",
        bordered: false,
        class: "text-white",
      },
      {
        textOnly: false,
        intent: "success",
        bordered: false,
        class: "text-white",
      },
      {
        textOnly: false,
        intent: "danger",
        bordered: false,
        class: "text-white",
      },
      {
        textOnly: false,
        intent: "warning",
        bordered: false,
        class: "text-white",
      },
      {
        textOnly: false,
        intent: "info",
        bordered: false,
        class: "text-white",
      },
      {
        disabled: false,
        class: "transition transform active:scale-95",
      },
    ],
    defaultVariants: {
      disabled: false,
      intent: "primary",
      size: "medium",
      bordered: false,
      textOnly: false,
      rounded: "tiny",
    },
  },
)

export const button = (...args: Parameters<typeof rawButton>) =>
  cn(rawButton(...args))
