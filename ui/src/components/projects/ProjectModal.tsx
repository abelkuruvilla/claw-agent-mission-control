'use client';

import { useState, useEffect } from 'react';
import { useProjectsStore } from '@/stores/projects';
import { useSettingsStore } from '@/stores/settings';
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
import type { Project } from '@/types';
import { toast } from 'sonner';

interface ProjectModalProps {
  isOpen: boolean;
  onClose: () => void;
  project?: Project;
}

const PROJECT_COLORS = [
  { name: 'Blue', value: '#3b82f6' },
  { name: 'Purple', value: '#8b5cf6' },
  { name: 'Green', value: '#10b981' },
  { name: 'Yellow', value: '#f59e0b' },
  { name: 'Red', value: '#ef4444' },
  { name: 'Pink', value: '#ec4899' },
  { name: 'Cyan', value: '#06b6d4' },
  { name: 'Indigo', value: '#6366f1' },
];

export function ProjectModal({ isOpen, onClose, project }: ProjectModalProps) {
  const { createProject, updateProject } = useProjectsStore();
  const { settings, fetchSettings } = useSettingsStore();
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    status: 'active' as 'active' | 'on-hold' | 'completed' | 'archived',
    color: '#3b82f6',
    location: '',
    default_branch: 'main',
    local_exec_branch: '',
    remote_merge_branch: '',
  });
  const [submitting, setSubmitting] = useState(false);

  // Fetch settings on mount to get default project directory
  useEffect(() => {
    fetchSettings();
  }, [fetchSettings]);

  useEffect(() => {
    if (project) {
      setFormData({
        name: project.name,
        description: project.description,
        status: project.status,
        color: project.color || '#3b82f6',
        location: project.location || '',
        default_branch: (project as any).default_branch || 'main',
        local_exec_branch: (project as any).local_exec_branch || '',
        remote_merge_branch: (project as any).remote_merge_branch || '',
      });
    } else {
      // For new projects, use default project directory from settings
      const defaultDir = settings?.default_project_directory || '';
      setFormData({
        name: '',
        description: '',
        status: 'active',
        color: '#3b82f6',
        location: defaultDir,
        default_branch: 'main',
        local_exec_branch: '',
        remote_merge_branch: '',
      });
    }
  }, [project, isOpen, settings]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.name.trim()) {
      toast.error('Project name is required');
      return;
    }

    if (!formData.description.trim()) {
      toast.error('Project description is required');
      return;
    }

    setSubmitting(true);
    
    try {
      if (project) {
        await updateProject(project.id, formData);
        toast.success('Project updated successfully');
      } else {
        await createProject(formData);
        toast.success('Project created successfully');
      }
      
      onClose();
    } catch (error) {
      toast.error(project ? 'Failed to update project' : 'Failed to create project');
      console.error('Project save error:', error);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-2xl bg-[#161b22] border-[#30363d]">
        <DialogHeader>
          <DialogTitle className="text-[#f0f6fc]">
            {project ? 'Edit Project' : 'Create New Project'}
          </DialogTitle>
          <DialogDescription className="text-[#8b949e]">
            Organize your tasks into projects for better tracking
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name" className="text-[#f0f6fc]">Project Name</Label>
            <Input
              id="name"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="e.g., User Authentication System"
              required
              className="bg-[#0d1117] border-[#30363d] text-[#f0f6fc] placeholder:text-[#6e7681]"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="description" className="text-[#f0f6fc]">Description</Label>
            <Textarea
              id="description"
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              placeholder="Describe the project goals and scope"
              rows={3}
              required
              className="bg-[#0d1117] border-[#30363d] text-[#f0f6fc] placeholder:text-[#6e7681]"
            />
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="status" className="text-[#f0f6fc]">Status</Label>
              <Select
                value={formData.status}
                onValueChange={(value) => setFormData({ ...formData, status: value as any })}
              >
                <SelectTrigger className="bg-[#0d1117] border-[#30363d] text-[#f0f6fc]">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="bg-[#0d1117] border-[#30363d]">
                  <SelectItem value="active">Active</SelectItem>
                  <SelectItem value="on-hold">On Hold</SelectItem>
                  <SelectItem value="completed">Completed</SelectItem>
                  <SelectItem value="archived">Archived</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label htmlFor="color" className="text-[#f0f6fc]">Color</Label>
              <Select
                value={formData.color}
                onValueChange={(value) => setFormData({ ...formData, color: value })}
              >
                <SelectTrigger className="bg-[#0d1117] border-[#30363d] text-[#f0f6fc]">
                  <div className="flex items-center gap-2">
                    <div 
                      className="h-4 w-4 rounded"
                      style={{ backgroundColor: formData.color }}
                    />
                    <span>
                      {PROJECT_COLORS.find(c => c.value === formData.color)?.name || 'Custom'}
                    </span>
                  </div>
                </SelectTrigger>
                <SelectContent className="bg-[#0d1117] border-[#30363d]">
                  {PROJECT_COLORS.map(color => (
                    <SelectItem key={color.value} value={color.value}>
                      <div className="flex items-center gap-2">
                        <div 
                          className="h-4 w-4 rounded"
                          style={{ backgroundColor: color.value }}
                        />
                        <span className="text-[#f0f6fc]">{color.name}</span>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="location" className="text-[#f0f6fc]">
              Project Location
              <span className="text-xs text-[#6e7681] ml-2">(optional)</span>
            </Label>
            <Input
              id="location"
              value={formData.location}
              onChange={(e) => setFormData({ ...formData, location: e.target.value })}
              placeholder={settings?.default_project_directory || '~/projects/my-project'}
              className="bg-[#0d1117] border-[#30363d] text-[#f0f6fc] placeholder:text-[#6e7681]"
            />
            <p className="text-xs text-[#6e7681]">
              Directory path for this project. Defaults to settings if empty.
            </p>
          </div>

          {/* Git Branch Settings */}
          <div className="border-t border-[#30363d] pt-4 mt-4">
            <h3 className="text-sm font-medium text-[#f0f6fc] mb-3">Git Branch Settings</h3>
            <div className="grid gap-3">
              <div className="grid gap-1">
                <Label htmlFor="default_branch" className="text-[#f0f6fc] text-xs">
                  Default Branch
                </Label>
                <Input
                  id="default_branch"
                  value={formData.default_branch}
                  onChange={(e) => setFormData({ ...formData, default_branch: e.target.value })}
                  placeholder="main"
                  className="bg-[#0d1117] border-[#30363d] text-[#f0f6fc]"
                />
              </div>
              <div className="grid gap-1">
                <Label htmlFor="local_exec_branch" className="text-[#f0f6fc] text-xs">
                  Local Exec Branch
                </Label>
                <Input
                  id="local_exec_branch"
                  value={formData.local_exec_branch}
                  onChange={(e) => setFormData({ ...formData, local_exec_branch: e.target.value })}
                  placeholder="feature/my-task"
                  className="bg-[#0d1117] border-[#30363d] text-[#f0f6fc]"
                />
              </div>
              <div className="grid gap-1">
                <Label htmlFor="remote_merge_branch" className="text-[#f0f6fc] text-xs">
                  Remote Merge Branch
                </Label>
                <Input
                  id="remote_merge_branch"
                  value={formData.remote_merge_branch}
                  onChange={(e) => setFormData({ ...formData, remote_merge_branch: e.target.value })}
                  placeholder="main"
                  className="bg-[#0d1117] border-[#30363d] text-[#f0f6fc]"
                />
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button 
              type="button" 
              variant="outline" 
              onClick={onClose}
              className="bg-[#0d1117] border-[#30363d] text-[#f0f6fc]"
              disabled={submitting}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={submitting}>
              {submitting ? 'Saving...' : (project ? 'Update Project' : 'Create Project')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
