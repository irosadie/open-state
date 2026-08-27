import { LoaderCircleIcon } from "lucide-react"

const LoadingSpinner = () => {
  return (
    <div className="flex flex-1 h-full w-full items-center justify-center">
      <LoaderCircleIcon
        className="animate-spin mx-auto -mt-40 text-muted-foreground"
        size={48}
      />
    </div>
  )
}

export default LoadingSpinner
