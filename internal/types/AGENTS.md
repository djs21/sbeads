# Core Types

## Purpose

Data models, constants, and ID generation utilities for sbeads.

## Ownership

- Data structures: `Issue` and `Dependency`
- Cryptographic Base36 identifier generation (`GenerateID`, `GenerateIDWithPrefix`, `GenerateShortHash`)
- Shared character sets and formatting constants

## Local Contracts

- Identifiers follow the format `<prefix>-<hash3>`, where `<prefix>` defaults to `sbeads` and `<hash3>` is a 3-character random string from Base36 charset `0123456789abcdefghijklmnopqrstuvwxyz`
- `Issue` and `Dependency` structs remain plain Go structures with standard types and `time.Time`
- Status fields use integer flags (`IsBlocked`, `IsClosed`, `IsClaimed`)

## Work Guidance

- Keep types minimal and free from external dependencies outside standard library
- Maintain compatibility with database columns in `backend`

## Verification

- Run test suite: `go test ./internal/types/...`

## Child DOX Index

(No child DOX files)
