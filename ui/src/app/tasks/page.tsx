'use client';

import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { useDrop } from 'react-dnd';
import { DndProvider } from 'react-dnd';
import { HTML5Backend } from 'react-dnd-html5-backend';
import { Button } from '@/components/ui/button';
import { Plus, ChevronDown, ChevronRight, Filter, X, ArrowUp, ArrowDown, Search } from 'lucide-react';
import { TaskCard } from '@/components/tasks/TaskCard';
import { TaskModal } from '@/components/tasks/TaskModal';
import { TaskDetailModal } from '@/components/tasks/TaskDetailModal';
import { useTasksStore, useAgentsStore, useProjectsStore } from '@/stores';
import { useIsMobile } from '@/hooks/useIsMobile';
import type { Task, TaskStatus, Project } from '@/types';

const ITEM_TYPE = 'TASK';

const COLUMNS: { status: TaskStatus; title: string; color: string }[] = [
  { status: 'queued', title: 'Queued', color: '#f97316' },
  { status: 'backlog', title: 'Backlog', color: '#6b7280' },
  { status: 'planning', title: 'Planning', color: '#8b5cf6' },
  { status: 'discussing', title: 'Discussing', color: '#3b82f6' },
  { status: 'executing', title: 'Executing', color: '#10b981' },
  { status: 'verifying', title: 'Verifying', color: '#f59e0b' },
  { status: 'review', title: 'Review', color: '#ec4899' },
  { status: 'done', title: 'Done', color: '#22c55e' },
  { status: 'failed', title: 'Failed', color: '#ef4444' },
];

// Fuzzy match: substring or sequential character match (case-insensitive)
function fuzzyMatch(query: string, text: string): boolean {
  if (!query) return true;
  const q = query.toLowerCase();
  const t = text.toLowerCase();
  // Substring match
  if (t.includes(q)) return true;
  // Sequential character match (fuzzy)
  let qi = 0;
  for (let ti = 0; ti < t.length && qi < q.length; ti++) {
    if (t[ti] === q[qi]) qi++;
  }
  return qi === q.length;
}

interface ColumnProps {
  status: TaskStatus;
  title: string;
  color: string;
  tasks: Task[];
  agents: any[];
  projects: Project[];
  onTaskClick: (task: Task) => void;
  onDrop: (taskId: string, newStatus: TaskStatus) => void;
  onAddTask: (status: TaskStatus) => void;
}

/**
 * Desktop Kanban Column (with drag & drop).
 * Memoized to prevent re-renders when other columns change.
 */
const Column: React.FC<ColumnProps> = React.memo(({ 
  status, 
  title, 
  color,
  tasks, 
  agents,
  projects,
  onTaskClick, 
  onDrop,
  onAddTask,
}) => {
  const [{ isOver }, drop] = useDrop(() => ({
    accept: ITEM_TYPE,
    drop: (item: { id: string }) => onDrop(item.id, status),
    collect: (monitor) => ({
      isOver: monitor.isOver(),
    }),
  }), [status, onDrop]);

  return (
    <div
      ref={drop as any}
      className={`flex flex-col rounded-lg border bg-[#161b22] p-3 transition-colors w-[280px] min-w-[280px] max-w-[280px] shrink-0 h-full max-h-full min-h-0 ${
        isOver ? 'border-blue-500 bg-blue-950/20' : 'border-[#30363d]'
      }`}
    >
      <div className="mb-3 flex items-center justify-between shrink-0">
        <div className="flex items-center gap-2">
          <div 
            className="h-2 w-2 rounded-full" 
            style={{ backgroundColor: color }}
          />
          <h2 className="font-semibold text-white text-sm">
            {title}
          </h2>
          <span className="text-sm text-slate-500">({tasks.length})</span>
        </div>
        <button
          onClick={() => onAddTask(status)}
          className="flex h-6 w-6 items-center justify-center rounded hover:bg-[#30363d] text-slate-400 hover:text-white transition-colors"
          title={`Add task to ${title}`}
        >
          <Plus className="h-4 w-4" />
        </button>
      </div>
      <div className="flex-1 min-h-0 space-y-2 overflow-y-auto">
        {tasks.length === 0 ? (
          <div className="flex h-24 items-center justify-center rounded-lg border-2 border-dashed border-[#30363d]">
            <p className="text-sm text-slate-600">Empty</p>
          </div>
        ) : (
          tasks.map(task => {
            const agent = task.agent_id ? agents.find(a => a.id === task.agent_id) : undefined;
            const project = task.project_id ? projects.find(p => p.id === task.project_id) : undefined;
            return (
              <TaskCard
                key={task.id}
                task={task}
                agent={agent}
                project={project}
                onClick={() => onTaskClick(task)}
              />
            );
          })
        )}
      </div>
    </div>
  );
});
Column.displayName = 'Column';

// Mobile Accordion Section (collapsible)
interface AccordionSectionProps {
  status: TaskStatus;
  title: string;
  color: string;
  tasks: Task[];
  agents: any[];
  projects: Project[];
  onTaskClick: (task: Task) => void;
  onAddTask: (status: TaskStatus) => void;
  isExpanded: boolean;
  onToggle: () => void;
}

const AccordionSection: React.FC<AccordionSectionProps> = React.memo(({
  status,
  title,
  color,
  tasks,
  agents,
  projects,
  onTaskClick,
  onAddTask,
  isExpanded,
  onToggle,
}) => {
  return (
    <div className="border border-[#30363d] rounded-lg overflow-hidden bg-[#161b22]">
      {/* Header */}
      <button
        onClick={onToggle}
        className="w-full flex items-center justify-between p-3 hover:bg-[#21262d] transition-colors"
      >
        <div className="flex items-center gap-3">
          {isExpanded ? (
            <ChevronDown className="h-4 w-4 text-slate-400" />
          ) : (
            <ChevronRight className="h-4 w-4 text-slate-400" />
          )}
          <div 
            className="h-2.5 w-2.5 rounded-full" 
            style={{ backgroundColor: color }}
          />
          <span className="font-medium text-white">{title}</span>
          <span className="text-sm text-slate-500 bg-[#21262d] px-2 py-0.5 rounded-full">
            {tasks.length}
          </span>
        </div>
        <button
          onClick={(e) => {
            e.stopPropagation();
            onAddTask(status);
          }}
          className="flex h-7 w-7 items-center justify-center rounded hover:bg-[#30363d] text-slate-400 hover:text-white transition-colors"
        >
          <Plus className="h-4 w-4" />
        </button>
      </button>

      {/* Content - only rendered when expanded */}
      {isExpanded && (
        <div className="border-t border-[#30363d] p-2 space-y-2">
          {tasks.length === 0 ? (
            <div className="flex h-16 items-center justify-center">
              <p className="text-sm text-slate-600">No tasks</p>
            </div>
          ) : (
            tasks.map(task => {
              const agent = task.agent_id ? agents.find(a => a.id === task.agent_id) : undefined;
              const project = task.project_id ? projects.find(p => p.id === task.project_id) : undefined;
              return (
                <TaskCard
                  key={task.id}
                  task={task}
                  agent={agent}
                  project={project}
                  onClick={() => onTaskClick(task)}
                />
              );
            })
          )}
        </div>
      )}
    </div>
  );
});
AccordionSection.displayName = 'AccordionSection';

export default function TasksPage() {
  const tasks = useTasksStore((s) => s.tasks);
  const fetchTasks = useTasksStore((s) => s.fetchTasks);
  const createTask = useTasksStore((s) => s.createTask);
  const updateTaskStatus = useTasksStore((s) => s.updateTaskStatus);
  const updateTask = useTasksStore((s) => s.updateTask);
  const deleteTask = useTasksStore((s) => s.deleteTask);
  const agents = useAgentsStore((s) => s.agents);
  const fetchAgents = useAgentsStore((s) => s.fetchAgents);
  const projects = useProjectsStore((s) => s.projects);
  const fetchProjects = useProjectsStore((s) => s.fetchProjects);

  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [createTaskStatus, setCreateTaskStatus] = useState<TaskStatus>('backlog');
  const [selectedTask, setSelectedTask] = useState<Task | null>(null);
  const [projectFilter, setProjectFilter] = useState<string>('all');
  const [agentFilter, setAgentFilter] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState<string>('');
  const [sortBy, setSortBy] = useState<'created_at' | 'updated_at' | 'name' | 'priority'>('created_at');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');
  const isMobile = useIsMobile();
  
  // Track expanded sections for mobile accordion
  const [expandedSections, setExpandedSections] = useState<Set<TaskStatus>>(
    new Set(['queued', 'backlog', 'executing'])
  );

  // Load data on mount (tasks, agents, projects so Create/Edit Task modal has full lists)
  useEffect(() => {
    fetchTasks();
    fetchAgents();
    fetchProjects();
  }, [fetchTasks, fetchAgents, fetchProjects]);

  const filteredTasks = useMemo(() => {
    return tasks.filter(task => {
      if (projectFilter !== 'all' && task.project_id !== projectFilter) return false;
      if (agentFilter !== 'all') {
        if (agentFilter === 'unassigned') return !task.agent_id;
        if (task.agent_id !== agentFilter) return false;
      }
      if (searchQuery && !fuzzyMatch(searchQuery, task.title)) return false;
      return true;
    });
  }, [tasks, projectFilter, agentFilter, searchQuery]);

  const hasActiveFilters = projectFilter !== 'all' || agentFilter !== 'all' || searchQuery !== '';

  /**
   * Group tasks by status in a single pass - O(n) instead of O(n * columns).
   */
  const tasksByStatus = useMemo(() => {
    const grouped = new Map<TaskStatus, Task[]>();
    for (const col of COLUMNS) {
      grouped.set(col.status, []);
    }
    for (const task of filteredTasks) {
      const bucket = grouped.get(task.status);
      if (bucket) {
        bucket.push(task);
      }
    }
    // Sort each bucket
    const compareFn = (a: Task, b: Task): number => {
      let result = 0;
      switch (sortBy) {
        case 'updated_at':
          result = new Date(a.updated_at).getTime() - new Date(b.updated_at).getTime();
          break;
        case 'name':
          result = a.title.localeCompare(b.title);
          break;
        case 'priority':
          result = (a.priority ?? 0) - (b.priority ?? 0);
          break;
        default: // created_at
          result = new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
      }
      return sortOrder === 'asc' ? result : -result;
    };
    for (const [, tasks] of grouped) {
      tasks.sort(compareFn);
    }
    return grouped;
  }, [filteredTasks, sortBy, sortOrder]);

  const clearFilters = useCallback(() => {
    setProjectFilter('all');
    setAgentFilter('all');
    setSearchQuery('');
  }, []);

  const handleAddTaskToColumn = useCallback((status: TaskStatus) => {
    setCreateTaskStatus(status);
    setIsCreateModalOpen(true);
  }, []);

  const handleDrop = useCallback(async (taskId: string, newStatus: TaskStatus) => {
    try {
      await updateTaskStatus(taskId, newStatus);
    } catch (error) {
      console.error('Failed to update task status:', error);
    }
  }, [updateTaskStatus]);

  const handleCreateTask = useCallback(async (taskData: Partial<Task>) => {
    try {
      await createTask(taskData);
    } catch (error) {
      console.error('Failed to create task:', error);
      throw error;
    }
  }, [createTask]);

  const handleUpdateTask = useCallback(async (id: string, updates: Partial<Task>) => {
    try {
      await updateTask(id, updates);
    } catch (error) {
      console.error('Failed to update task:', error);
      throw error;
    }
  }, [updateTask]);

  const handleDeleteTask = useCallback(async (id: string) => {
    try {
      await deleteTask(id);
    } catch (error) {
      console.error('Failed to delete task:', error);
      throw error;
    }
  }, [deleteTask]);

  const toggleSection = useCallback((status: TaskStatus) => {
    setExpandedSections(prev => {
      const next = new Set(prev);
      if (next.has(status)) {
        next.delete(status);
      } else {
        next.add(status);
      }
      return next;
    });
  }, []);

  const handleTaskClick = useCallback((task: Task) => {
    setSelectedTask(task);
  }, []);

  // Desktop content wrapped in DndProvider - fills viewport; columns scroll internally
  const desktopContent = (
    <DndProvider backend={HTML5Backend}>
      <div className="flex-1 min-h-0 overflow-x-auto overflow-y-hidden px-6 py-4">
        <div className="flex gap-4 h-full min-h-0" style={{ minWidth: 'max-content' }}>
          {COLUMNS.map(column => (
            <Column
              key={column.status}
              status={column.status}
              title={column.title}
              color={column.color}
              tasks={tasksByStatus.get(column.status) || []}
              agents={agents}
              projects={projects}
              onTaskClick={handleTaskClick}
              onDrop={handleDrop}
              onAddTask={handleAddTaskToColumn}
            />
          ))}
        </div>
      </div>
    </DndProvider>
  );

  // Mobile content without DndProvider (saves memory)
  const mobileContent = (
    <div className="flex-1 min-h-0 overflow-y-auto px-2 py-4 space-y-2">
      {COLUMNS.map(column => (
        <AccordionSection
          key={column.status}
          status={column.status}
          title={column.title}
          color={column.color}
          tasks={tasksByStatus.get(column.status) || []}
          agents={agents}
          projects={projects}
          onTaskClick={handleTaskClick}
          onAddTask={handleAddTaskToColumn}
          isExpanded={expandedSections.has(column.status)}
          onToggle={() => toggleSection(column.status)}
        />
      ))}
    </div>
  );

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex-none px-2 md:px-6 py-4 border-b border-[#30363d]">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-xl md:text-2xl font-bold text-white">Tasks</h1>
            <p className="mt-1 text-xs md:text-sm text-slate-400 hidden sm:block">
              Manage and track task execution across agents
            </p>
          </div>
          <Button onClick={() => handleAddTaskToColumn('backlog')} className="gap-2" size={isMobile === true ? "sm" : "default"}>
            <Plus className="h-4 w-4" />
            <span className="hidden sm:inline">Create Task</span>
            <span className="sm:hidden">Add</span>
          </Button>
        </div>

        {/* Filters */}
        <div className="flex items-center gap-3 mt-3 flex-wrap">
          <Filter className="h-4 w-4 text-slate-500 shrink-0" />

          {/* Search input */}
          <div className="relative">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-slate-500 pointer-events-none" />
            <input
              type="text"
              placeholder="Search tasks..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="bg-[#0d1117] border border-[#30363d] text-sm text-slate-300 placeholder-slate-600 rounded-md pl-8 pr-3 py-1.5 focus:outline-none focus:border-blue-500 transition-colors min-w-[180px]"
            />
          </div>

          <select
            value={projectFilter}
            onChange={(e) => setProjectFilter(e.target.value)}
            className="bg-[#0d1117] border border-[#30363d] text-sm text-slate-300 rounded-md px-2.5 py-1.5 focus:outline-none focus:border-blue-500 transition-colors appearance-none cursor-pointer min-w-[140px]"
          >
            <option value="all">All Projects</option>
            {projects.map(p => (
              <option key={p.id} value={p.id}>{p.name}</option>
            ))}
          </select>

          <select
            value={agentFilter}
            onChange={(e) => setAgentFilter(e.target.value)}
            className="bg-[#0d1117] border border-[#30363d] text-sm text-slate-300 rounded-md px-2.5 py-1.5 focus:outline-none focus:border-blue-500 transition-colors appearance-none cursor-pointer min-w-[140px]"
          >
            <option value="all">All Agents</option>
            <option value="unassigned">Unassigned</option>
            {agents.map(a => (
              <option key={a.id} value={a.id}>{a.name}</option>
            ))}
          </select>

          {/* Sort By */}
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as 'created_at' | 'updated_at' | 'name' | 'priority')}
            className="bg-[#0d1117] border border-[#30363d] text-sm text-slate-300 rounded-md px-2.5 py-1.5 focus:outline-none focus:border-blue-500 transition-colors appearance-none cursor-pointer min-w-[130px]"
          >
            <option value="created_at">Created At</option>
            <option value="updated_at">Updated At</option>
            <option value="name">Name</option>
            <option value="priority">Priority</option>
          </select>

          {/* Sort Order Toggle */}
          <button
            onClick={() => setSortOrder(prev => prev === 'asc' ? 'desc' : 'asc')}
            className="flex items-center gap-1 text-xs text-slate-400 hover:text-white transition-colors px-2 py-1.5 rounded-md hover:bg-[#21262d] border border-[#30363d]"
            title={`Currently ${sortOrder === 'asc' ? 'ascending' : 'descending'} — click to toggle`}
          >
            {sortOrder === 'asc' ? <ArrowUp className="h-3.5 w-3.5" /> : <ArrowDown className="h-3.5 w-3.5" />}
            {sortOrder === 'asc' ? 'Asc' : 'Desc'}
          </button>

          {hasActiveFilters && (
            <button
              onClick={clearFilters}
              className="flex items-center gap-1 text-xs text-slate-400 hover:text-white transition-colors px-2 py-1.5 rounded-md hover:bg-[#21262d]"
            >
              <X className="h-3 w-3" />
              Clear
            </button>
          )}

          {hasActiveFilters && (
            <span className="text-xs text-slate-500">
              {filteredTasks.length} of {tasks.length} tasks
            </span>
          )}
        </div>
      </div>

      {/* Content: Only mount DndProvider on desktop. 
          Wait for hydration (isMobile !== undefined) before deciding - 
          default to mobile content to avoid HTML5 drag-drop errors on touch devices */}
      {isMobile === undefined ? (
        <div className="flex-1 flex items-center justify-center">
          <div className="text-slate-500">Loading...</div>
        </div>
      ) : isMobile ? mobileContent : desktopContent}

      {/* Modals */}
      <TaskModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        onSubmit={handleCreateTask}
        initialStatus={createTaskStatus}
      />

      {selectedTask && (
        <TaskDetailModal
          task={tasks.find((t) => t.id === selectedTask.id) ?? selectedTask}
          isOpen={!!selectedTask}
          onClose={() => setSelectedTask(null)}
          onUpdate={handleUpdateTask}
          onDelete={handleDeleteTask}
          onOpenTask={(taskId) => {
            const task = tasks.find((t) => t.id === taskId);
            if (task) {
              setSelectedTask(task);
            }
          }}
        />
      )}
    </div>
  );
}