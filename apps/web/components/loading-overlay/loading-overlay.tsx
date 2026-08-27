import Image from "next/image"

const LoadingOverlay = () => {
  return (
    <div className="absolute top-0 h-screen w-screen flex items-center justify-center bg-transparent">
      <div className="absolute top-0 left-0 h-full w-full bg-black opacity-10" />
      <div className="relative z-10">
        <Image width={500} height={200} alt="logo" src="/imgs/logo.png" />
      </div>
    </div>
  )
}

export default LoadingOverlay
