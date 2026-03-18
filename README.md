# Shipment Tracking gRPC Microservice

## Overview

This repository contains a shipment tracking gRPC microservice built for the take-home task.

The service supports four operations:

- create a shipment
- retrieve shipment details
- add a shipment status event
- retrieve shipment status history

The implementation follows a clean / hexagonal style:

- `internal/domain/shipment` owns business rules and invariant protection
- `internal/application` orchestrates use cases through ports
- `internal/adapters/...` provides in-memory persistence, clock/ID adapters, and gRPC transport
- `cmd/shipmentd` wires the service into a runnable gRPC server

## Run

Generate protobuf stubs:

```bash
task proto
```

Start the service:

```bash
task run
```

The gRPC server listens on `:8080` by default.

Use a different address if needed:

```bash
SHIPMENTD_ADDR=:9090 task run
```

## Test

Run the test suite:

```bash
task test
```

Additional local verification commands:

```bash
go vet ./...
task lint
task check
```

## Architecture

### Domain

`Shipment` is the aggregate root.

The domain layer is isolated from gRPC, protobuf, and persistence concerns. It models:

- shipment identity and business fields
- shipment lifecycle state
- append-only status event history
- validated value objects for driver, unit, and money

Core invariant protection lives here:

- initial status is always `pending`
- current status is always derived from the latest valid event
- invalid lifecycle transitions are rejected
- duplicate status transitions are rejected
- out-of-order events are rejected
- `driverRevenue` cannot exceed `shipmentAmount`

### Application

The application layer exposes transport-agnostic commands, queries, and DTOs for:

- `CreateShipment`
- `GetShipment`
- `AddStatusEvent`
- `GetShipmentHistory`

It is responsible for:

- mapping input into domain value objects
- using the repository, clock, and ID generator ports
- keeping request-context handling outside the domain

### Adapters

Implemented adapters:

- in-memory shipment repository
- UTC system clock
- shipment ID generator
- gRPC server adapter

The gRPC adapter only:

- maps protobuf requests into application commands/queries
- calls the application service
- maps DTOs and errors back into protobuf / gRPC responses

## Protocol Buffers

The contract is defined in:

- `api/proto/shipment/v1/shipment.proto`

RPCs:

- `CreateShipment`
- `GetShipment`
- `AddShipmentStatusEvent`
- `GetShipmentHistory`

Contract choices:

- read/update RPCs use `reference_number` as the business lookup key
- shipment status is a protobuf enum
- timestamps use `google.protobuf.Timestamp`
- money uses integer minor units
- create-request money fields are `optional int64` so omission is distinguishable from an intentional zero

Generated files:

- `api/proto/shipment/v1/shipment.pb.go`
- `api/proto/shipment/v1/shipment_grpc.pb.go`

`task proto` installs local codegen tools into `./bin` and keeps Buf cache data inside the repo-local `.cache` directory.

## Design Decisions

### Business identifier

The service uses two identifiers:

- `reference number` is the external, business-facing lookup key
- `shipment ID` is an internally generated technical identifier

Read and update operations are reference-number-centric because that best matches the task language and expected client behavior.

### Event history as source of truth

Shipment state is derived from event history instead of duplicated mutable fields.

That keeps:

- current status
- created timestamp
- updated timestamp

consistent with the recorded status timeline.

### Money model

`Money` stores minor units as `int64`.

This avoids floating-point precision issues and keeps arithmetic safe for the scope of the task.

The value object also carries an internal `valid` flag so the aggregate can distinguish:

- an intentionally created zero amount via `NewMoney(0)`
- an omitted / zero-value `Money{}`

To preserve that invariant across boundaries:

- domain constructors accept `*Money`
- application create commands accept `*int64`
- protobuf create requests use `optional int64`

### Rehydration

The domain exposes `Rehydrate(...)` for repository use.

This keeps persistence concerns outside the aggregate while still validating loaded state:

- non-empty history
- valid initial event
- contiguous sequence numbers
- chronological ordering
- valid lifecycle transitions

### In-memory repository

The memory repository stores detached aggregate copies and uses copy-on-write update semantics.

That means:

- callers cannot mutate stored state through returned pointers
- failed updates do not partially persist aggregate changes

### gRPC error mapping

Errors are mapped consistently as follows:

- invalid input / domain validation failures -> `InvalidArgument`
- shipment not found -> `NotFound`
- duplicate reference number -> `AlreadyExists`
- canceled context -> `Canceled`
- deadline exceeded -> `DeadlineExceeded`
- unexpected failures -> `Internal`

## Assumptions

- Shipment lifecycle is:
  - `pending -> picked_up -> in_transit -> delivered`
  - `pending -> cancelled`
- `delivered` is terminal.
- `cancelled` is terminal.
- Cancellation is allowed only from `pending`.
- Status timestamps are normalized to UTC.
- Origin and destination are treated as literal user-provided labels after trimming.
- The current implementation assumes a single currency.
- The service uses an in-memory repository for the take-home scope, so data is lost on restart.
- Status event timestamps are assigned by the server clock, not supplied by the client over gRPC.

## Test Coverage

The test suite currently covers:

- shipment creation
- valid status transitions
- invalid status transitions
- terminal-state behavior
- event chronology
- rehydration validation
- money presence validation
- application boundary behavior
- in-memory repository copy-on-write safety
- gRPC happy-path request handling
- gRPC error-to-status-code mapping

## Repository Layout

```text
cmd/shipmentd
api/proto/shipment/v1
internal/domain/shipment
internal/application
internal/ports
internal/adapters/clock
internal/adapters/id
internal/adapters/grpc
internal/adapters/repository/memory
```

## Optional Improvements Not Implemented

- persistent storage
- Docker / docker-compose
- structured logging
- configuration package
- interceptors / auth / observability
- integration tests against an external client tool
