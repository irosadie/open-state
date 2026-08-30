"use client"

import { Button } from "$/components/button"
import { Input } from "$/components/input"
import { PanelCard } from "$/components/panel-card"
import { authConfig } from "$/configs/auth"
import { useAuthLogin, useAuthRegister } from "$/hooks/transactions/use-auth"
import { useAuthorization } from "$/providers/authorization-provider"
import { resolveAuthorizedPath } from "$/utils/rbac"
import {
  type RegisterProps,
  registerPayloadSchema,
  registerSchema,
} from "@openstate/schemas"
import { useSession } from "next-auth/react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { type FormEvent, useEffect, useMemo, useState } from "react"

type RegisterErrors = Partial<Record<keyof RegisterProps, string>>

const getErrorMessage = (error: unknown) => {
  if (
    typeof error === "object" &&
    error !== null &&
    "message" in error &&
    typeof error.message === "string"
  ) {
    return error.message
  }

  if (error instanceof Error) {
    return error.message
  }

  return "Pendaftaran gagal. Silakan coba lagi."
}

export default function RegisterContent() {
  const router = useRouter()
  const { status: sessionStatus } = useSession()
  const authorization = useAuthorization()
  const registerMutation = useAuthRegister()
  const loginMutation = useAuthLogin()
  const [formError, setFormError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<RegisterErrors>({})
  const [form, setForm] = useState<RegisterProps>({
    name: "",
    email: "",
    password: "",
    confirmPassword: "",
  })

  const isPending = registerMutation.isPending || loginMutation.isPending
  const authorizedPath = useMemo(
    () => resolveAuthorizedPath(null, authorization.permissions),
    [authorization.permissions],
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

  const handleChange = (field: keyof RegisterProps, value: string) => {
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

    const parsedForm = registerSchema.safeParse(form)

    if (!parsedForm.success) {
      const nextErrors: RegisterErrors = {}

      for (const issue of parsedForm.error.issues) {
        const field = issue.path[0]

        if (
          field === "name" ||
          field === "email" ||
          field === "password" ||
          field === "confirmPassword"
        ) {
          nextErrors[field] = issue.message
        }
      }

      setFieldErrors(nextErrors)
      return
    }

    void (async () => {
      try {
        const payload = registerPayloadSchema.parse(parsedForm.data)

        await registerMutation.mutateAsync(payload)

        await loginMutation.mutateAsync({
          email: parsedForm.data.email,
          password: parsedForm.data.password,
          callbackUrl: authConfig.defaultRedirectPath,
        })

        router.refresh()
      } catch (error) {
        setFormError(getErrorMessage(error))
      }
    })()
  }

  return (
    <main className="mx-auto flex min-h-screen w-full max-w-md items-center px-6 py-12">
      <PanelCard
        className="w-full rounded-3xl"
        title="Create Account"
        description="Daftar akun user untuk mulai membeli dan mengikuti ujian"
      >
        <form className="space-y-4" onSubmit={handleSubmit}>
          <Input
            autoComplete="name"
            label="Full Name"
            name="name"
            type="text"
            value={form.name}
            error={fieldErrors.name}
            onChange={(event) => handleChange("name", event.target.value)}
            placeholder="Nama lengkap"
            required
          />
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
            autoComplete="new-password"
            label="Password"
            name="password"
            type="password"
            value={form.password}
            error={fieldErrors.password}
            onChange={(event) => handleChange("password", event.target.value)}
            placeholder="Minimal 8 karakter"
            required
          />
          <Input
            autoComplete="new-password"
            label="Confirm Password"
            name="confirmPassword"
            type="password"
            value={form.confirmPassword}
            error={fieldErrors.confirmPassword}
            onChange={(event) =>
              handleChange("confirmPassword", event.target.value)
            }
            placeholder="Ulangi password"
            required
          />

          {formError ? (
            <p className="text-sm text-danger-500">{formError}</p>
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
            type="submit"
            loading={isPending}
          >
            Create Account
          </Button>

          <p className="text-center text-sm text-slate-600">
            Sudah punya akun?{" "}
            <Link
              href={authConfig.loginPath}
              className="font-semibold text-brand-dark underline-offset-4 hover:underline"
            >
              Masuk di sini
            </Link>
          </p>
        </form>
      </PanelCard>
    </main>
  )
}
