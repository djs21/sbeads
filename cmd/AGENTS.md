# Command Line Applications

## Purpose

Root container for command-line binary entry points of the sbeads project.

## Ownership

- Command-line application directories under `cmd/`
- Top-level executable structure and build targets

## Local Contracts

- Each subdirectory in `cmd/` represents an independent executable program
- CLI tools use Go standard packaging under `package main`
- Entry points coordinate with internal domains and backend storage through structured application contexts

## Work Guidance

- Avoid embedding business logic directly in top-level command dispatchers
- Direct command-line parsing concerns to specific binary subpackages

## Verification

- Run test suite: `go test ./cmd/...`

## Child DOX Index

- [bd/AGENTS.md](bd/AGENTS.md): CLI implementation and subcommands for the `sbeads` / `bd` tool
