'use client';

import React from 'react';
import { useDrag } from 'react-dnd';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { GripVertical, CheckCircle, Folder } from 'lucide-react';
import { calculateProgress, getStoryCounts } from '@/lib/task-utils';
import { getPriorityColor } from '@/lib/status-utils';
import type { Task, Agent, Project } from '@/types';

const ITEM_TYPE = 'TASK';

/** Status badge configuration - static lookup to avoid repeated conditionals */
const STATUS_BADGE_CONFIG: Record<string, { text: string; className: string; icon?: boolean }> = {
  queued: { text: 'Queued', className: 'text-orange-400 border-orange-500/30' },
  planning: { text: 'Planning', className: 'text-purple-400 border-purple-500/30' },
  discussing: { text: 'Discussing', className: 'text-blue-400 border-blue-500/30' },
  executing: { text: 'Executing', className: 'text-green-400 border-green-500/30' },
  verifying: { text: 'Verifying', className: 'text-amber-400 border-amber-500/30' },
  review: { text: 'Review', className: 'text-cyan-400 border-cyan-500/30' },
  done: { text: 'Done', className: 'text-green-400 border-green-500/30', icon: true },
  failed: { text: 'Failed', className: 'text-red-400 border-red-500/30' },
  backlog: { text: 'Backlog', className: 'text-slate-400 border-slate-500/30' },
};

interface TaskCardProps {
  task: Task;
  agent?: Agent;
  project?: Project;
  onClick: () => void;
}

/**
 * Memoized task card component.
 * Only re-renders when task, agent, or onClick reference changes.
 */
export const TaskCard: React.FC<TaskCardProps> = React.memo(({ task, agent, project, onClick }) => {
  const [{ isDragging }, drag] = useDrag(() => ({
    type: ITEM_TYPE,
    item: { id: task.id },
    collect: (monitor) => ({
      isDragging: monitor.isDragging(),
    }),
  }), [task.id]);

  const progress = calculateProgress(task);
  const { total: storiesTotal, passed: storiesPassed } = getStoryCounts(task);

  const badgeConfig = STATUS_BADGE_CONFIG[task.status];

  return (
    <div
      ref={drag as any}
      onClick={onClick}
      className={`cursor-pointer transition-opacity w-full ${isDragging ? 'opacity-50' : 'opacity-100'}`}
    >
      <Card className="border-[#30363d] bg-[#161b22] hover:border-blue-500 transition-colors h-full min-h-[140px] max-h-[160px] overflow-hidden flex flex-col">
        <CardContent className="p-3 flex-1 flex flex-col min-h-0 gap-2 overflow-hidden">
          {/* Shrinkable area: title + description + story progress */}
          <div className="flex-1 min-h-0 flex flex-col gap-1.5 overflow-hidden">
            <div className="flex items-start gap-2 shrink-0">
              <GripVertical className="h-4 w-4 text-slate-600 mt-0.5 flex-shrink-0" />
              <div className="flex-1 min-w-0">
                <h3 className="font-medium text-white text-sm line-clamp-2 leading-tight">
                  {task.title}
                </h3>
              </div>
              <div className={`h-2 w-2 rounded-full flex-shrink-0 mt-1 ${getPriorityColor(task.priority)}`} />
            </div>
            {task.description && (
              <p className="text-xs text-slate-400 line-clamp-1 shrink-0">
                {task.description}
              </p>
            )}
            {(task.status === 'executing' || task.status === 'verifying') && storiesTotal > 0 && (
              <div className="space-y-1 shrink-0">
                <div className="flex items-center justify-between text-xs">
                  <span className="text-slate-500">Stories</span>
                  <span className="text-slate-400">{storiesPassed}/{storiesTotal}</span>
                </div>
                <Progress value={progress} className="h-1" />
              </div>
            )}
          </div>

          {/* Metadata row: always visible, never shrunk */}
          <div className="flex items-center justify-between gap-2 shrink-0 min-h-[22px]">
            <div className="flex flex-1 items-center gap-2 min-w-0 overflow-hidden">
              {project && (
                <span
                  className="flex flex-1 items-center gap-1 min-w-0 text-[11px] text-slate-400 truncate"
                  title={project.name}
                >
                  <Folder className="h-3 w-3 shrink-0 text-slate-500" aria-hidden />
                  <span
                    className="h-1.5 w-1.5 rounded-full shrink-0"
                    style={{ backgroundColor: project.color || '#6b7280' }}
                  />
                  <span className="truncate">{project.name}</span>
                </span>
              )}
            </div>
            {agent && (
              <span className="text-[11px] text-slate-500 truncate shrink-0 max-w-[72px]" title={agent.name}>
                {agent.name}
              </span>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
});

TaskCard.displayName = 'TaskCard';
