import { z } from "zod"

const registerBaseSchema = z.object({
  name: z
    .string()
    .min(1, "Name is required")
    .max(120, "Name must be 120 characters or less"),
  email: z.string().min(1, "Email is required").email("Invalid email format"),
  password: z
    .string()
    .min(8, "Password must be at least 8 characters")
    .max(64, "Password must be 64 characters or less"),
  confirmPassword: z.string().min(1, "Password confirmation is required"),
})

export const registerSchema = registerBaseSchema.refine(
  (data) => data.password === data.confirmPassword,
  {
    path: ["confirmPassword"],
    message: "Password confirmation does not match",
  },
)

export const registerPayloadSchema = registerBaseSchema.omit({
  confirmPassword: true,
})

export type RegisterProps = z.infer<typeof registerSchema>
export type RegisterPayloadProps = z.infer<typeof registerPayloadSchema>
