import type { 
  Agent, Task, Event, Settings, Project, Phase, Story,
  ApiResponse, ApiError, ChatSession, ChatMessage, Comment
} from '@/types';

const API_BASE = '/api/v1';

async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const error: ApiError = await response.json();
    throw new Error(error.error?.message || 'API Error');
  }
  const data = await response.json();
  return data.data || data;
}

// Agents API
export const agentsApi = {
  list: async (): Promise<Agent[]> => {
    const res = await fetch(`${API_BASE}/agents`);
    return handleResponse<Agent[]>(res);
  },
  
  get: async (id: string): Promise<Agent> => {
    const res = await fetch(`${API_BASE}/agents/${id}`);
    return handleResponse<Agent>(res);
  },
  
  create: async (agent: Partial<Agent>): Promise<Agent> => {
    const res = await fetch(`${API_BASE}/agents`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(agent),
    });
    return handleResponse<Agent>(res);
  },
  
  update: async (id: string, updates: Partial<Agent>): Promise<Agent> => {
    const res = await fetch(`${API_BASE}/agents/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(updates),
    });
    return handleResponse<Agent>(res);
  },
  
  delete: async (id: string): Promise<void> => {
    await fetch(`${API_BASE}/agents/${id}`, { method: 'DELETE' });
  },

  getQueue: async (id: string): Promise<{ agent_id: string; queue_depth: number; tasks: Task[] }> => {
    const res = await fetch(`${API_BASE}/agents/${id}/queue`);
    return handleResponse<{ agent_id: string; queue_depth: number; tasks: Task[] }>(res);
  },

  dequeueNext: async (id: string): Promise<{ agent_id: string; task: Task | null; remaining_queue: number }> => {
    const res = await fetch(`${API_BASE}/agents/${id}/queue/next`, { method: 'POST' });
    return handleResponse<{ agent_id: string; task: Task | null; remaining_queue: number }>(res);
  },
};

// Tasks API
export const tasksApi = {
  list: async (params?: { status?: string; agent_id?: string }): Promise<Task[]> => {
    const searchParams = new URLSearchParams();
    if (params?.status) searchParams.set('status', params.status);
    if (params?.agent_id) searchParams.set('agent_id', params.agent_id);
    const res = await fetch(`${API_BASE}/tasks?${searchParams}`);
    return handleResponse<Task[]>(res);
  },
  
  get: async (id: string): Promise<Task> => {
    const res = await fetch(`${API_BASE}/tasks/${id}?include=phases,stories`);
    return handleResponse<Task>(res);
  },
  
  create: async (task: Partial<Task>): Promise<Task> => {
    const res = await fetch(`${API_BASE}/tasks`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(task),
    });
    return handleResponse<Task>(res);
  },
  
  update: async (id: string, updates: Partial<Task>): Promise<Task> => {
    const res = await fetch(`${API_BASE}/tasks/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(updates),
    });
    return handleResponse<Task>(res);
  },
  
  updateStatus: async (id: string, status: string): Promise<Task> => {
    const res = await fetch(`${API_BASE}/tasks/${id}/status`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ status }),
    });
    return handleResponse<Task>(res);
  },

  /** Re-notify the assigned agent (resets retry count, sets status to backlog). Use when a task is stuck. */
  retry: async (id: string, body?: { retry_at?: string }): Promise<Task> => {
    const res = await fetch(`${API_BASE}/tasks/${id}/retry`, {
      method: 'POST',
      headers: body ? { 'Content-Type': 'application/json' } : undefined,
      body: body ? JSON.stringify(body) : undefined,
    });
    return handleResponse<Task>(res);
  },
  
  delete: async (id: string): Promise<void> => {
    await fetch(`${API_BASE}/tasks/${id}`, { method: 'DELETE' });
  },

  listSubtasks: async (parentId: string): Promise<Task[]> => {
    const res = await fetch(`${API_BASE}/tasks/${parentId}/subtasks`);
    return handleResponse<Task[]>(res);
  },

  start: async (id: string): Promise<void> => {
    await fetch(`${API_BASE}/tasks/${id}/start`, { method: 'POST' });
  },
  
  stop: async (id: string): Promise<void> => {
    await fetch(`${API_BASE}/tasks/${id}/stop`, { method: 'POST' });
  },

  approveDelegation: async (subtaskId: string): Promise<{ status: string }> => {
    const res = await fetch(`${API_BASE}/tasks/${subtaskId}/approve`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
    });
    return handleResponse<{ status: string }>(res);
  },

  requestChanges: async (subtaskId: string, comment: string): Promise<{ status: string }> => {
    const res = await fetch(`${API_BASE}/tasks/${subtaskId}/request-changes`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ comment }),
    });
    return handleResponse<{ status: string }>(res);
  },
};

// Comments API
export const commentsApi = {
  listByTask: async (taskId: string): Promise<Comment[]> => {
    const res = await fetch(`${API_BASE}/tasks/${taskId}/comments`);
    return handleResponse<Comment[]>(res);
  },

  create: async (taskId: string, author: string, content: string): Promise<Comment> => {
    const res = await fetch(`${API_BASE}/tasks/${taskId}/comments`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ author, content }),
    });
    return handleResponse<Comment>(res);
  },

  delete: async (commentId: string): Promise<void> => {
    await fetch(`${API_BASE}/comments/${commentId}`, { method: 'DELETE' });
  },
};

// Events API
export const eventsApi = {
  list: async (params?: { task_id?: string; agent_id?: string; limit?: number }): Promise<Event[]> => {
    const searchParams = new URLSearchParams();
    if (params?.task_id) searchParams.set('task_id', params.task_id);
    if (params?.agent_id) searchParams.set('agent_id', params.agent_id);
    if (params?.limit) searchParams.set('limit', String(params.limit));
    const res = await fetch(`${API_BASE}/events?${searchParams}`);
    return handleResponse<Event[]>(res);
  },
};

// Settings API
export const settingsApi = {
  get: async (): Promise<Settings> => {
    const res = await fetch(`${API_BASE}/settings`);
    return handleResponse<Settings>(res);
  },
  
  update: async (updates: Partial<Settings>): Promise<Settings> => {
    const res = await fetch(`${API_BASE}/settings`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(updates),
    });
    return handleResponse<Settings>(res);
  },
  
  testConnection: async (): Promise<{ connected: boolean }> => {
    const res = await fetch(`${API_BASE}/settings/test-connection`, { method: 'POST' });
    return handleResponse<{ connected: boolean }>(res);
  },
};

// Projects API
export const projectsApi = {
  list: async (): Promise<Project[]> => {
    const res = await fetch(`${API_BASE}/projects`);
    return handleResponse<Project[]>(res);
  },
  
  get: async (id: string): Promise<Project> => {
    const res = await fetch(`${API_BASE}/projects/${id}`);
    return handleResponse<Project>(res);
  },
  
  create: async (project: Partial<Project>): Promise<Project> => {
    const res = await fetch(`${API_BASE}/projects`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(project),
    });
    return handleResponse<Project>(res);
  },
  
  update: async (id: string, updates: Partial<Project>): Promise<Project> => {
    const res = await fetch(`${API_BASE}/projects/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(updates),
    });
    return handleResponse<Project>(res);
  },
  
  delete: async (id: string): Promise<void> => {
    await fetch(`${API_BASE}/projects/${id}`, { method: 'DELETE' });
  },
};

// Status API
export const statusApi = {
  get: async () => {
    const res = await fetch(`${API_BASE}/status`);
    return handleResponse(res);
  },
  
  health: async () => {
    const res = await fetch('/health');
    return handleResponse(res);
  },
};

// Models API - fetches configured models from OpenClaw
export interface ModelConfig {
  id: string;
  alias?: string;
}

export const modelsApi = {
  list: async (): Promise<ModelConfig[]> => {
    const res = await fetch(`${API_BASE}/models`);
    const data = await res.json();
    return data.data || [];
  },
};

// Chat API
export const chatApi = {
  startSession: async (agentId: string): Promise<ChatSession> => {
    const res = await fetch(`${API_BASE}/agents/${agentId}/sessions`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
    });
    return handleResponse<ChatSession>(res);
  },

  endSession: async (agentId: string, sessionId: string): Promise<ChatSession> => {
    const res = await fetch(`${API_BASE}/agents/${agentId}/sessions/${sessionId}`, {
      method: 'DELETE',
    });
    return handleResponse<ChatSession>(res);
  },

  listSessions: async (agentId: string): Promise<ChatSession[]> => {
    const res = await fetch(`${API_BASE}/agents/${agentId}/sessions`);
    return handleResponse<ChatSession[]>(res);
  },

  getMessages: async (agentId: string, sessionId: string): Promise<ChatMessage[]> => {
    const res = await fetch(`${API_BASE}/agents/${agentId}/sessions/${sessionId}/messages`);
    return handleResponse<ChatMessage[]>(res);
  },

  sendMessage: async (agentId: string, sessionId: string, content: string): Promise<ChatMessage> => {
    const res = await fetch(`${API_BASE}/agents/${agentId}/sessions/${sessionId}/messages`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ content }),
    });
    return handleResponse<ChatMessage>(res);
  },
};

// Consolidated API export
export const api = {
  agents: agentsApi,
  tasks: tasksApi,
  events: eventsApi,
  settings: settingsApi,
  projects: projectsApi,
  comments: commentsApi,
  status: statusApi,
  models: modelsApi,
  chat: chatApi,
};
