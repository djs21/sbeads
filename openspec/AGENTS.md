# OpenSpec Changes and Specifications

## Purpose

Change management, design proposals, functional specifications, and implementation tracking for sbeads using OpenSpec conventions.

## Ownership

- Active, implemented, and proposed changes under `openspec/changes/`
- Change proposals (`proposal.md`), architectural designs (`design.md`), specifications (`specs/`), and task tracking (`tasks.md`)

## Local Contracts

- Each change lives in its own dedicated directory under `openspec/changes/<change-name>/`
- Change directories include the following standard artifacts:
  - `proposal.md`: Context, motivation, and scope of proposed changes
  - `design.md`: Technical design, architectural trade-offs, and decisions
  - `specs/`: Granular behavioral and capability specifications
  - `tasks.md`: Task checklist with completion tracking
- Task checklists track implementation progress via Markdown checkboxes

## Work Guidance

- Keep task checklists up to date as tasks are worked and completed
- Ensure specifications accurately describe public CLI and internal interface contracts

## Verification

(No automated verification suite for OpenSpec documents)

## Child DOX Index

(No child DOX files)
