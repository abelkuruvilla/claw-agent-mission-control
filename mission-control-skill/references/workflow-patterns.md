# Common Workflow Patterns

Reference workflows for different project types. Use these as templates when
planning with the GSD protocol.

**Important:** Agent names in these patterns are illustrative. Always discover
available agents via `GET /agents` and match based on their `description` and
`identity_md`. The agent roster changes over time.

All delegation steps below use subtasks with `parent_task_id` — see
`agents-api.md` for the full subtask creation API.

## Software Development Project

```
1. Orchestrator receives task
   → GET /agents to discover available specialists
   → Update task status to "planning"

2. Create subtask: "Research: [topic]" → assign to researcher agent
   - Technology evaluation
   - Competitor analysis
   - Feasibility assessment
   → Wait for subtask to reach "done"
   → Read researcher's progress_txt for findings

3. Orchestrator plans based on research
   - Create GSD artifacts (requirements_md, roadmap_md)
   - Break work into stories

4. Create subtask: "Design: [feature]" → assign to designer agent
   - PRD with user stories and acceptance criteria
   - User flows (happy path + edge cases)
   - Design tokens / UI direction
   → Wait for subtask to reach "done"

5. Create subtask: "Implement: [feature]" → assign to engineer agent
   - Implement per PRD and design specs
   - Write tests alongside code
   - Commit per story
   → Wait for subtask to reach "done"

6. Create subtask: "QA: [feature]" → assign to qa agent
   - Code review
   - Functional testing against acceptance criteria
   - Accessibility and security checks
   → Wait for subtask to reach "done"

7. Orchestrator assembles final delivery
   - Verify all subtasks passed
   - Update STATE.md
   - Mark parent task "done"
```

## Content / Marketing Project

```
1. Orchestrator receives task
   → GET /agents to discover available specialists
   → Update task status to "planning"

2. Create subtask: "Research: [topic]" → assign to researcher agent
   - Topic research
   - SEO keyword analysis
   - Audience insights
   - Content gap analysis
   → Wait for subtask to reach "done"

3. Create subtask: "Write: [content]" → assign to content agent
   - Write content per research brief
   - Apply SEO optimization
   - Include CTAs
   → Wait for subtask to reach "done"

4. Create subtask: "Review: [content]" → assign to qa agent
   - Proofread for grammar and clarity
   - Verify SEO elements
   - Check brand voice consistency
   → Wait for subtask to reach "done"

5. Orchestrator delivers final content
   - Verify all subtasks passed
   - Mark parent task "done"
```

## Design-Only Project

```
1. Orchestrator receives task
   → GET /agents to discover available specialists
   → Update task status to "planning"

2. Create subtask: "Research: [design topic]" → assign to researcher agent
   - Competitor UI/UX analysis
   - Market positioning research
   - User research insights
   → Wait for subtask to reach "done"

3. Create subtask: "Design: [deliverable]" → assign to designer agent
   - Full design system
   - User flows
   - Wireframes / component specs
   - Design tokens
   → Wait for subtask to reach "done"

4. Create subtask: "Review: [design]" → assign to qa agent
   - Design consistency review
   - Accessibility check
   - Responsive behavior verification
   → Wait for subtask to reach "done"

5. Orchestrator delivers design package
   - Verify all subtasks passed
   - Mark parent task "done"
```

## Full Product Launch

```
1. Orchestrator receives task
   → GET /agents to discover available specialists
   → Update task status to "planning"
   → Plan parallel tracks as subtasks:

   Track A (Product):
   Subtask: Research → Subtask: Design → Subtask: Engineer → Subtask: QA

   Track B (Marketing):
   Subtask: Research → Subtask: Content → Subtask: QA

2. Orchestrator monitors both tracks
   - Poll subtask statuses via GET /tasks/$PARENT_ID/subtasks
   - Create next subtask in each track as the previous one completes

3. Orchestrator synchronizes
   - Engineering subtask delivers working product
   - Content subtask delivers launch content
   - QA subtasks verify both tracks
   - Mark parent task "done" when all tracks complete
```

## Hotfix / Bug Fix

```
1. Orchestrator receives task (high priority)
   → GET /agents to discover available specialists
   → Update task status to "executing"

2. Create subtask: "Fix: [bug description]" → assign to engineer agent
   - Reproduce the bug
   - Implement fix
   - Write regression test
   → Wait for subtask to reach "done"

3. Create subtask: "Verify fix: [bug]" → assign to qa agent
   - Verify fix resolves issue
   - Run regression suite
   → Wait for subtask to reach "done"

4. Orchestrator marks parent task "done"
   (skip review for critical fixes if pre-approved)
```

## Delegation Decision Matrix

| Condition | Action |
|-----------|--------|
| Task is well-defined, single domain | Create one subtask assigned to the relevant specialist |
| Task spans multiple domains | Create subtasks per domain, assign to respective agents |
| Task requires research before planning | Create research subtask first, then plan based on results |
| Task is urgent with known solution | Skip research subtask, go straight to engineering subtask |
| Agent is busy (`status: working`) | Assign anyway — the task will be queued and auto-dispatched when the agent is free. Use a higher priority (lower number) for urgent work to jump the queue. Alternatively, assign to another idle agent with similar capabilities |
| Subtask failed | Read error from subtask progress_txt, re-scope and create a new subtask |
| No suitable specialist exists | Execute the work directly as a last resort |
| High-priority task for a busy agent | Assign with a low priority number (e.g., 1). The task will jump ahead of lower-priority items in the agent's queue |
| Multiple tasks for the same agent | Assign all — they will queue in priority order. The agent processes them one at a time |
