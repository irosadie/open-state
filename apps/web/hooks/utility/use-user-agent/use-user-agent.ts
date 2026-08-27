import { useEffect, useState } from "react"

const useUserAgent = () => {
  const [userAgent, setUserAgent] = useState("")

  useEffect(() => {
    setUserAgent(navigator.userAgent)
  }, [])

  return { userAgent }
}

export default useUserAgent
