# Phase Lifecycle Management

Phases represent high-level milestones from your GSD planning.

## Listing Phases

```bash
curl $MISSION_CONTROL_API_URL/tasks/$TASK_ID/phases
```

Returns an array of phases with their status, progress, and metadata.

## Updating Phase Progress

Report incremental progress (0.0 to 1.0) during execution:

```bash
curl -X POST "$MISSION_CONTROL_API_URL/phases/$PHASE_ID/progress" \
  -H "Content-Type: application/json" \
  -d '{"progress": 0.65, "message": "Planning phase complete, moving to execution..."}'
```

## Completing a Phase

When all stories in a phase are passed:

```bash
curl -X POST "$MISSION_CONTROL_API_URL/phases/$PHASE_ID/complete" \
  -H "Content-Type: application/json" \
  -d '{
    "summary": "Created roadmap with 5 phases and 12 stories",
    "artifacts": {
      "files_created": ["PROJECT.md", "REQUIREMENTS.md", "ROADMAP.md"],
      "files_modified": [],
      "commit_sha": "abc123"
    }
  }'
```

### Artifacts Object

| Field | Type | Description |
|-------|------|-------------|
| `files_created` | string[] | New files produced in this phase |
| `files_modified` | string[] | Existing files changed |
| `commit_sha` | string | Git commit SHA for the phase's work (optional) |

## Failing a Phase

When a phase cannot be completed:

```bash
curl -X POST "$MISSION_CONTROL_API_URL/phases/$PHASE_ID/fail" \
  -H "Content-Type: application/json" \
  -d '{
    "error": "Cannot determine requirements without access to existing API docs",
    "recoverable": true,
    "suggestion": "Need link to API documentation"
  }'
```

| Field | Type | Description |
|-------|------|-------------|
| `error` | string | What went wrong |
| `recoverable` | boolean | Can this be fixed with more info or a different approach? |
| `suggestion` | string | What would unblock this phase |

## Phase Transition Pattern

```
Phase status flow:
  pending → in_progress → completed
                        → failed (recoverable → retry)
                        → failed (unrecoverable → escalate)
```

When a phase completes, trigger the next phase by delegating its stories to the
appropriate agents.