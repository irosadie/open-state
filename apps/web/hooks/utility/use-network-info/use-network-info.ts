import { useEffect, useState } from "react"

export type NetworkInfoResponse = {
  isOnline: boolean
}

const useNetworkInfo = (): NetworkInfoResponse => {
  const [isOnline, setIsOnline] = useState(true)

  useEffect(() => {
    const updateNetworkStatus = () => {
      setIsOnline(window.navigator.onLine)
    }

    updateNetworkStatus()
    window.addEventListener("online", updateNetworkStatus)
    window.addEventListener("offline", updateNetworkStatus)

    return () => {
      window.removeEventListener("online", updateNetworkStatus)
      window.removeEventListener("offline", updateNetworkStatus)
    }
  }, [])

  return { isOnline }
}

export default useNetworkInfo
