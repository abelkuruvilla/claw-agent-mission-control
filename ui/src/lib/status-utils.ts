import type { EventType } from '@/types';

/** Status color map for agent statuses */
const AGENT_STATUS_COLORS: Record<string, string> = {
  working: 'bg-green-500',
  idle: 'bg-slate-500',
  paused: 'bg-yellow-500',
  error: 'bg-red-500',
};

/** Status color map for project statuses */
const PROJECT_STATUS_COLORS: Record<string, string> = {
  active: 'bg-green-500',
  'on-hold': 'bg-yellow-500',
  completed: 'bg-blue-500',
  archived: 'bg-slate-500',
};

/** Badge variant map for agent statuses */
const AGENT_BADGE_VARIANTS: Record<string, 'default' | 'secondary' | 'outline' | 'destructive'> = {
  working: 'default',
  idle: 'secondary',
  paused: 'outline',
  error: 'destructive',
};

/** Badge variant map for project statuses */
const PROJECT_BADGE_VARIANTS: Record<string, 'default' | 'secondary' | 'outline' | 'destructive'> = {
  active: 'default',
  'on-hold': 'outline',
  completed: 'secondary',
  archived: 'secondary',
};

/** Priority color classes */
const PRIORITY_COLORS: Record<number, string> = {
  1: 'bg-red-500',
  2: 'bg-orange-500',
  3: 'bg-yellow-500',
  4: 'bg-blue-500',
  5: 'bg-gray-500',
};

/**
 * Returns the Tailwind background color class for an agent status.
 * O(1) lookup instead of switch statement.
 */
export function getAgentStatusColor(status: string): string {
  return AGENT_STATUS_COLORS[status] || 'bg-slate-500';
}

/**
 * Returns the Tailwind background color class for a project status.
 * O(1) lookup instead of switch statement.
 */
export function getProjectStatusColor(status: string): string {
  return PROJECT_STATUS_COLORS[status] || 'bg-slate-500';
}

/**
 * Returns badge variant for agent status.
 */
export function getAgentBadgeVariant(status: string): 'default' | 'secondary' | 'outline' | 'destructive' {
  return AGENT_BADGE_VARIANTS[status] || 'secondary';
}

/**
 * Returns badge variant for project status.
 */
export function getProjectBadgeVariant(status: string): 'default' | 'secondary' | 'outline' | 'destructive' {
  return PROJECT_BADGE_VARIANTS[status] || 'secondary';
}

/**
 * Returns the Tailwind background color class for a priority level.
 */
export function getPriorityColor(priority: number): string {
  return PRIORITY_COLORS[priority] || 'bg-gray-500';
}

/**
 * Returns badge variant for an event type.
 */
export function getEventBadgeVariant(type: EventType): 'default' | 'destructive' | 'secondary' | 'outline' {
  if (type.includes('completed') || type.includes('passed')) return 'default';
  if (type.includes('failed') || type.includes('error')) return 'destructive';
  if (type.includes('created')) return 'secondary';
  return 'outline';
}

/**
 * Formats an event type string into a human-readable label.
 * e.g. 'task_completed' -> 'Task Completed'
 */
export function formatEventType(type: string): string {
  return type
    .split('_')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
}
