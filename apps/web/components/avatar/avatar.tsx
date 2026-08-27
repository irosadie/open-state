"use client"

import { cn } from "$/utils/cn"
import * as AvatarPrimitive from "@radix-ui/react-avatar"
import React from "react"

const AvatarComponent = React.forwardRef<
  React.ElementRef<typeof AvatarPrimitive.Root>,
  React.ComponentPropsWithoutRef<typeof AvatarPrimitive.Root>
>(({ className, ...props }, ref) => (
  <AvatarPrimitive.Root
    ref={ref}
    className={cn(
      "relative flex h-10 w-10 shrink-0 overflow-hidden rounded-full",
      className,
    )}
    {...props}
  />
))

AvatarComponent.displayName = AvatarPrimitive.Root.displayName

const AvatarImage = React.forwardRef<
  React.ElementRef<typeof AvatarPrimitive.Image>,
  React.ComponentPropsWithoutRef<typeof AvatarPrimitive.Image>
>(({ className, ...props }, ref) => (
  <AvatarPrimitive.Image
    ref={ref}
    className={cn("aspect-square h-full w-full", className)}
    {...props}
  />
))

AvatarImage.displayName = AvatarPrimitive.Image.displayName

const AvatarFallback = React.forwardRef<
  React.ElementRef<typeof AvatarPrimitive.Fallback>,
  React.ComponentPropsWithoutRef<typeof AvatarPrimitive.Fallback>
>(({ className, ...props }, ref) => (
  <AvatarPrimitive.Fallback
    ref={ref}
    className={cn(
      "flex h-full w-full items-center justify-center rounded-full bg-muted",
      className,
    )}
    {...props}
  />
))

AvatarFallback.displayName = AvatarPrimitive.Fallback.displayName

export { Avatar, AvatarImage, AvatarFallback }

type AvatarProps = {
  src?: string
  name?: string
  className?: string
}

const Avatar = (props: AvatarProps) => {
  const { src = undefined, name = "User" } = props

  const generateAlias = (name: string): string => {
    const words = name.split(" ").filter(Boolean)

    if (words.length === 0) {
      return "US"
    }

    if (words.length === 1) {
      const [firstWord] = words
      return (firstWord ?? "US").slice(0, 2).toUpperCase()
    }

    if (words.length === 2) {
      const [firstWord, secondWord] = words
      return `${firstWord?.[0] ?? ""}${secondWord?.[0] ?? ""}`.toUpperCase()
    }

    return words
      .slice(0, 3)
      .map((word) => word[0] ?? "")
      .join("")
      .toUpperCase()
  }

  return (
    <AvatarComponent>
      <AvatarImage src={src} />
      <AvatarFallback>{generateAlias(name)}</AvatarFallback>
    </AvatarComponent>
  )
}

export default Avatar
