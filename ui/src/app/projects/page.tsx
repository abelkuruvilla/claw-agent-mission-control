'use client';

import { useEffect, useState, useMemo, useCallback } from 'react';
import { useProjectsStore } from '@/stores/projects';
import { useTasksStore } from '@/stores/tasks';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { Plus, Search, FolderKanban, ChevronRight } from 'lucide-react';
import { ProjectModal } from '@/components/projects/ProjectModal';
import { ProjectDetailSheet } from '@/components/projects/ProjectDetailSheet';
import { getProjectStatusColor, getProjectBadgeVariant } from '@/lib/status-utils';
import type { Project } from '@/types';

export default function ProjectsPage() {
  const projects = useProjectsStore((s) => s.projects);
  const fetchProjects = useProjectsStore((s) => s.fetchProjects);
  const loading = useProjectsStore((s) => s.loading);
  const tasks = useTasksStore((s) => s.tasks);
  const fetchTasks = useTasksStore((s) => s.fetchTasks);

  const [searchQuery, setSearchQuery] = useState('');
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedProject, setSelectedProject] = useState<Project | null>(null);

  useEffect(() => {
    fetchProjects();
    fetchTasks();
  }, [fetchProjects, fetchTasks]);

  // Memoize filtered projects
  const filteredProjects = useMemo(() => {
    if (!searchQuery) return projects;
    const lower = searchQuery.toLowerCase();
    return projects.filter(project =>
      project.name.toLowerCase().includes(lower) ||
      project.description.toLowerCase().includes(lower)
    );
  }, [projects, searchQuery]);

  /**
   * Pre-compute stats for all projects in a single pass over tasks.
   * Replaces the old getProjectStats() which filtered the entire tasks
   * array per project per render (O(projects * tasks) -> O(tasks)).
   */
  const projectStatsMap = useMemo(() => {
    const statsMap = new Map<string, { total: number; done: number; progress: number }>();
    const countMap = new Map<string, { total: number; done: number }>();

    // Single pass over all tasks
    for (const task of tasks) {
      if (!task.project_id) continue;
      let entry = countMap.get(task.project_id);
      if (!entry) {
        entry = { total: 0, done: 0 };
        countMap.set(task.project_id, entry);
      }
      entry.total++;
      if (task.status === 'done') entry.done++;
    }

    // Convert counts to stats with progress
    for (const [projectId, counts] of countMap) {
      statsMap.set(projectId, {
        ...counts,
        progress: counts.total > 0 ? Math.round((counts.done / counts.total) * 100) : 0,
      });
    }

    return statsMap;
  }, [tasks]);

  const handleOpenModal = useCallback(() => setIsModalOpen(true), []);
  const handleCloseModal = useCallback(() => setIsModalOpen(false), []);
  const handleSelectProject = useCallback((project: Project) => setSelectedProject(project), []);
  const handleCloseSheet = useCallback(() => setSelectedProject(null), []);

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-slate-400">Loading projects...</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-[#f0f6fc]">Projects</h1>
          <p className="mt-2 text-[#8b949e]">
            Organize tasks by project
          </p>
        </div>
        <Button onClick={handleOpenModal} className="gap-2">
          <Plus className="h-4 w-4" />
          New Project
        </Button>
      </div>

      {/* Search */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[#8b949e]" />
        <Input
          placeholder="Search projects..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="pl-10 bg-[#0d1117] border-[#30363d] text-[#f0f6fc]"
        />
      </div>

      {/* Projects Grid */}
      {filteredProjects.length === 0 ? (
        <Card className="border-[#30363d] bg-[#161b22]">
          <CardContent className="flex flex-col items-center justify-center py-12">
            <FolderKanban className="h-12 w-12 text-[#30363d] mb-4" />
            <p className="text-[#8b949e] text-center">
              {searchQuery ? 'No projects found matching your search' : 'No projects yet'}
            </p>
            {!searchQuery && (
              <Button onClick={handleOpenModal} className="mt-4 gap-2">
                <Plus className="h-4 w-4" />
                Create Your First Project
              </Button>
            )}
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {filteredProjects.map(project => {
            const stats = projectStatsMap.get(project.id) || { total: 0, done: 0, progress: 0 };
            
            return (
              <Card
                key={project.id}
                className="border-[#30363d] bg-[#161b22] cursor-pointer hover:border-[#8b949e] transition-colors"
                onClick={() => handleSelectProject(project)}
              >
                <CardContent className="p-6">
                  <div className="space-y-4">
                    {/* Header */}
                    <div className="flex items-start justify-between">
                      <div className="flex items-center gap-3">
                        <div 
                          className="flex h-10 w-10 items-center justify-center rounded-lg"
                          style={{ backgroundColor: project.color || '#6366f1' }}
                        >
                          <FolderKanban className="h-5 w-5 text-white" />
                        </div>
                        <div>
                          <h3 className="font-semibold text-[#f0f6fc]">{project.name}</h3>
                          <p className="text-xs text-[#8b949e]">{stats.total} tasks</p>
                        </div>
                      </div>
                      <div className={`h-2 w-2 rounded-full ${getProjectStatusColor(project.status)}`} />
                    </div>

                    {/* Description */}
                    <p className="text-sm text-[#8b949e] line-clamp-2">
                      {project.description}
                    </p>

                    {/* Progress */}
                    {stats.total > 0 && (
                      <div className="space-y-1">
                        <div className="flex justify-between text-xs text-[#8b949e]">
                          <span>Progress</span>
                          <span>{stats.progress}%</span>
                        </div>
                        <Progress value={stats.progress} className="h-1" />
                      </div>
                    )}

                    {/* Footer */}
                    <div className="flex items-center justify-between">
                      <Badge variant={getProjectBadgeVariant(project.status)} className="capitalize">
                        {project.status.replace('-', ' ')}
                      </Badge>
                      <ChevronRight className="h-4 w-4 text-[#8b949e]" />
                    </div>
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}

      {/* Create Modal */}
      <ProjectModal
        isOpen={isModalOpen}
        onClose={handleCloseModal}
        project={selectedProject || undefined}
      />

      {/* Detail Sheet */}
      <ProjectDetailSheet
        project={selectedProject}
        isOpen={!!selectedProject}
        onClose={handleCloseSheet}
        onEdit={() => {
          setIsModalOpen(true);
        }}
      />
    </div>
  );
}
