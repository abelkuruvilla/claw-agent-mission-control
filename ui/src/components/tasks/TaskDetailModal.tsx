'use client';

import React, { useState, useEffect, useCallback } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { Card, CardContent } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Textarea } from '@/components/ui/textarea';
import { Input } from '@/components/ui/input';
import { 
  Trash2,
  Edit,
  CheckCircle2,
  XCircle,
  AlertTriangle,
  Info,
  Save,
  X,
  FileText,
  GitBranch,
  ClipboardCheck,
  ScrollText,
  Plus,
  ListTree,
  Bell,
  ArrowRightLeft,
  Send,
  CircleDot,
  Activity,
  RotateCcw,
  ArrowLeft,
  ChevronRight,
} from 'lucide-react';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { TaskModal } from './TaskModal';
import { useAgentsStore, useEventsStore } from '@/stores';
import { useTasksStore } from '@/stores';
import { tasksApi, commentsApi } from '@/services/api';
import { SchedulePicker } from '@/components/ui/SchedulePicker';
import { calculateProgress } from '@/lib/task-utils';
import type { Task, Agent, TaskStatus, Comment } from '@/types';

const STATUS_OPTIONS: { value: TaskStatus; label: string; color: string }[] = [
  { value: 'queued', label: 'Queued', color: '#f97316' },
  { value: 'backlog', label: 'Backlog', color: '#6b7280' },
  { value: 'planning', label: 'Planning', color: '#8b5cf6' },
  { value: 'discussing', label: 'Discussing', color: '#3b82f6' },
  { value: 'executing', label: 'Executing', color: '#10b981' },
  { value: 'verifying', label: 'Verifying', color: '#f59e0b' },
  { value: 'review', label: 'Review', color: '#ec4899' },
  { value: 'done', label: 'Done', color: '#22c55e' },
  { value: 'failed', label: 'Failed', color: '#ef4444' },
];

interface ArtifactField {
  key: keyof Task;
  label: string;
  icon: React.ReactNode;
  placeholder: string;
  rows: number;
  language?: string;
}

const ARTIFACT_FIELDS: ArtifactField[] = [
  {
    key: 'project_md',
    label: 'Project Overview',
    icon: <FileText className="h-4 w-4" />,
    placeholder: '# Project Overview\n\nDescribe the project context, goals, and scope...',
    rows: 10,
    language: 'markdown',
  },
  {
    key: 'requirements_md',
    label: 'Requirements',
    icon: <ClipboardCheck className="h-4 w-4" />,
    placeholder: '# Requirements\n\n## Functional Requirements\n- ...\n\n## Non-Functional Requirements\n- ...',
    rows: 12,
    language: 'markdown',
  },
  {
    key: 'roadmap_md',
    label: 'Roadmap',
    icon: <ScrollText className="h-4 w-4" />,
    placeholder: '# Roadmap\n\n## Phase 1: ...\n## Phase 2: ...',
    rows: 10,
    language: 'markdown',
  },
  {
    key: 'state_md',
    label: 'Current State',
    icon: <Info className="h-4 w-4" />,
    placeholder: '# Current State\n\nDescribe the current state of work, blockers, decisions...',
    rows: 8,
    language: 'markdown',
  },
  {
    key: 'prd_json',
    label: 'PRD (JSON)',
    icon: <FileText className="h-4 w-4" />,
    placeholder: '{\n  "title": "",\n  "user_stories": [],\n  "acceptance_criteria": []\n}',
    rows: 10,
    language: 'json',
  },
  {
    key: 'progress_txt',
    label: 'Progress Log',
    icon: <ScrollText className="h-4 w-4" />,
    placeholder: '[2026-02-18 10:00] Task started...\n[2026-02-18 10:30] Completed research phase...',
    rows: 12,
    language: 'text',
  },
  {
    key: 'quality_checks',
    label: 'Quality Checks',
    icon: <ClipboardCheck className="h-4 w-4" />,
    placeholder: 'Define quality checks: test commands, lint rules, acceptance criteria...',
    rows: 6,
    language: 'text',
  },
];

interface TaskDetailModalProps {
  task: Task;
  isOpen: boolean;
  onClose: () => void;
  onUpdate: (id: string, updates: Partial<Task>) => Promise<void>;
  onDelete: (id: string) => Promise<void>;
  onOpenTask?: (taskId: string) => void;
}

export const TaskDetailModal: React.FC<TaskDetailModalProps> = ({ 
  task, 
  isOpen, 
  onClose,
  onUpdate,
  onDelete,
  onOpenTask,
}) => {
  const agents = useAgentsStore((s) => s.agents);
  const events = useEventsStore((s) => s.events);
  const fetchEvents = useEventsStore((s) => s.fetchEvents);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [loading, setLoading] = useState(false);

  // Artifact inline editing state
  const [editingField, setEditingField] = useState<string | null>(null);
  const [editValue, setEditValue] = useState('');
  const [savingField, setSavingField] = useState(false);

  // Git branch inline editing
  const [editingGitBranch, setEditingGitBranch] = useState(false);
  const [gitBranchValue, setGitBranchValue] = useState('');

  // Subtask state
  const [subtasks, setSubtasks] = useState<Task[]>([]);
  const [subtasksLoading, setSubtasksLoading] = useState(false);
  const [isCreateSubtaskOpen, setIsCreateSubtaskOpen] = useState(false);
  
  // Parent task navigation state
  const [parentTask, setParentTask] = useState<Task | null>(null);
  const [parentTaskLoading, setParentTaskLoading] = useState(false);
  const [breadcrumb, setBreadcrumb] = useState<Task[]>([]);
  
  const createTask = useTasksStore((s) => s.createTask);

  // Comment and approval state
  const [comments, setComments] = useState<Comment[]>([]);
  const [changeRequestComment, setChangeRequestComment] = useState('');
  const [expandedSubtask, setExpandedSubtask] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [retryLoading, setRetryLoading] = useState(false);
  const [retryAtEnabled, setRetryAtEnabled] = useState(false);
  const [retryAt, setRetryAt] = useState('');

  const assignedAgent = task.agent_id 
    ? agents.find(a => a.id === task.agent_id)
    : null;

  const fetchSubtasks = useCallback(async () => {
    setSubtasksLoading(true);
    try {
      const data = await tasksApi.listSubtasks(task.id);
      setSubtasks(data);
    } catch (err) {
      console.error('Failed to fetch subtasks:', err);
    } finally {
      setSubtasksLoading(false);
    }
  }, [task.id]);

  const fetchParentTask = useCallback(async () => {
    if (!task.parent_task_id) {
      setParentTask(null);
      return;
    }
    setParentTaskLoading(true);
    try {
      const data = await tasksApi.get(task.parent_task_id);
      setParentTask(data);
      // Build breadcrumb from parent
      setBreadcrumb(prev => {
        const newCrumb = [...prev, data];
        return newCrumb.slice(-5); // Keep max 5 in breadcrumb
      });
    } catch (err) {
      console.error('Failed to fetch parent task:', err);
      setParentTask(null);
    } finally {
      setParentTaskLoading(false);
    }
  }, [task.parent_task_id]);

  // Navigate to a task (for breadcrumb clicks)
  const [navigateToTaskId, setNavigateToTaskId] = useState<string | null>(null);
  const handleNavigateToTask = useCallback(async (taskId: string) => {
    setNavigateToTaskId(taskId);
  }, []);

  // Effect to handle navigation
  useEffect(() => {
    if (navigateToTaskId && onOpenTask) {
      onOpenTask(navigateToTaskId);
      setNavigateToTaskId(null);
    }
  }, [navigateToTaskId, onOpenTask]);

  useEffect(() => {
    if (isOpen && task.id) {
      fetchEvents({ task_id: task.id });
      fetchSubtasks();
      fetchParentTask();
    }
  }, [isOpen, task.id, fetchEvents, fetchSubtasks, fetchParentTask]);

  // Reset editing state when task or modal changes
  useEffect(() => {
    setEditingField(null);
    setEditingGitBranch(false);
  }, [task.id, isOpen]);

  const taskEvents = events.filter(e => e.task_id === task.id);

  const handleStatusChange = async (newStatus: TaskStatus) => {
    setLoading(true);
    try {
      await onUpdate(task.id, { status: newStatus });
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async () => {
    if (confirm('Are you sure you want to delete this task?')) {
      setLoading(true);
      try {
        await onDelete(task.id);
        onClose();
      } catch (error) {
        console.error('Failed to delete task:', error);
        setLoading(false);
      }
    }
  };

  const handleUpdateTask = async (updates: Partial<Task>) => {
    await onUpdate(task.id, updates);
  };

  const handleCreateSubtask = async (subtaskData: Partial<Task>) => {
    await createTask(subtaskData);
    await fetchSubtasks();
  };

  const fetchComments = useCallback(async (taskId: string) => {
    try {
      const data = await commentsApi.listByTask(taskId);
      setComments(data);
    } catch (err) {
      console.error('Failed to fetch comments:', err);
    }
  }, []);

  const handleApprove = async (subtaskId: string) => {
    setActionLoading(subtaskId);
    try {
      await tasksApi.approveDelegation(subtaskId);
      await fetchSubtasks();
      await fetchEvents({ task_id: task.id });
    } catch (err) {
      console.error('Failed to approve delegation:', err);
    } finally {
      setActionLoading(null);
    }
  };

  const activeStatuses: TaskStatus[] = ['executing', 'planning', 'discussing', 'verifying'];
  const canRetry = activeStatuses.includes(task.status) || ['failed', 'queued', 'backlog'].includes(task.status);

  const handleRetry = async () => {
    setRetryLoading(true);
    try {
      const payload = retryAtEnabled && retryAt
        ? { retry_at: retryAt }
        : undefined;
      const updated = await tasksApi.retry(task.id, payload);
      await onUpdate(task.id, updated);
      await fetchEvents({ task_id: task.id });
      setRetryAtEnabled(false);
      setRetryAt('');
    } catch (err) {
      console.error('Failed to retry task:', err);
    } finally {
      setRetryLoading(false);
    }
  };

  const handleRequestChanges = async (subtaskId: string) => {
    if (!changeRequestComment.trim()) return;
    setActionLoading(subtaskId);
    try {
      await tasksApi.requestChanges(subtaskId, changeRequestComment);
      setChangeRequestComment('');
      setExpandedSubtask(null);
      await fetchSubtasks();
      await fetchEvents({ task_id: task.id });
    } catch (err) {
      console.error('Failed to request changes:', err);
    } finally {
      setActionLoading(null);
    }
  };

  const getSubtaskStatusColor = (status: string) => {
    return STATUS_OPTIONS.find(s => s.value === status)?.color || '#6b7280';
  };

  const isSubtaskApproved = useCallback((subtaskId: string): boolean => {
    return taskEvents.some((e) => {
      if (e.type !== 'delegation_approved') return false;
      try {
        const d = e.details ? (JSON.parse(e.details) as { subtask_id?: string }) : {};
        return d.subtask_id === subtaskId;
      } catch {
        return false;
      }
    });
  }, [taskEvents]);

  const getEventStyle = (type: string): { icon: React.ReactNode; color: string; bg: string } => {
    switch (type) {
      case 'task_created':
        return { icon: <Plus className="h-4 w-4 text-emerald-400" />, color: 'text-emerald-300', bg: 'bg-emerald-500/15' };
      case 'subtask_created':
        return { icon: <ListTree className="h-4 w-4 text-orange-400" />, color: 'text-orange-300', bg: 'bg-orange-500/15' };
      case 'status_changed':
        return { icon: <ArrowRightLeft className="h-4 w-4 text-blue-400" />, color: 'text-blue-300', bg: 'bg-blue-500/15' };
      case 'task_assigned':
        return { icon: <CircleDot className="h-4 w-4 text-purple-400" />, color: 'text-purple-300', bg: 'bg-purple-500/15' };
      case 'agent_notified':
        return { icon: <Send className="h-4 w-4 text-cyan-400" />, color: 'text-cyan-300', bg: 'bg-cyan-500/15' };
      case 'subtask_result_received':
        return { icon: <CheckCircle2 className="h-4 w-4 text-green-400" />, color: 'text-green-300', bg: 'bg-green-500/15' };
      case 'orchestrator_notified':
        return { icon: <Bell className="h-4 w-4 text-amber-400" />, color: 'text-amber-300', bg: 'bg-amber-500/15' };
      case 'orchestrator_acknowledged':
        return { icon: <CheckCircle2 className="h-4 w-4 text-emerald-400" />, color: 'text-emerald-300', bg: 'bg-emerald-500/15' };
      case 'notification_error':
      case 'execution_error':
        return { icon: <XCircle className="h-4 w-4 text-red-400" />, color: 'text-red-300', bg: 'bg-red-500/15' };
      case 'task_failed':
        return { icon: <XCircle className="h-4 w-4 text-red-400" />, color: 'text-red-300', bg: 'bg-red-500/15' };
      case 'task_completed':
        return { icon: <CheckCircle2 className="h-4 w-4 text-green-400" />, color: 'text-green-300', bg: 'bg-green-500/15' };
      case 'verification_passed':
        return { icon: <ClipboardCheck className="h-4 w-4 text-green-400" />, color: 'text-green-300', bg: 'bg-green-500/15' };
      case 'verification_failed':
        return { icon: <AlertTriangle className="h-4 w-4 text-yellow-400" />, color: 'text-yellow-300', bg: 'bg-yellow-500/15' };
      default:
        return { icon: <Info className="h-4 w-4 text-slate-400" />, color: 'text-slate-300', bg: 'bg-slate-500/15' };
    }
  };

  const formatEventType = (type: string): string => {
    return type.replace(/_/g, ' ');
  };

  const formatEventTime = (isoString: string): string => {
    const d = new Date(isoString);
    return d.toLocaleString(undefined, {
      month: 'short', day: 'numeric',
      hour: '2-digit', minute: '2-digit', second: '2-digit',
    });
  };

  const formatEventDetails = (details: string): string => {
    try {
      return JSON.stringify(JSON.parse(details), null, 2);
    } catch {
      return details;
    }
  };

  const startEditingField = useCallback((fieldKey: string) => {
    setEditingField(fieldKey);
    setEditValue((task as any)[fieldKey] || '');
  }, [task]);

  const cancelEditingField = useCallback(() => {
    setEditingField(null);
    setEditValue('');
  }, []);

  const saveField = useCallback(async (fieldKey: string) => {
    setSavingField(true);
    try {
      await onUpdate(task.id, { [fieldKey]: editValue });
      setEditingField(null);
      setEditValue('');
    } catch (error) {
      console.error(`Failed to save ${fieldKey}:`, error);
    } finally {
      setSavingField(false);
    }
  }, [task.id, editValue, onUpdate]);

  const startEditingGitBranch = useCallback(() => {
    setEditingGitBranch(true);
    setGitBranchValue(task.git_branch || '');
  }, [task.git_branch]);

  const saveGitBranch = useCallback(async () => {
    setSavingField(true);
    try {
      await onUpdate(task.id, { git_branch: gitBranchValue } as any);
      setEditingGitBranch(false);
    } catch (error) {
      console.error('Failed to save git branch:', error);
    } finally {
      setSavingField(false);
    }
  }, [task.id, gitBranchValue, onUpdate]);

  const progress = calculateProgress(task);

  const artifactCount = ARTIFACT_FIELDS.filter(f => (task as any)[f.key]).length;

  return (
    <>
      <Dialog open={isOpen} onOpenChange={onClose}>
        <DialogContent className="max-w-[90vw] w-full max-h-[90vh] overflow-y-auto overflow-x-hidden bg-[#161b22] border-[#30363d]">
          <DialogHeader>
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <DialogTitle className="text-xl text-white pr-8">{task.title}</DialogTitle>
                <div className="flex items-center gap-3 mt-3 flex-wrap">
                  <Select
                    value={task.status}
                    onValueChange={(value) => handleStatusChange(value as TaskStatus)}
                    disabled={loading}
                  >
                    <SelectTrigger className="w-[160px] bg-slate-950 border-[#30363d] text-white h-8">
                      <div className="flex items-center gap-2">
                        <div 
                          className="h-2 w-2 rounded-full" 
                          style={{ backgroundColor: STATUS_OPTIONS.find(s => s.value === task.status)?.color }}
                        />
                        <SelectValue />
                      </div>
                    </SelectTrigger>
                    <SelectContent className="bg-slate-950 border-[#30363d]">
                      {STATUS_OPTIONS.map(status => (
                        <SelectItem key={status.value} value={status.value}>
                          <div className="flex items-center gap-2">
                            <div 
                              className="h-2 w-2 rounded-full" 
                              style={{ backgroundColor: status.color }}
                            />
                            <span className="text-white">{status.label}</span>
                          </div>
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <Badge variant="outline">Priority {task.priority}</Badge>
                  {canRetry && (
                    <div className="flex items-center gap-2 flex-wrap">
                      <Button
                        variant="outline"
                        size="sm"
                        className="h-8 border-amber-500/50 text-amber-400 hover:bg-amber-500/10"
                        onClick={handleRetry}
                        disabled={retryLoading}
                      >
                        <RotateCcw className={`h-3.5 w-3.5 mr-1.5 ${retryLoading ? 'animate-spin' : ''}`} />
                        {retryLoading ? 'Retrying…' : (retryAtEnabled && retryAt ? 'Schedule Retry' : 'Retry')}
                      </Button>
                      <div className="flex items-center gap-1">
                        <input
                          type="checkbox"
                          id="retry-at-toggle"
                          checked={retryAtEnabled}
                          onChange={e => setRetryAtEnabled(e.target.checked)}
                          className="h-3 w-3"
                        />
                        <label htmlFor="retry-at-toggle" className="text-xs text-slate-400">Schedule</label>
                      </div>
                      {retryAtEnabled && (
                        <SchedulePicker
                          value={retryAt}
                          onChange={setRetryAt}
                          label="Retry at"
                          className="mt-1"
                        />
                      )}
                    </div>
                  )}
                  {activeStatuses.includes(task.status) && (
                    <span className="text-xs text-slate-400">Stuck? Retry re-notifies the agent.</span>
                  )}
                  {task.scheduled_at && (
                    <div className="flex items-center gap-2">
                      <span className="text-xs text-blue-400">⏰ Scheduled: {new Date(task.scheduled_at).toLocaleString()}</span>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-6 px-2 text-xs text-red-400 hover:text-red-300"
                        onClick={async () => {
                          setLoading(true);
                          try {
                            await onUpdate(task.id, { clear_schedule: true } as any);
                          } finally {
                            setLoading(false);
                          }
                        }}
                        disabled={loading}
                      >
                        ✕ Remove
                      </Button>
                    </div>
                  )}
                  {task.retry_at && (
                    <span className="text-xs text-yellow-400">🔄 Retry scheduled: {new Date(task.retry_at).toLocaleString()}</span>
                  )}
                  {task.git_branch && !editingGitBranch && (
                    <Badge variant="outline" className="text-emerald-400 border-emerald-500/30 gap-1">
                      <GitBranch className="h-3 w-3" />
                      {task.git_branch}
                    </Badge>
                  )}
                </div>
              </div>
            </div>
          </DialogHeader>

          {/* Breadcrumb Navigation */}
          {(parentTask || breadcrumb.length > 0) && (
            <div className="flex items-center gap-1 text-sm text-slate-400 bg-slate-900/50 rounded-lg px-3 py-2 mb-2 overflow-x-auto">
              <button
                onClick={() => {
                  const taskId = breadcrumb[0]?.id || parentTask?.id;
                  if (taskId) onOpenTask?.(taskId);
                }}
                className="hover:text-white hover:underline truncate max-w-[150px]"
                title="Go to parent"
              >
                {parentTaskLoading ? 'Loading...' : (parentTask?.title || 'Parent Task')}
              </button>
              <ChevronRight className="h-4 w-4 shrink-0" />
              <span className="text-white truncate max-w-[200px]">{task.title}</span>
              {breadcrumb.length > 1 && (
                <>
                  <ChevronRight className="h-4 w-4 shrink-0" />
                  <span className="text-slate-500">+{breadcrumb.length - 1} more</span>
                </>
              )}
            </div>
          )}

          <Tabs defaultValue="details" className="w-full min-w-0 overflow-hidden">
            <TabsList className="flex w-full overflow-x-auto shrink-0 gap-1">
              <TabsTrigger value="details" className="flex-1 min-w-fit">Details</TabsTrigger>
              <TabsTrigger value="subtasks" className="flex-1 min-w-fit gap-1">
                Subtasks
                {subtasks.length > 0 && (
                  <span className="ml-1 text-[10px] bg-orange-500/20 text-orange-400 rounded-full px-1.5">
                    {subtasks.length}
                  </span>
                )}
              </TabsTrigger>
              <TabsTrigger value="artifacts" className="flex-1 min-w-fit gap-1">
                Artifacts
                {artifactCount > 0 && (
                  <span className="ml-1 text-[10px] bg-blue-500/20 text-blue-400 rounded-full px-1.5">
                    {artifactCount}
                  </span>
                )}
              </TabsTrigger>
              <TabsTrigger value="events" className="flex-1 min-w-fit gap-1">
                Events
                {taskEvents.length > 0 && (
                  <span className="ml-1 text-[10px] bg-emerald-500/20 text-emerald-400 rounded-full px-1.5">
                    {taskEvents.length}
                  </span>
                )}
              </TabsTrigger>
            </TabsList>

            {/* ===== DETAILS TAB ===== */}
            <TabsContent value="details" className="space-y-4 mt-4 w-full min-w-0 overflow-x-hidden outline-none">
              {/* Parent Task */}
              {parentTask && (
                <div className="rounded-lg border border-[#30363d] bg-slate-950 p-3">
                  <h3 className="text-sm font-medium text-slate-400 mb-2 flex items-center gap-2">
                    <ArrowLeft className="h-4 w-4" />
                    Parent Task
                  </h3>
                  <button
                    onClick={() => onOpenTask?.(parentTask.id)}
                    className="flex items-center gap-2 w-full text-left hover:bg-slate-900/50 -mx-1 px-1 py-1.5 rounded transition-colors"
                  >
                    <div className="flex-1 min-w-0">
                      <p className="text-white font-medium truncate">{parentTask.title}</p>
                      <div className="flex items-center gap-2 mt-1">
                        <Badge 
                          variant="outline" 
                          className="text-xs"
                          style={{ borderColor: STATUS_OPTIONS.find(s => s.value === parentTask.status)?.color }}
                        >
                          {parentTask.status}
                        </Badge>
                        {parentTask.git_branch && (
                          <span className="text-xs text-slate-500 font-mono">{parentTask.git_branch}</span>
                        )}
                      </div>
                    </div>
                    <ArrowRightLeft className="h-4 w-4 text-slate-500" />
                  </button>
                </div>
              )}

              <div>
                <h3 className="text-sm font-medium text-slate-400 mb-2">Description</h3>
                <p className="text-slate-200 whitespace-pre-wrap">{task.description || 'No description provided.'}</p>
              </div>

              <div>
                <div className="flex items-center justify-between mb-2">
                  <h3 className="text-sm font-medium text-slate-400">Progress</h3>
                  <span className="text-sm text-slate-300">{progress}%</span>
                </div>
                <Progress value={progress} className="h-2" />
              </div>

              {/* Git Branch */}
              <div>
                <div className="flex items-center justify-between mb-2">
                  <h3 className="text-sm font-medium text-slate-400 flex items-center gap-2">
                    <GitBranch className="h-4 w-4" />
                    Git Branch
                  </h3>
                  {!editingGitBranch && (
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-6 px-2 text-xs text-slate-400 hover:text-white"
                      onClick={startEditingGitBranch}
                    >
                      <Edit className="h-3 w-3 mr-1" />
                      Edit
                    </Button>
                  )}
                </div>
                {editingGitBranch ? (
                  <div className="flex items-center gap-2">
                    <Input
                      value={gitBranchValue}
                      onChange={(e) => setGitBranchValue(e.target.value)}
                      placeholder="e.g., feature/task-123"
                      className="bg-slate-950 border-[#30363d] text-white h-8 flex-1"
                    />
                    <Button
                      size="sm"
                      className="h-8 px-3"
                      onClick={saveGitBranch}
                      disabled={savingField}
                    >
                      <Save className="h-3 w-3 mr-1" />
                      Save
                    </Button>
                    <Button
                      size="sm"
                      variant="ghost"
                      className="h-8 px-2"
                      onClick={() => setEditingGitBranch(false)}
                    >
                      <X className="h-3 w-3" />
                    </Button>
                  </div>
                ) : (
                  <p className="text-sm text-slate-200 font-mono">
                    {task.git_branch || <span className="text-slate-500 italic">Not set</span>}
                  </p>
                )}
              </div>

              {/* Agent Assignment */}
              <div>
                <h3 className="text-sm font-medium text-slate-400 mb-2">Assigned Agent</h3>
                {assignedAgent ? (
                  <div className="flex items-center gap-3 rounded-lg border border-[#30363d] bg-slate-950 p-3">
                    <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-purple-600">
                      <span className="text-sm font-medium text-white">
                        {assignedAgent.name.charAt(0)}
                      </span>
                    </div>
                    <div>
                      <div className="font-medium text-white">{assignedAgent.name}</div>
                      <div className="text-xs text-slate-400">{assignedAgent.model}</div>
                    </div>
                  </div>
                ) : (
                  <p className="text-sm text-slate-500">No agent assigned</p>
                )}
              </div>

              {/* Metadata */}
              <div className="grid gap-4 md:grid-cols-2">
                <div>
                  <h3 className="text-sm font-medium text-slate-400 mb-1">Created</h3>
                  <p className="text-sm text-slate-200">
                    {new Date(task.created_at).toLocaleString()}
                  </p>
                </div>
                <div>
                  <h3 className="text-sm font-medium text-slate-400 mb-1">Last Updated</h3>
                  <p className="text-sm text-slate-200">
                    {new Date(task.updated_at).toLocaleString()}
                  </p>
                </div>
                {task.started_at && (
                  <div>
                    <h3 className="text-sm font-medium text-slate-400 mb-1">Started</h3>
                    <p className="text-sm text-slate-200">
                      {new Date(task.started_at).toLocaleString()}
                    </p>
                  </div>
                )}
                {task.completed_at && (
                  <div>
                    <h3 className="text-sm font-medium text-slate-400 mb-1">Completed</h3>
                    <p className="text-sm text-slate-200">
                      {new Date(task.completed_at).toLocaleString()}
                    </p>
                  </div>
                )}
              </div>
            </TabsContent>

            {/* ===== SUBTASKS TAB ===== */}
            <TabsContent value="subtasks" className="space-y-4 mt-4 w-full min-w-0 overflow-x-hidden outline-none">
              <div className="flex items-center justify-between gap-3 mb-2 w-full min-w-0">
                <div>
                  <h3 className="text-sm font-medium text-slate-400">
                    Delegated subtasks assigned to specialist agents.
                  </h3>
                  <div className="flex items-center gap-2 mt-1">
                    <Badge variant="outline" className={`text-[10px] ${task.delegation_mode === 'manual' ? 'text-amber-400 border-amber-500/30' : 'text-emerald-400 border-emerald-500/30'}`}>
                      {task.delegation_mode === 'manual' ? 'Manual Approval' : 'Auto Delegation'}
                    </Badge>
                  </div>
                </div>
                <Button
                  size="sm"
                  className="gap-1.5"
                  onClick={() => setIsCreateSubtaskOpen(true)}
                >
                  <Plus className="h-3.5 w-3.5" />
                  Create Subtask
                </Button>
              </div>
              {subtasksLoading ? (
                <p className="text-sm text-slate-500 text-center py-8">Loading subtasks...</p>
              ) : subtasks.length > 0 ? (
                <div className="space-y-3 w-full min-w-0">
                  {subtasks.map((subtask) => {
                    const subtaskAgent = subtask.agent_id
                      ? agents.find(a => a.id === subtask.agent_id)
                      : null;
                    const isExpanded = expandedSubtask === subtask.id;
                    const isDone = subtask.status === 'done' || subtask.status === 'failed';
                    const approved = isSubtaskApproved(subtask.id);
                    const needsApproval = task.delegation_mode === 'manual' && isDone && !approved;
                    return (
                      <div
                        key={subtask.id}
                        className={`rounded-lg border bg-slate-950 overflow-hidden transition-colors ${needsApproval ? 'border-amber-500/40' : 'border-[#30363d]'}`}
                      >
                        {/* Subtask header row */}
                        <div
                          className="flex items-center gap-3 p-3 cursor-pointer hover:bg-[#0d1117] transition-colors w-full min-w-0"
                          onClick={() => {
                            setExpandedSubtask(isExpanded ? null : subtask.id);
                            if (!isExpanded) fetchComments(subtask.id);
                          }}
                        >
                          <ListTree className="h-4 w-4 text-orange-400 flex-shrink-0" />
                          <div className="flex-1 min-w-0">
                            <div className="text-sm font-medium text-white truncate">{subtask.title}</div>
                            <div className="flex items-center gap-2 mt-0.5 flex-wrap">
                              {subtaskAgent ? (
                                <span className="text-xs text-blue-400">{subtaskAgent.name}</span>
                              ) : (
                                <span className="text-xs text-slate-500">Unassigned</span>
                              )}
                            </div>
                          </div>
                          <Badge variant="outline" className="text-xs capitalize flex-shrink-0">
                            <div
                              className="h-1.5 w-1.5 rounded-full mr-1.5"
                              style={{ backgroundColor: getSubtaskStatusColor(subtask.status) }}
                            />
                            {subtask.status}
                          </Badge>
                          {needsApproval && (
                            <Badge className="text-[10px] bg-amber-500/20 text-amber-400 border border-amber-500/30 flex-shrink-0">
                              Awaiting Approval
                            </Badge>
                          )}
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-6 px-2 text-xs text-slate-400 hover:text-white flex-shrink-0"
                            onClick={(e) => {
                              e.stopPropagation();
                              onOpenTask?.(subtask.id);
                            }}
                            title="Open in modal"
                          >
                            <ArrowRightLeft className="h-3 w-3" />
                          </Button>
                        </div>

                        {/* Expanded details */}
                        {isExpanded && (
                          <div className="border-t border-[#30363d] p-4 space-y-3 bg-[#0d1117]">
                            {/* Description */}
                            {subtask.description && (
                              <div>
                                <h4 className="text-xs font-medium text-slate-400 mb-1">Description</h4>
                                <p className="text-xs text-slate-300 whitespace-pre-wrap">{subtask.description}</p>
                              </div>
                            )}

                            {/* Progress / Results */}
                            {subtask.progress_txt && (
                              <div>
                                <h4 className="text-xs font-medium text-slate-400 mb-1">Progress / Results</h4>
                                <pre className="text-xs text-slate-300 bg-[#161b22] border border-[#30363d] rounded p-2 max-h-[200px] overflow-y-auto whitespace-pre-wrap font-mono">
                                  {subtask.progress_txt}
                                </pre>
                              </div>
                            )}

                            {/* Comments */}
                            {comments.length > 0 && (
                              <div>
                                <h4 className="text-xs font-medium text-slate-400 mb-1">Comments</h4>
                                <div className="space-y-2">
                                  {comments.filter(c => c.task_id === subtask.id).map(comment => (
                                    <div key={comment.id} className="bg-[#161b22] border border-[#30363d] rounded p-2">
                                      <div className="flex items-center gap-2 mb-1">
                                        <span className="text-[11px] font-medium text-blue-400">{comment.author}</span>
                                        <span className="text-[10px] text-slate-600">
                                          {new Date(comment.created_at).toLocaleString()}
                                        </span>
                                      </div>
                                      <p className="text-xs text-slate-300">{comment.content}</p>
                                    </div>
                                  ))}
                                </div>
                              </div>
                            )}

                            {/* Approval actions for manual mode */}
                            {needsApproval && (
                              <div className="pt-2 space-y-3">
                                <Separator className="bg-[#30363d]" />
                                <div className="flex items-center gap-2">
                                  <Button
                                    size="sm"
                                    className="gap-1.5 bg-emerald-600 hover:bg-emerald-700"
                                    onClick={() => handleApprove(subtask.id)}
                                    disabled={actionLoading === subtask.id}
                                  >
                                    <CheckCircle2 className="h-3.5 w-3.5" />
                                    {actionLoading === subtask.id ? 'Approving...' : 'Approve & Continue'}
                                  </Button>
                                  <Button
                                    size="sm"
                                    variant="outline"
                                    className="gap-1.5 text-amber-400 border-amber-500/30 hover:bg-amber-500/10"
                                    onClick={() => setExpandedSubtask(subtask.id)}
                                  >
                                    <Edit className="h-3.5 w-3.5" />
                                    Request Changes
                                  </Button>
                                </div>
                                <div className="flex gap-2">
                                  <Textarea
                                    value={changeRequestComment}
                                    onChange={(e) => setChangeRequestComment(e.target.value)}
                                    placeholder="Describe what changes are needed..."
                                    rows={2}
                                    className="bg-[#161b22] border-[#30363d] text-slate-200 text-xs resize-none flex-1"
                                  />
                                  <Button
                                    size="sm"
                                    variant="destructive"
                                    className="self-end"
                                    onClick={() => handleRequestChanges(subtask.id)}
                                    disabled={!changeRequestComment.trim() || actionLoading === subtask.id}
                                  >
                                    <Send className="h-3.5 w-3.5" />
                                  </Button>
                                </div>
                              </div>
                            )}
                          </div>
                        )}
                      </div>
                    );
                  })}
                </div>
              ) : (
                <Card className="border-[#30363d] bg-slate-950">
                  <CardContent className="p-8">
                    <div className="text-center">
                      <ListTree className="h-8 w-8 text-slate-600 mx-auto mb-3" />
                      <p className="text-sm text-slate-500 mb-2">No subtasks yet.</p>
                      <p className="text-xs text-slate-600 mb-4">
                        Create subtasks to delegate work to specialist agents.
                      </p>
                      <Button
                        size="sm"
                        className="gap-1.5"
                        onClick={() => setIsCreateSubtaskOpen(true)}
                      >
                        <Plus className="h-3.5 w-3.5" />
                        Create Subtask
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              )}
            </TabsContent>

            {/* ===== ARTIFACTS TAB ===== */}
            <TabsContent value="artifacts" className="space-y-4 mt-4 w-full min-w-0 overflow-x-hidden outline-none">
              <div className="text-sm text-slate-400 mb-2">
                GSD planning artifacts and execution documents. Click <strong className="text-slate-300">Edit</strong> on any field to modify.
              </div>
              {ARTIFACT_FIELDS.map((field) => {
                const value = (task as any)[field.key] as string | undefined;
                const isEditing = editingField === field.key;

                return (
                  <div key={field.key} className="rounded-lg border border-[#30363d] bg-slate-950 overflow-hidden">
                    {/* Field Header */}
                    <div className="flex items-center justify-between px-4 py-2.5 bg-[#0d1117] border-b border-[#30363d]">
                      <div className="flex items-center gap-2 text-sm font-medium text-slate-300">
                        {field.icon}
                        {field.label}
                        {field.language && (
                          <span className="text-[10px] text-slate-500 uppercase tracking-wider ml-1">
                            {field.language}
                          </span>
                        )}
                      </div>
                      <div className="flex items-center gap-1">
                        {isEditing ? (
                          <>
                            <Button
                              size="sm"
                              className="h-6 px-2 text-xs"
                              onClick={() => saveField(field.key as string)}
                              disabled={savingField}
                            >
                              <Save className="h-3 w-3 mr-1" />
                              {savingField ? 'Saving...' : 'Save'}
                            </Button>
                            <Button
                              size="sm"
                              variant="ghost"
                              className="h-6 px-2 text-xs text-slate-400"
                              onClick={cancelEditingField}
                            >
                              <X className="h-3 w-3" />
                            </Button>
                          </>
                        ) : (
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-6 px-2 text-xs text-slate-400 hover:text-white"
                            onClick={() => startEditingField(field.key as string)}
                          >
                            <Edit className="h-3 w-3 mr-1" />
                            Edit
                          </Button>
                        )}
                      </div>
                    </div>

                    {/* Field Content */}
                    <div className="p-4">
                      {isEditing ? (
                        <Textarea
                          value={editValue}
                          onChange={(e) => setEditValue(e.target.value)}
                          placeholder={field.placeholder}
                          rows={field.rows}
                          className="bg-[#0d1117] border-[#30363d] text-slate-200 font-mono text-sm resize-y min-h-[100px]"
                          autoFocus
                        />
                      ) : value ? (
                        <pre className="text-sm text-slate-200 whitespace-pre-wrap break-words font-mono leading-relaxed max-h-[300px] overflow-y-auto">
                          {value}
                        </pre>
                      ) : (
                        <p className="text-sm text-slate-500 italic">
                          No content yet. Click Edit to add {field.label.toLowerCase()}.
                        </p>
                      )}
                    </div>
                  </div>
                );
              })}
            </TabsContent>

            {/* ===== EVENTS TAB ===== */}
            <TabsContent value="events" className="space-y-3 mt-4 w-full min-w-0 overflow-x-hidden outline-none">
              {taskEvents.length > 0 ? (
                <div className="space-y-1">
                  {taskEvents.map((event, idx) => {
                    const { icon, color, bg } = getEventStyle(event.type);
                    const isLast = idx === taskEvents.length - 1;
                    return (
                      <div key={event.id} className="flex gap-3 relative">
                        {/* Timeline connector */}
                        {!isLast && (
                          <div className="absolute left-[15px] top-[28px] bottom-0 w-px bg-[#30363d]" />
                        )}
                        {/* Icon */}
                        <div
                          className={`flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full ${bg} z-10`}
                        >
                          {icon}
                        </div>
                        {/* Content */}
                        <div className="flex-1 min-w-0 pb-4">
                          <div className="flex items-start justify-between gap-2">
                            <div className="min-w-0">
                              <p className={`text-sm font-medium ${color}`}>
                                {event.message}
                              </p>
                              <div className="flex items-center gap-2 mt-0.5 flex-wrap">
                                <span className="text-[11px] text-slate-500 font-mono">
                                  {formatEventType(event.type)}
                                </span>
                                {event.agent_id && (
                                  <span className="text-[11px] text-blue-400/70">
                                    {agents.find(a => a.id === event.agent_id)?.name || event.agent_id}
                                  </span>
                                )}
                              </div>
                            </div>
                            <span className="text-[11px] text-slate-600 whitespace-nowrap flex-shrink-0">
                              {formatEventTime(event.created_at)}
                            </span>
                          </div>
                          {event.details && (
                            <pre className="mt-1.5 text-[11px] text-slate-500 bg-[#0d1117] border border-[#30363d] rounded px-2 py-1 overflow-x-auto font-mono">
                              {formatEventDetails(event.details)}
                            </pre>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              ) : (
                <Card className="border-[#30363d] bg-slate-950">
                  <CardContent className="p-8">
                    <div className="text-center">
                      <Activity className="h-8 w-8 text-slate-600 mx-auto mb-3" />
                      <p className="text-sm text-slate-500 mb-1">No events yet</p>
                      <p className="text-xs text-slate-600">
                        Events will appear here as the task progresses — subtask creation, agent notifications, status changes, and more.
                      </p>
                    </div>
                  </CardContent>
                </Card>
              )}
            </TabsContent>
          </Tabs>

          <Separator className="bg-[#30363d]" />

          {/* Actions */}
          <div className="flex items-center justify-end gap-2">
            <Button
              onClick={() => setIsEditModalOpen(true)}
              size="sm"
              variant="outline"
              className="gap-2"
              disabled={loading}
            >
              <Edit className="h-4 w-4" />
              Edit
            </Button>
            <Button onClick={handleDelete} size="sm" variant="destructive" disabled={loading}>
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* Edit Modal */}
      <TaskModal
        isOpen={isEditModalOpen}
        onClose={() => setIsEditModalOpen(false)}
        onSubmit={handleUpdateTask}
        task={task}
      />

      {/* Create Subtask Modal */}
      <TaskModal
        isOpen={isCreateSubtaskOpen}
        onClose={() => setIsCreateSubtaskOpen(false)}
        onSubmit={handleCreateSubtask}
        parentTaskId={task.id}
        parentTaskTitle={task.title}
      />
    </>
  );
};
