# Shipment Tracking Service

## Current stage

- Implemented: domain aggregate, lifecycle validation, event history, rehydration, and unit tests
- Not implemented yet: application layer, repository adapter, protobuf contract, gRPC transport, runnable service binary

## Run and test

```bash
task test
```

Optional local quality commands:

```bash
task fmt
task lint
task check
```

## Architecture direction

The service is being built in clean/hexagonal style:

- Domain owns shipment business rules and invariant protection
- Application will orchestrate use cases
- Adapters will later handle persistence and gRPC

The current domain package depends only on the Go standard library.

## Domain model

### Shipment

`Shipment` is the aggregate root.

It owns:

- shipment identity and business fields
- shipment lifecycle state
- append-only event history
- cross-field invariants such as `driverRevenue <= shipmentAmount`

Validation enforced by the aggregate:

- shipment ID must be non-empty after trimming
- reference number must be non-empty after trimming
- origin and destination must be non-empty after trimming
- origin and destination must differ
- shipment amount and driver revenue must both be explicitly present
- driver revenue cannot exceed shipment amount
- initial status is always `pending`
- every new status event must be a valid transition
- duplicate status transitions are rejected
- out-of-order event timestamps are rejected

Design decisions:

- current status is derived from the last event instead of stored separately
- created/updated timestamps are derived from event history
- event history is the source of truth for shipment state
- `Events()` returns a copy so callers cannot mutate aggregate history from outside

### Status

`Status` is a constrained domain type with a small explicit lifecycle:

- `pending -> picked_up -> in_transit -> delivered`
- `pending -> cancelled`

Assumptions encoded in code and tests:

- `delivered` is terminal
- `cancelled` is terminal
- cancellation is allowed only from `pending`
- statuses are case-sensitive, and non-trimmed string values; invalid or unknown values are rejected

This lifecycle is intentionally small and pragmatic for the task.

### Event

`Event` models a historical shipment status change.

Each event contains:

- status
- sequence number
- occurrence time

Validation enforced for every event:

- sequence must be greater than zero
- status must be one of the allowed domain statuses
- occurrence time must be non-zero
- occurrence time is normalized to UTC

Design decisions:

- event fields are private so invalid events cannot be assembled freely from outside the package
- `RehydrateEvent(...)` is the exported constructor for rebuilding persisted history
- the aggregate uses the same canonical validation path when creating new events internally

### Driver

`Driver` is a validated value object.

Validation:

- driver ID must be non-empty after trimming
- driver name must be non-empty after trimming

Design decisions:

- fields are private and exposed through getters
- constructor validation keeps caller code from bypassing invariants

### Unit

`Unit` is a validated value object representing the transport unit.

Validation:

- unit ID must be non-empty after trimming
- registration number must be non-empty after trimming

Design decisions:

- fields are private and exposed through getters
- constructor validation keeps the value object consistent everywhere it is used

### Money

`Money` stores an amount in minor units as `int64`.

Validation:

- negative amounts are rejected
- aggregate constructors reject `nil` money pointers
- aggregate constructors reject zero-value or uninitialized `Money` values via the internal `valid` flag

Design decisions:

- minor units avoid floating-point precision bugs
- `int64` is enough for the scope of this task and keeps the model simple
- `valid bool` exists to distinguish:
  - an intentionally created zero amount from `NewMoney(0)`
  - an omitted or zero-value `Money{}`
- aggregate params use `*Money` so omission can be detected explicitly at the boundary
- the current model assumes a single currency for now

## Constructors and rehydration

### NewShipment

`NewShipment(NewParams)` is used for new aggregate creation.

It:

- validates shipment metadata
- validates driver and unit value objects
- validates explicit money presence
- creates the initial `pending` event
- normalizes the initial event time to UTC

### Rehydrate

`Rehydrate(RehydrateParams)` exists for rebuilding a shipment from persisted state.

It:

- validates the same shipment metadata as creation
- validates money and supporting value objects again at the aggregate boundary
- validates the full event stream before rebuilding the aggregate

Why this exists:

- repositories need a safe way to rebuild aggregates from storage
- replaying persisted data through command methods would be awkward and error-prone
- rehydration keeps persistence concerns out of the domain object API while still protecting invariants

## Assumptions captured in tests

The current tests intentionally lock down these domain assumptions:

- a newly created shipment always starts with exactly one `pending` event
- valid lifecycle transitions advance status and sequence correctly
- `pending -> cancelled` is allowed
- `picked_up -> cancelled` is not allowed
- `pending -> delivered` is not allowed
- delivered and cancelled shipments reject further transitions
- invalid transitions do not mutate aggregate state
- out-of-order event times are rejected
- rehydration rejects empty history
- rehydration rejects non-pending initial status
- rehydration rejects non-contiguous event sequences
- rehydration rejects non-chronological history
- rehydration rejects invalid money presence
- event timestamps are normalized to UTC
- history returned by `Events()` cannot be mutated back into the aggregate
- the exported API is sufficient to create and advance a shipment from outside the package

## Non-goals for the current slice

These are intentionally deferred to later layers:

- protobuf and gRPC types
- repository ports and adapters
- database schema or persistence details
- service wiring and transport error mapping
- multi-currency support

## Next planned layers

- application service / use cases
- repository port plus in-memory adapter
- `.proto` contract
- gRPC server adapter
- runnable `cmd/shipmentd`
