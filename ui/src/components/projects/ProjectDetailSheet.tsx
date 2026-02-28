'use client';

import React from 'react';
import { useProjectsStore } from '@/stores/projects';
import { useTasksStore } from '@/stores/tasks';
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Trash2, FolderKanban, Folder, Edit } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';
import { getProjectStatusColor } from '@/lib/status-utils';
import type { Project } from '@/types';

interface ProjectDetailSheetProps {
  project: Project | null;
  isOpen: boolean;
  onClose: () => void;
  onEdit?: () => void;
}

export function ProjectDetailSheet({ project, isOpen, onClose, onEdit }: ProjectDetailSheetProps) {
  const deleteProject = useProjectsStore((s) => s.deleteProject);
  const tasks = useTasksStore((s) => s.tasks);

  if (!project) return null;

  const projectTasks = tasks.filter(t => t.project_id === project.id);

  const handleDelete = async () => {
    if (confirm(`Are you sure you want to delete ${project.name}?`)) {
      await deleteProject(project.id);
      onClose();
    }
  };

  const tasksByStatus = {
    backlog: projectTasks.filter(t => t.status === 'backlog' || t.status === 'queued').length,
    inProgress: projectTasks.filter(t => ['planning', 'discussing', 'executing', 'verifying'].includes(t.status)).length,
    done: projectTasks.filter(t => t.status === 'done').length,
    failed: projectTasks.filter(t => t.status === 'failed').length,
  };

  return (
    <Sheet open={isOpen} onOpenChange={onClose}>
      <SheetContent className="w-[500px] sm:max-w-[500px] bg-[#161b22] border-[#30363d] overflow-hidden flex flex-col">
        <SheetHeader>
          <div className="flex items-center gap-3">
            <div 
              className="flex h-12 w-12 items-center justify-center rounded-lg"
              style={{ backgroundColor: project.color || '#6366f1' }}
            >
              <FolderKanban className="h-6 w-6 text-white" />
            </div>
            <div className="flex-1">
              <SheetTitle className="text-white flex items-center gap-2">
                {project.name}
                <div className={`h-2 w-2 rounded-full ${getProjectStatusColor(project.status)}`} />
              </SheetTitle>
              <SheetDescription className="text-slate-400">
                {project.description}
              </SheetDescription>
            </div>
          </div>
          <div className="flex gap-2 pt-2">
            <Badge variant="outline" className="capitalize">{project.status}</Badge>
            {onEdit && (
              <Button variant="outline" size="sm" onClick={onEdit}>
                <Edit className="h-4 w-4 mr-1" /> Edit
              </Button>
            )}
            <Button variant="destructive" size="sm" onClick={handleDelete}>
              <Trash2 className="h-4 w-4 mr-1" /> Delete
            </Button>
          </div>
        </SheetHeader>

        <ScrollArea className="flex-1 mt-4">
          <div className="space-y-4">
            {/* Stats */}
            <div className="grid grid-cols-2 gap-4">
              <Card className="border-slate-800 bg-slate-900">
                <CardContent className="pt-4">
                  <div className="text-2xl font-bold text-white">{projectTasks.length}</div>
                  <div className="text-sm text-slate-400">Total Tasks</div>
                </CardContent>
              </Card>
              <Card className="border-slate-800 bg-slate-900">
                <CardContent className="pt-4">
                  <div className="text-2xl font-bold text-green-500">{tasksByStatus.done}</div>
                  <div className="text-sm text-slate-400">Completed</div>
                </CardContent>
              </Card>
            </div>

            {/* Task Breakdown */}
            <Card className="border-slate-800 bg-slate-900">
              <CardHeader className="pb-2">
                <CardTitle className="text-sm text-white">Task Breakdown</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-slate-400">Backlog</span>
                  <span className="text-white">{tasksByStatus.backlog}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-slate-400">In Progress</span>
                  <span className="text-white">{tasksByStatus.inProgress}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-slate-400">Done</span>
                  <span className="text-white">{tasksByStatus.done}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-slate-400">Failed</span>
                  <span className="text-white">{tasksByStatus.failed}</span>
                </div>
              </CardContent>
            </Card>

            {/* Details */}
            <Card className="border-slate-800 bg-slate-900">
              <CardHeader className="pb-2">
                <CardTitle className="text-sm text-white">Details</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2 text-sm">
                {project.location && (
                  <div className="flex items-center gap-2">
                    <Folder className="h-4 w-4 text-slate-400" />
                    <span className="text-slate-300 font-mono text-xs">{project.location}</span>
                  </div>
                )}
                <div>
                  <span className="text-slate-400">Created: </span>
                  <span className="text-slate-300">
                    {formatDistanceToNow(new Date(project.created_at), { addSuffix: true })}
                  </span>
                </div>
                <div>
                  <span className="text-slate-400">Updated: </span>
                  <span className="text-slate-300">
                    {formatDistanceToNow(new Date(project.updated_at), { addSuffix: true })}
                  </span>
                </div>
              </CardContent>
            </Card>

            {/* Recent Tasks */}
            <Card className="border-slate-800 bg-slate-900">
              <CardHeader className="pb-2">
                <CardTitle className="text-sm text-white">Recent Tasks</CardTitle>
              </CardHeader>
              <CardContent>
                {projectTasks.length === 0 ? (
                  <p className="text-sm text-slate-500">No tasks in this project</p>
                ) : (
                  <div className="space-y-2">
                    {projectTasks.slice(0, 5).map(task => (
                      <div key={task.id} className="flex items-center justify-between py-2 border-b border-slate-800 last:border-0">
                        <span className="text-sm text-slate-300 truncate">{task.title}</span>
                        <Badge variant="outline" className="text-xs">{task.status}</Badge>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </ScrollArea>
      </SheetContent>
    </Sheet>
  );
}
