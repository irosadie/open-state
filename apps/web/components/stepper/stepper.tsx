"use client"

import { cn } from "$/utils/cn"
import { Check } from "lucide-react"

interface Step {
  id: number
  title: string
  description: string
  isCompleted?: boolean
}

interface StepperProps {
  steps: Step[]
  currentStep: number
  onStepClick?: (stepId: number) => void
}

export function Stepper({ steps, currentStep, onStepClick }: StepperProps) {
  return (
    <div className="flex items-center justify-center">
      {steps.map((step, index) => {
        const isCompleted = step.isCompleted ?? currentStep > step.id
        const isCurrent = currentStep === step.id
        const isPending = !isCompleted && !isCurrent

        return (
          <div key={step.id} className="flex items-center">
            {/* Step Circle */}
            <button
              type="button"
              className={cn(
                "flex flex-col items-center border-0 bg-transparent p-0",
                onStepClick &&
                  "cursor-pointer hover:opacity-80 transition-opacity",
              )}
              onClick={() => onStepClick?.(step.id)}
              disabled={!onStepClick}
            >
              <div
                className={cn(
                  "w-10 h-10 rounded-full flex items-center justify-center text-sm font-medium border-2",
                  {
                    "bg-primary-500 border-primary-500 text-white": isCompleted,
                    "bg-white border-gray-300 text-gray-400": isPending,
                  },
                )}
              >
                {isCompleted ? <Check size={16} /> : step.id}
              </div>

              {/* Step Title */}
              <div className="mt-2 text-center max-w-24">
                <p
                  className={cn("text-xs font-medium", {
                    "text-primary-600": isCompleted || isCurrent,
                    "text-gray-400": isPending,
                  })}
                >
                  {step.title}
                </p>
                <p
                  className={cn("text-xs mt-1", {
                    "text-gray-600": isCompleted || isCurrent,
                    "text-gray-400": isPending,
                  })}
                >
                  {step.description}
                </p>
              </div>
            </button>

            {/* Connector Line */}
            {index < steps.length - 1 && (
              <div
                className={cn("h-0.5 w-16 mx-4", {
                  "bg-primary-500": isCompleted,
                  "bg-gray-300": !isCompleted,
                })}
              />
            )}
          </div>
        )
      })}
    </div>
  )
}

interface StepperNavigationProps {
  currentStep: number
  totalSteps: number
  onNext: () => void
  onPrevious: () => void
  onSubmit: () => void
  isLoading: boolean
  canProceed: boolean
  submitLabel?: string
}

export function StepperNavigation({
  currentStep,
  totalSteps,
  onNext,
  onPrevious,
  onSubmit,
  isLoading,
  canProceed,
  submitLabel = "Submit",
}: StepperNavigationProps) {
  const isFirstStep = currentStep === 1
  const isLastStep = currentStep === totalSteps

  return (
    <div className="flex items-center justify-between pt-6 border-t border-gray-100">
      <button
        type="button"
        onClick={onPrevious}
        disabled={isFirstStep}
        className={cn("px-4 py-2 text-sm font-medium rounded-md", {
          "bg-gray-100 text-gray-400 cursor-not-allowed": isFirstStep,
          "bg-white text-gray-700 border border-gray-300 hover:bg-gray-50":
            !isFirstStep,
        })}
      >
        Previous
      </button>

      <div className="text-sm text-gray-500">
        Step {currentStep} of {totalSteps}
      </div>

      {isLastStep ? (
        <button
          type="button"
          onClick={onSubmit}
          disabled={!canProceed || isLoading}
          className={cn("px-4 py-2 text-sm font-medium rounded-md", {
            "bg-primary-500 text-white hover:bg-primary-600":
              canProceed && !isLoading,
            "bg-gray-300 text-gray-400 cursor-not-allowed":
              !canProceed || isLoading,
          })}
        >
          {isLoading ? "Submitting..." : submitLabel}
        </button>
      ) : (
        <button
          type="button"
          onClick={onNext}
          disabled={!canProceed}
          className={cn("px-4 py-2 text-sm font-medium rounded-md", {
            "bg-primary-500 text-white hover:bg-primary-600": canProceed,
            "bg-gray-300 text-gray-400 cursor-not-allowed": !canProceed,
          })}
        >
          Next
        </button>
      )}
    </div>
  )
}
