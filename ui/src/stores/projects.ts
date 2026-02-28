import { create } from 'zustand';
import { api } from '@/services/api';
import type { Project } from '@/types';

interface ProjectsState {
  projects: Project[];
  loading: boolean;
  error: string | null;
  selectedProject: Project | null;
  
  fetchProjects: () => Promise<void>;
  fetchProject: (id: string) => Promise<void>;
  createProject: (project: Partial<Project>) => Promise<Project>;
  updateProject: (id: string, updates: Partial<Project>) => Promise<void>;
  deleteProject: (id: string) => Promise<void>;
  setSelectedProject: (project: Project | null) => void;
}

export const useProjectsStore = create<ProjectsState>((set, get) => ({
  projects: [],
  loading: false,
  error: null,
  selectedProject: null,
  
  fetchProjects: async () => {
    set({ loading: true, error: null });
    try {
      const projects = await api.projects.list();
      set({ projects, loading: false });
    } catch (error) {
      set({ error: (error as Error).message, loading: false });
    }
  },
  
  fetchProject: async (id: string) => {
    set({ loading: true, error: null });
    try {
      const project = await api.projects.get(id);
      set({ selectedProject: project, loading: false });
    } catch (error) {
      set({ error: (error as Error).message, loading: false });
    }
  },
  
  createProject: async (project: Partial<Project>) => {
    const created = await api.projects.create(project);
    set((state) => ({ projects: [...state.projects, created] }));
    return created;
  },
  
  updateProject: async (id: string, updates: Partial<Project>) => {
    const updated = await api.projects.update(id, updates);
    set((state) => ({
      projects: state.projects.map((p) => (p.id === id ? updated : p)),
      selectedProject: state.selectedProject?.id === id ? updated : state.selectedProject,
    }));
  },
  
  deleteProject: async (id: string) => {
    await api.projects.delete(id);
    set((state) => ({
      projects: state.projects.filter((p) => p.id !== id),
      selectedProject: state.selectedProject?.id === id ? null : state.selectedProject,
    }));
  },
  
  setSelectedProject: (project) => set({ selectedProject: project }),
}));
