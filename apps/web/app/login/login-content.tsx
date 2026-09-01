"use client"

import { Button } from "$/components/button"
import { Input } from "$/components/input"
import { PanelCard } from "$/components/panel-card"
import { authConfig } from "$/configs/auth"
import { useAuthorization } from "$/providers/authorization-provider"
import { resolveAuthorizedPath } from "$/utils/rbac"
import { type LoginProps, loginSchema } from "@openstate/schemas"
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
  if (value?.startsWith("/") && !value.startsWith("//")) {
    return value
  }

  return authConfig.defaultRedirectPath
}

type LoginErrors = Partial<Record<keyof LoginProps, string>>

export default function LoginContent() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { status: sessionStatus } = useSession()
  const authorization = useAuthorization()
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

  const authorizedPath = useMemo(
    () => resolveAuthorizedPath(callbackUrl, authorization.permissions),
    [authorization.permissions, callbackUrl],
  )

  useEffect(() => {
    if (
      sessionStatus === "authenticated" &&
      authorization.status === "ready" &&
      authorizedPath
    ) {
      router.replace(authorizedPath)
    }
  }, [authorizedPath, authorization.status, router, sessionStatus])

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
      })()
    })
  }

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-md items-center px-6 py-12">
      <PanelCard
        className="w-full rounded-3xl"
        title="Sign In"
        description="OpenState credentials flow via NextAuth and proxy BFF"
      >
        <form className="space-y-4" onSubmit={handleSubmit}>
          <Input
            autoComplete="email"
            data-testid="login-email"
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
            data-testid="login-password"
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
            <p className="text-sm text-danger-500" data-testid="login-error">
              {formError}
            </p>
          ) : null}

          {sessionStatus === "authenticated" &&
          authorization.status === "ready" &&
          !authorizedPath ? (
            <p className="text-sm text-danger-500">
              Your account has no accessible application area.
            </p>
          ) : null}

          <Button
            className="w-full justify-center"
            data-testid="login-submit"
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
