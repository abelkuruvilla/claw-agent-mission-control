import type { Task } from '@/types';

/**
 * Calculates task progress as a percentage (0-100).
 * Uses story counts first, falls back to stories array, then phases.
 * Shared utility to avoid duplicating this logic across components.
 */
export function calculateProgress(task: Task): number {
  if (task.status === 'done') return 100;
  if (task.status === 'failed') return 0;

  // Use story counts if available (Ralph execution phase)
  if (task.stories_total && task.stories_total > 0) {
    return Math.round(((task.stories_passed || 0) / task.stories_total) * 100);
  }

  // Fall back to stories array if available
  if (task.stories && task.stories.length > 0) {
    const total = task.stories.length;
    const passed = task.stories.filter(s => s.passes).length;
    return Math.round((passed / total) * 100);
  }

  // Fall back to phases (GSD planning progress)
  if (task.phases && task.phases.length > 0) {
    const total = task.phases.length;
    const completed = task.phases.filter(p => p.status === 'done').length;
    return Math.round((completed / total) * 100);
  }

  return 0;
}

/**
 * Returns the total and passed story counts for a task.
 * Prefers summary counts, falls back to array computation.
 */
export function getStoryCounts(task: Task): { total: number; passed: number } {
  const total = task.stories_total || (task.stories?.length || 0);
  const passed = task.stories_passed || (task.stories?.filter(s => s.passes).length || 0);
  return { total, passed };
}
