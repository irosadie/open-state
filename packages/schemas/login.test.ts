import { describe, expect, it } from "vitest"
import { type LoginProps, loginSchema } from "./login"

describe("Login Schema Validation", () => {
  describe("Valid Login Data", () => {
    it("should validate correct email and password", () => {
      const validData: LoginProps = {
        email: "admin@kirimmobil.id",
        password: "Admin123!",
      }

      const result = loginSchema.safeParse(validData)

      expect(result.success).toBe(true)
    })

    it("should validate email with different valid formats", () => {
      const validEmails = [
        "user@example.com",
        "user.name@example.com",
        "user+tag@example.com",
        "user@subdomain.example.com",
      ]

      for (const email of validEmails) {
        const validData: LoginProps = {
          email,
          password: "password123",
        }
        const result = loginSchema.safeParse(validData)

        expect(result.success).toBe(true)
      }
    })
  })

  describe("Invalid Email Validation", () => {
    it("should reject empty email", () => {
      const invalidData: LoginProps = {
        email: "",
        password: "password123",
      }

      const result = loginSchema.safeParse(invalidData)

      expect(result.success).toBe(false)
    })

    it("should reject invalid email format", () => {
      const invalidEmails = ["notanemail", "missing@", "@nodomain"]

      for (const email of invalidEmails) {
        const invalidData: LoginProps = {
          email,
          password: "password123",
        }
        const result = loginSchema.safeParse(invalidData)

        expect(result.success).toBe(false)
      }
    })
  })

  describe("Invalid Password Validation", () => {
    it("should reject empty password", () => {
      const invalidData: LoginProps = {
        email: "test@example.com",
        password: "",
      }

      const result = loginSchema.safeParse(invalidData)

      expect(result.success).toBe(false)
    })
  })

  describe("Login Request Schema - API Contract Compliance", () => {
    it("should match OpenAPI LoginRequest schema", () => {
      const apiRequest: LoginProps = {
        email: "admin@kirimmobil.id",
        password: "Admin123!",
      }

      const result = loginSchema.safeParse(apiRequest)

      expect(result.success).toBe(true)

      const parsed = loginSchema.parse(apiRequest)

      expect(typeof parsed.email).toBe("string")
      expect(typeof parsed.password).toBe("string")
    })

    it("should handle maximum length inputs", () => {
      const longEmail = `${"a".repeat(100)}@example.com`
      const longPassword = "a".repeat(1000)

      const data: LoginProps = {
        email: longEmail,
        password: longPassword,
      }

      const result = loginSchema.safeParse(data)

      expect(result.success).toBe(true)
    })
  })
})
