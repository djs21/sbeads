## ADDED Requirements

### Requirement: REQ-STATUS-1: Status Badge Evaluation
The status engine SHALL evaluate the computed state of an issue following precedence:
1. If `is_closed == 1`, badge SHALL be `[X]` (Closed).
2. Else if `is_claimed == 1`, badge SHALL be `[C]` (Claimed).
3. Else if `is_blocked == 1` OR there exists an unclosed parent dependency, badge SHALL be `[B]` (Blocked).
4. Otherwise, badge SHALL be `[O]` (Open / Ready).

#### Scenario: Evaluating open issue without blockers
- **GIVEN** an issue with is_closed=0, is_claimed=0, is_blocked=0 and no unclosed parent dependencies
- **WHEN** status badge is computed
- **THEN** status is StatusOpen and badge is `[O]`

#### Scenario: Evaluating blocked issue with open parent
- **GIVEN** an issue with is_closed=0, is_claimed=0 and an unclosed parent dependency
- **WHEN** status badge is computed
- **THEN** status is StatusBlocked and badge is `[B]`

### Requirement: REQ-STATUS-2: Epic Badge Marker
When an issue has `issue_type == "epic"`, the formatter SHALL include a prominent `[EPIC]` badge immediately following the status badge.

#### Scenario: Formatter renders epic badge
- **GIVEN** an issue with issue_type="epic"
- **WHEN** one-liner string is formatted
- **THEN** output includes `[EPIC]` badge

### Requirement: REQ-STATUS-3: Priority & Type Indicator
Every formatted issue line SHALL display priority rating and issue type in the format `(P<priority>/<type>)`.

#### Scenario: Formatter renders priority and type
- **GIVEN** an issue with priority=1 and issue_type="feature"
- **WHEN** one-liner string is formatted
- **THEN** output includes `(P1/feature)`

### Requirement: REQ-STATUS-4: Claimed Agent Display
When an issue is in `[C]` (Claimed) state and `agent` is set, the formatted line SHALL append `[@<agent>]` at the end.

#### Scenario: Formatter renders claimed agent
- **GIVEN** an issue with is_claimed=1 and agent="worker-1"
- **WHEN** one-liner string is formatted
- **THEN** output ends with `[@worker-1]`
