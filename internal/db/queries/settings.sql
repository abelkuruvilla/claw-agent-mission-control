-- name: GetSettings :one
SELECT * FROM settings WHERE id = 'default' LIMIT 1;

-- name: UpdateSettings :one
UPDATE settings SET
    openclaw_gateway_url = ?, openclaw_gateway_token = ?,
    default_model = ?, max_parallel_executions = ?,
    default_project_directory = ?,
    gsd_depth = ?, gsd_mode = ?, gsd_research_enabled = ?, gsd_plan_check_enabled = ?, gsd_verifier_enabled = ?,
    ralph_max_iterations = ?, ralph_auto_commit = ?, theme = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = 'default' RETURNING *;
