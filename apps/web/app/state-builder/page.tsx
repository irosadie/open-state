import { StateBuilder } from "$/components/state-builder"

export const metadata = {
  title: "State Builder",
  description: "Visual workflow & conversation state builder",
}

export default function StateBuilderPage() {
  return (
    <main className="h-screen w-screen overflow-hidden">
      <StateBuilder />
    </main>
  )
}
