import { cn } from "$/utils/cn"

type LoadingContentProps = {
  lineNumber?: number
}

const LoadingContent = (props: LoadingContentProps) => {
  const { lineNumber = 4 } = props
  const weightSkeleton = [
    "max-w-[30%]",
    "max-w-[90%]",
    "max-w-[60%]",
    "max-w-[80%]",
  ]
  const skeletonLines = Array.from({ length: lineNumber }, (_, index) => {
    const randomChooseIndex = Math.floor(Math.random() * weightSkeleton.length)

    return {
      id: `line-${index + 1}`,
      className: weightSkeleton[index] ?? weightSkeleton[randomChooseIndex],
    }
  })

  return (
    <div className="gap-3 flex flex-col my-4">
      {skeletonLines.map((line) => (
        <div
          key={line.id}
          className={cn(
            "animate-pulse bg-main-100 h-6 rounded",
            line.className,
          )}
        />
      ))}
    </div>
  )
}

export default LoadingContent
