export type AgentStatus = 'idle' | 'working' | 'paused' | 'error';
export type TaskStatus = 'queued' | 'backlog' | 'planning' | 'discussing' | 'executing' | 'verifying' | 'review' | 'done' | 'failed';
// Note: TaskApproach removed - all tasks use GSD for planning and Ralph for execution
export type EventType = 
  | 'task_created' | 'task_completed' | 'task_assigned' | 'task_started'
  | 'task_failed' | 'phase_started' | 'phase_completed' | 'phase_failed'
  | 'story_passed' | 'story_failed' | 'agent_spawned' | 'execution_error'
  | 'verification_passed' | 'verification_failed' | 'commit_created'
  | 'subtask_created' | 'status_changed' | 'agent_notified'
  | 'subtask_result_received' | 'orchestrator_notified'
  | 'orchestrator_acknowledged' | 'notification_error'
  | 'pending_approval' | 'delegation_approved' | 'changes_requested'
  | 'task_queued' | 'task_dequeued';

export type DelegationMode = 'auto' | 'manual';

export interface Agent {
  id: string;
  name: string;
  description: string;
  status: AgentStatus;
  model: string;
  workspace_path?: string;
  agent_dir_path?: string;
  soul_md?: string;
  identity_md?: string;
  agents_md?: string;
  user_md?: string;
  tools_md?: string;
  heartbeat_md?: string;
  memory_md?: string;
  mention_patterns?: string[];
  current_task_id?: string;
  active_session_key?: string;
  created_at: string;
  updated_at: string;
}

export interface Task {
  id: string;
  title: string;
  description: string;
  agent_id?: string;
  project_id?: string;
  parent_task_id?: string;
  status: TaskStatus;
  priority: number;
  git_branch?: string;
  project_md?: string;
  requirements_md?: string;
  roadmap_md?: string;
  state_md?: string;
  prd_json?: string;
  progress_txt?: string;
  quality_checks?: string;
  delegation_mode?: DelegationMode;
  phases?: Phase[];
  stories?: Story[];
  subtasks?: Task[];
  created_at: string;
  updated_at: string;
  started_at?: string;
  completed_at?: string;
  scheduled_at?: string;
  retry_at?: string;
  // Story progress for UI display
  stories_total?: number;
  stories_passed?: number;
}
// Note: No "approach" field - all tasks use GSD for planning and Ralph for execution

export interface Phase {
  id: string;
  task_id: string;
  sequence: number;
  title: string;
  description: string;
  status: 'pending' | 'discussing' | 'planning' | 'executing' | 'verifying' | 'done' | 'failed';
  context_md?: string;
  research_md?: string;
  plan_md?: string;
  summary_md?: string;
  uat_md?: string;
  verification_result?: string;
  created_at: string;
  updated_at: string;
}

export interface Story {
  id: string;
  task_id: string;
  sequence: number;
  title: string;
  description: string;
  passes: boolean;
  priority: number;
  acceptance_criteria: string[];
  iterations: number;
  last_error?: string;
  created_at: string;
  updated_at: string;
}

export interface Event {
  id: string;
  task_id?: string;
  agent_id?: string;
  type: EventType;
  message: string;
  details?: string;
  created_at: string;
}

export interface Comment {
  id: string;
  task_id: string;
  author: string;
  content: string;
  created_at: string;
}

export interface Project {
  id: string;
  name: string;
  description: string;
  status: 'active' | 'on-hold' | 'completed' | 'archived';
  color?: string;
  location?: string;
  created_at: string;
  updated_at: string;
}

export interface Settings {
  id: string;
  openclaw_gateway_url: string;
  openclaw_gateway_token: string;
  default_model: string;
  max_parallel_executions: number;
  default_project_directory: string;
  // GSD Settings (Planning Phase)
  gsd_depth: string;
  gsd_mode: string;
  gsd_research_enabled: boolean;
  gsd_plan_check_enabled: boolean;
  gsd_verifier_enabled: boolean;
  // Ralph Settings (Execution Phase)
  ralph_max_iterations: number;
  ralph_auto_commit: boolean;
  theme: string;
  updated_at: string;
}
// Note: No "default_approach" - all tasks use GSD for planning and Ralph for execution

export interface ApiResponse<T> {
  data: T;
  meta?: {
    total?: number;
    page?: number;
    per_page?: number;
    timestamp: string;
  };
}

export interface ApiError {
  error: {
    code: string;
    message: string;
    details?: Record<string, unknown>;
  };
}

export interface ChatSession {
  id: string;
  agent_id: string;
  openclaw_session_key?: string;
  status: 'active' | 'ended';
  started_at: string;
  ended_at?: string;
  message_count: number;
}

export interface ChatMessage {
  id: string;
  session_id: string;
  role: 'user' | 'agent';
  content: string;
  created_at: string;
}
