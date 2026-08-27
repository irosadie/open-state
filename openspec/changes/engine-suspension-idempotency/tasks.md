## 1. Suspension & resume (engine/suspend-resume)

> Skill: `api-feature`

- [x] 1.1 Add `SUSPENDED` to `WorkflowInstanceStatus`
- [x] 1.2 Implement `SuspendWorkflow(instanceId)` — set status SUSPENDED, preserve
      state/context/history/version (PRD §43)
- [x] 1.3 Implement `ResumeWorkflow(instanceId)` — restore to RUNNING/WAITING from
      saved state
- [x] 1.4 Unit test: suspend preserves context; resume continues from saved state

## 2. Optimistic concurrency (engine/concurrency)

> Skill: `api-feature`

- [x] 2.1 Add `Version` field to `WorkflowInstance`
- [x] 2.2 `ProcessEvent` requires expected version; on mismatch return
      `CONFLICT` (PRD §31)
- [x] 2.3 Update flow increments version atomically with the transition
- [x] 2.4 Extend `InstanceRepository` port with version-aware update
      (`UpdateWithVersion(instance, expectedVersion)`)
- [x] 2.5 Unit test: concurrent update conflict

## 3. Idempotency (engine/idempotency)

> Skill: `api-feature`

- [x] 3.1 Ensure `Event` has `idempotencyKey`
- [x] 3.2 Implement `ProcessEvent` idempotency check: if key already processed,
      return no-op (PRD §30)
- [x] 3.3 Extend `EventRepository` port with `IsProcessed(key)` /
      `MarkProcessed(key)`
- [x] 3.4 Unit test: duplicate event is deduped

## 4. Mid-flow interruption hook

> Skill: `api-feature`

- [x] 4.1 Verify executor supports interruption events (e.g. change-request)
      routing out of running state (PRD §43.1)
- [x] 4.2 Unit test: interruption routes out and can resume

## 5. Quality gate

> Skills: `api-code-review`

- [x] 5.1 `go build ./...`, `go vet ./...`, `go test ./...` pass
- [x] 5.2 `api-code-review` passes
- [x] 5.3 No infra/HTTP/LLM dependency added
