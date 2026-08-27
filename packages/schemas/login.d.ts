import type { z } from "zod"
export declare const loginSchema: z.ZodObject<
  {
    email: z.ZodString
    password: z.ZodString
  },
  "strip",
  z.ZodTypeAny,
  {
    email: string
    password: string
  },
  {
    email: string
    password: string
  }
>
export type LoginProps = z.infer<typeof loginSchema>
