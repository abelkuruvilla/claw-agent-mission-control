'use client';

import React, { useState, useEffect } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { useAgentsStore, useProjectsStore } from '@/stores';
import type { Task } from '@/types';
import { SchedulePicker } from '@/components/ui/SchedulePicker';

interface TaskModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (taskData: Partial<Task>) => Promise<void>;
  task?: Task;
  initialStatus?: string;
  parentTaskId?: string;
  parentTaskTitle?: string;
}

export const TaskModal: React.FC<TaskModalProps> = ({ 
  isOpen, 
  onClose,
  onSubmit,
  task,
  initialStatus,
  parentTaskId,
  parentTaskTitle,
}) => {
  const agents = useAgentsStore((s) => s.agents);
  const projects = useProjectsStore((s) => s.projects);
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState({
    title: '',
    description: '',
    priority: 3,
    agent_id: 'unassigned',
    project_id: 'none',
    status: 'backlog',
    git_branch: '',
    quality_checks: '',
    delegation_mode: 'auto' as 'auto' | 'manual',
    schedule_enabled: false,
    scheduled_at: '',
  });

  useEffect(() => {
    if (task) {
      setFormData({
        title: task.title,
        description: task.description,
        priority: task.priority,
        agent_id: task.agent_id || 'unassigned',
        project_id: task.project_id || 'none',
        status: task.status,
        git_branch: task.git_branch || '',
        quality_checks: task.quality_checks || '',
        delegation_mode: task.delegation_mode || 'auto',
        schedule_enabled: !!task.scheduled_at,
        scheduled_at: task.scheduled_at || '',
      });
    } else {
      setFormData({
        title: '',
        description: '',
        priority: 3,
        agent_id: 'unassigned',
        project_id: 'none',
        status: initialStatus || 'backlog',
        git_branch: '',
        quality_checks: '',
        delegation_mode: 'auto',
        schedule_enabled: false,
        scheduled_at: '',
      });
    }
  }, [task, isOpen, initialStatus]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    
    try {
      const isUnassigning = formData.agent_id === 'unassigned';
      const wasScheduled = !!task?.scheduled_at;
      const taskData: Partial<Task> & { clear_schedule?: boolean } = {
        title: formData.title,
        description: formData.description,
        priority: formData.priority,
        agent_id: isUnassigning ? '' : formData.agent_id,
        project_id: formData.project_id === 'none' ? '' : formData.project_id,
        parent_task_id: parentTaskId || undefined,
        status: (isUnassigning ? 'backlog' : formData.status) as any,
        git_branch: formData.git_branch || undefined,
        quality_checks: formData.quality_checks || undefined,
        delegation_mode: formData.delegation_mode,
        scheduled_at: formData.schedule_enabled && formData.scheduled_at
          ? formData.scheduled_at
          : undefined,
        clear_schedule: wasScheduled && !formData.schedule_enabled ? true : undefined,
      };

      await onSubmit(taskData);
      onClose();
    } catch (error) {
      console.error('Failed to save task:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-2xl max-h-[85vh] flex flex-col overflow-hidden bg-[#161b22] border-[#30363d]">
        <DialogHeader className="shrink-0">
          <DialogTitle className="text-white">
            {task ? 'Edit Task' : parentTaskId ? 'Create Subtask' : 'Create New Task'}
          </DialogTitle>
          <DialogDescription className="text-slate-400">
            {parentTaskId && parentTaskTitle ? (
              <>Subtask of: <span className="text-blue-400 font-medium">{parentTaskTitle}</span></>
            ) : (
              'Define the task details. Tasks use GSD for planning and Ralph Loop for execution.'
            )}
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4 overflow-y-auto min-h-0 flex-1 pr-1">
          <div className="space-y-2">
            <Label htmlFor="title" className="text-slate-200">Task Title</Label>
            <Input
              id="title"
              value={formData.title}
              onChange={(e) => setFormData({ ...formData, title: e.target.value })}
              placeholder="e.g., Build authentication system"
              required
              className="bg-slate-950 border-[#30363d] text-white"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="description" className="text-slate-200">Description</Label>
            <Textarea
              id="description"
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              placeholder="Describe what needs to be done..."
              rows={4}
              required
              className="bg-slate-950 border-[#30363d] text-white max-h-[30vh] resize-y"
            />
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="project" className="text-slate-200">
                Project
                <span className="text-xs text-slate-500 ml-2">(optional)</span>
              </Label>
              <Select
                value={formData.project_id}
                onValueChange={(value) => setFormData({ ...formData, project_id: value })}
              >
                <SelectTrigger className="bg-slate-950 border-[#30363d] text-white">
                  <SelectValue placeholder="Select a project..." />
                </SelectTrigger>
                <SelectContent className="bg-slate-950 border-[#30363d]">
                  <SelectItem value="none">No Project</SelectItem>
                  {projects.map(project => (
                    <SelectItem key={project.id} value={project.id}>
                      {project.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label htmlFor="priority" className="text-slate-200">Priority</Label>
              <Select
                value={formData.priority.toString()}
                onValueChange={(value) => setFormData({ ...formData, priority: parseInt(value) })}
              >
                <SelectTrigger className="bg-slate-950 border-[#30363d] text-white">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="bg-slate-950 border-[#30363d]">
                  <SelectItem value="1">1 - Critical (Red)</SelectItem>
                  <SelectItem value="2">2 - High (Orange)</SelectItem>
                  <SelectItem value="3">3 - Medium (Yellow)</SelectItem>
                  <SelectItem value="4">4 - Low (Blue)</SelectItem>
                  <SelectItem value="5">5 - Trivial (Gray)</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="agent" className="text-slate-200">
              Assign to Agent
              <span className="text-xs text-slate-500 ml-2">(optional)</span>
            </Label>
            <Select
              value={formData.agent_id}
              onValueChange={(value) =>
                setFormData({
                  ...formData,
                  agent_id: value,
                  ...(value === 'unassigned' ? { status: 'backlog' as const } : {}),
                })
              }
            >
              <SelectTrigger className="bg-slate-950 border-[#30363d] text-white">
                <SelectValue placeholder="Select an agent..." />
              </SelectTrigger>
              <SelectContent className="bg-slate-950 border-[#30363d]">
                <SelectItem value="unassigned">Unassigned</SelectItem>
                {agents.map(agent => (
                  <SelectItem key={agent.id} value={agent.id}>
                    {agent.name} ({agent.status})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="git_branch" className="text-slate-200">
                Git Branch
                <span className="text-xs text-slate-500 ml-2">(optional)</span>
              </Label>
              <Input
                id="git_branch"
                value={formData.git_branch}
                onChange={(e) => setFormData({ ...formData, git_branch: e.target.value })}
                placeholder="e.g., feature/task-123"
                className="bg-slate-950 border-[#30363d] text-white font-mono text-sm"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="quality_checks" className="text-slate-200">
                Quality Checks
                <span className="text-xs text-slate-500 ml-2">(optional)</span>
              </Label>
              <Input
                id="quality_checks"
                value={formData.quality_checks}
                onChange={(e) => setFormData({ ...formData, quality_checks: e.target.value })}
                placeholder="e.g., npm test, npm run lint"
                className="bg-slate-950 border-[#30363d] text-white font-mono text-sm"
              />
            </div>
          </div>

          {/* Delegation Mode */}
          {!parentTaskId && (
            <div className="space-y-2">
              <Label className="text-slate-200">
                Delegation Mode
                <span className="text-xs text-slate-500 ml-2">(subtask approval)</span>
              </Label>
              <Select
                value={formData.delegation_mode}
                onValueChange={(value) => setFormData({ ...formData, delegation_mode: value as 'auto' | 'manual' })}
              >
                <SelectTrigger className="bg-slate-950 border-[#30363d] text-white">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="bg-slate-950 border-[#30363d]">
                  <SelectItem value="auto">
                    <span className="text-white">Auto — agents delegate automatically</span>
                  </SelectItem>
                  <SelectItem value="manual">
                    <span className="text-white">Manual — require human approval between steps</span>
                  </SelectItem>
                </SelectContent>
              </Select>
              <p className="text-xs text-slate-500">
                {formData.delegation_mode === 'manual'
                  ? 'You will review each subtask result and approve before the orchestrator continues.'
                  : 'The orchestrator is notified automatically when subtasks complete.'}
              </p>
            </div>
          )}

          {/* Schedule for later */}
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <input
                type="checkbox"
                id="schedule-toggle"
                checked={formData.schedule_enabled}
                onChange={e => setFormData({ ...formData, schedule_enabled: e.target.checked })}
              />
              <label htmlFor="schedule-toggle" className="text-sm font-medium text-slate-300">
                {task?.scheduled_at && !formData.schedule_enabled
                  ? '⚡ Execute immediately (removes schedule)'
                  : 'Schedule for later'}
              </label>
            </div>
            {task?.scheduled_at && !formData.schedule_enabled && (
              <p className="text-xs text-amber-400 ml-5">
                Currently scheduled: {new Date(task.scheduled_at).toLocaleString()}
              </p>
            )}
            {formData.schedule_enabled && (
              <SchedulePicker
                value={formData.scheduled_at}
                onChange={(iso) => setFormData({ ...formData, scheduled_at: iso })}
                label="Execute at"
              />
            )}
          </div>

          <DialogFooter className="shrink-0 pt-4 sticky bottom-0 bg-[#161b22]">
            <Button type="button" variant="outline" onClick={onClose} disabled={loading}>
              Cancel
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? 'Saving...' : task ? 'Update Task' : 'Create Task'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
};
