## Purpose

Provides background job processing in Go using asynq (Redis-backed), replacing the BullMQ TypeScript worker.

## ADDED Requirements

### Requirement: Worker process connects to Redis
The system SHALL connect to Redis using the `REDIS_URL` environment variable to consume job queues via asynq.

#### Scenario: Successful Redis connection
- **WHEN** `REDIS_URL` is set to a valid Redis connection string
- **THEN** the worker SHALL connect and start polling queues

#### Scenario: Missing REDIS_URL fails fast
- **WHEN** `REDIS_URL` is not set
- **THEN** the worker process SHALL exit at startup with a clear error message

### Requirement: Job handlers registered on startup
The system SHALL register all job type handlers with the asynq server before starting to process jobs.

#### Scenario: Known job type processed
- **WHEN** a job with a registered type is enqueued
- **THEN** the worker SHALL dequeue and execute the corresponding handler

#### Scenario: Unknown job type handled gracefully
- **WHEN** a job with an unregistered type is dequeued
- **THEN** the worker SHALL log the error and not crash the process

### Requirement: Failed jobs are retried
The system SHALL retry failed jobs according to asynq's default retry policy (up to 25 retries with exponential backoff).

#### Scenario: Job retry on handler error
- **WHEN** a job handler returns an error
- **THEN** asynq SHALL re-enqueue the job for retry up to the configured maximum
