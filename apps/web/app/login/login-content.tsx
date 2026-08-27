"use client"

import { Button } from "$/components/button"
import { Input } from "$/components/input"
import { PanelCard } from "$/components/panel-card"
import { authConfig } from "$/configs/auth"
import { type LoginProps, loginSchema } from "@vibecoding-starter/schemas"
import { signIn, useSession } from "next-auth/react"
import { useRouter, useSearchParams } from "next/navigation"
import {
  type FormEvent,
  useEffect,
  useMemo,
  useState,
  useTransition,
} from "react"

const sanitizeCallbackUrl = (value: string | null) => {
  if (value?.startsWith("/")) {
    return value
  }

  return authConfig.defaultRedirectPath
}

type LoginErrors = Partial<Record<keyof LoginProps, string>>

export default function LoginContent() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { status } = useSession()
  const [isPending, startTransition] = useTransition()
  const [formError, setFormError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<LoginErrors>({})
  const [form, setForm] = useState<LoginProps>({
    email: "",
    password: "",
  })

  const callbackUrl = useMemo(
    () => sanitizeCallbackUrl(searchParams.get("callbackUrl")),
    [searchParams],
  )

  useEffect(() => {
    if (status === "authenticated") {
      router.replace(callbackUrl)
    }
  }, [callbackUrl, router, status])

  const handleChange = (field: keyof LoginProps, value: string) => {
    setForm((current) => ({
      ...current,
      [field]: value,
    }))
    setFieldErrors((current) => ({
      ...current,
      [field]: undefined,
    }))
    setFormError("")
  }

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()

    const parsedForm = loginSchema.safeParse(form)

    if (!parsedForm.success) {
      const nextErrors: LoginErrors = {}

      for (const issue of parsedForm.error.issues) {
        const field = issue.path[0]

        if (field === "email" || field === "password") {
          nextErrors[field] = issue.message
        }
      }

      setFieldErrors(nextErrors)
      return
    }

    startTransition(() => {
      void (async () => {
        const result = await signIn("credentials", {
          ...parsedForm.data,
          redirect: false,
          callbackUrl,
        })

        if (result?.error) {
          setFormError(result.error)
          return
        }

        router.replace(result?.url || callbackUrl)
        router.refresh()
      })()
    })
  }

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-md items-center px-6 py-12">
      <PanelCard
        className="w-full rounded-3xl"
        title="Sign In"
        description="Starter credentials flow via NextAuth and proxy BFF"
      >
        <form className="space-y-4" onSubmit={handleSubmit}>
          <Input
            autoComplete="email"
            label="Email"
            name="email"
            type="email"
            value={form.email}
            error={fieldErrors.email}
            onChange={(event) => handleChange("email", event.target.value)}
            placeholder="you@example.com"
            required
          />
          <Input
            autoComplete="current-password"
            label="Password"
            name="password"
            type="password"
            value={form.password}
            error={fieldErrors.password}
            onChange={(event) => handleChange("password", event.target.value)}
            placeholder="Enter your password"
            required
          />

          {formError ? (
            <p className="text-sm text-danger-500">{formError}</p>
          ) : null}

          <Button
            className="w-full justify-center"
            type="submit"
            loading={isPending}
          >
            Sign In
          </Button>
        </form>
      </PanelCard>
    </main>
  )
}
