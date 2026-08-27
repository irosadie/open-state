import { describe, expect, it } from "vitest"
import {
  type RegisterPayloadProps,
  type RegisterProps,
  registerPayloadSchema,
  registerSchema,
} from "./register"

describe("Register Schema Validation", () => {
  describe("Valid Register Data", () => {
    it("should validate complete register payload", () => {
      const validData: RegisterProps = {
        name: "Baim Wong",
        email: "baim@example.com",
        password: "Password123",
        confirmPassword: "Password123",
      }

      const result = registerSchema.safeParse(validData)

      expect(result.success).toBe(true)
    })

    it("should validate payload schema without confirmPassword", () => {
      const validPayload: RegisterPayloadProps = {
        name: "Baim Wong",
        email: "baim@example.com",
        password: "Password123",
      }

      const result = registerPayloadSchema.safeParse(validPayload)

      expect(result.success).toBe(true)
    })
  })

  describe("Invalid Register Data", () => {
    it("should reject mismatched password confirmation", () => {
      const invalidData: RegisterProps = {
        name: "Baim Wong",
        email: "baim@example.com",
        password: "Password123",
        confirmPassword: "Password321",
      }

      const result = registerSchema.safeParse(invalidData)

      expect(result.success).toBe(false)

      if (result.success) {
        throw new Error("Expected register schema to reject invalid data")
      }

      expect(result.error.issues[0]?.path).toEqual(["confirmPassword"])
    })

    it("should reject short password", () => {
      const invalidData: RegisterProps = {
        name: "Baim Wong",
        email: "baim@example.com",
        password: "12345",
        confirmPassword: "12345",
      }

      const result = registerSchema.safeParse(invalidData)

      expect(result.success).toBe(false)
    })
  })
})
