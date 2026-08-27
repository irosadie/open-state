import { cn } from "$/utils/cn"

type DisplayWithSkeletonProps = {
  text: string
  isLoading: boolean
  skeletonClass?: string
}

const htmlToText = (value: string) => {
  if (typeof window !== "undefined" && "DOMParser" in window) {
    const document = new window.DOMParser().parseFromString(value, "text/html")

    return document.body.textContent ?? ""
  }

  return value.replace(/<[^>]*>/g, "")
}

const DisplayWithSkeleton = (props: DisplayWithSkeletonProps) => {
  const { text, isLoading, skeletonClass = "w-[90%]" } = props

  if (isLoading) {
    return (
      <div
        className={cn("h-6 bg-main-100 rounded animate-pulse", skeletonClass)}
      />
    )
  }
  return <span>{htmlToText(text)}</span>
}

export default DisplayWithSkeleton
