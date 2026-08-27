"use client"

import { cn } from "$/utils/cn"
import * as TabsPrimitive from "@radix-ui/react-tabs"
import type { VariantProps } from "class-variance-authority"
import * as React from "react"
import { tabsContent, tabsList, tabsTrigger } from "./tabs.style"

const Tabs = TabsPrimitive.Root

const TabsList = React.forwardRef<
  React.ElementRef<typeof TabsPrimitive.List>,
  React.ComponentPropsWithoutRef<typeof TabsPrimitive.List> &
    VariantProps<typeof tabsList>
>(({ className, intent, ...props }, ref) => (
  <TabsPrimitive.List
    ref={ref}
    className={cn(tabsList({ intent }), className)}
    {...props}
  />
))

TabsList.displayName = TabsPrimitive.List.displayName

const TabsTrigger = React.forwardRef<
  React.ElementRef<typeof TabsPrimitive.Trigger>,
  React.ComponentPropsWithoutRef<typeof TabsPrimitive.Trigger> &
    VariantProps<typeof tabsTrigger>
>(({ className, intent, size, ...props }, ref) => (
  <TabsPrimitive.Trigger
    ref={ref}
    className={cn(tabsTrigger({ intent, size }), className)}
    {...props}
  />
))

TabsTrigger.displayName = TabsPrimitive.Trigger.displayName

const TabsContent = React.forwardRef<
  React.ElementRef<typeof TabsPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof TabsPrimitive.Content> &
    VariantProps<typeof tabsContent>
>(({ className, intent, ...props }, ref) => (
  <TabsPrimitive.Content
    ref={ref}
    className={cn(tabsContent({ intent }), className)}
    {...props}
  />
))

TabsContent.displayName = TabsPrimitive.Content.displayName

export { Tabs, TabsList, TabsTrigger, TabsContent }
