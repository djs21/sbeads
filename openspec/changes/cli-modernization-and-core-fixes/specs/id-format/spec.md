## ADDED Requirements

### Requirement: REQ-IDF-1: 3-character hash ID generation with configurable/default project prefix
The system SHALL generate new issue identifiers using a 3-character lowercase hexadecimal or alphanumeric hash string prefixed with the project or tracker namespace (e.g., `sbeads-xxx` or `issue-xxx`), and SHALL automatically resolve collisions upon generation.

#### Scenario: Generation with default prefix
- **WHEN** an issue is created without custom configuration
- **THEN** the generated identifier matches the regular expression pattern `^(sbeads|issue)-[0-9a-f]{3}$`

#### Scenario: Collision avoidance during ID generation
- **WHEN** the generator generates a 3-character hash that already exists in the target storage
- **THEN** the system generates another distinct 3-character hash without failing the creation request

#### Scenario: Subcommand lookup by short ID
- **WHEN** user invokes commands like `sbeads show <id>`, `sbeads claim <id>`, or `sbeads close <id>` with the 3-character hash ID
- **THEN** the system matches and operates on the corresponding issue record
