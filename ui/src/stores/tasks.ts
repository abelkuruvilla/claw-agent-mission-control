import { create } from 'zustand';
import type { Task, TaskStatus } from '@/types';
import { tasksApi } from '@/services/api';

interface TasksState {
  tasks: Task[];
  loading: boolean;
  error: string | null;
  selectedTask: Task | null;
  /** Tracks in-flight fetchTasks to prevent duplicate concurrent requests */
  _fetchPromise: Promise<void> | null;

  fetchTasks: () => Promise<void>;
  fetchTask: (id: string) => Promise<void>;
  createTask: (task: Partial<Task>) => Promise<Task>;
  updateTask: (id: string, updates: Partial<Task>) => Promise<void>;
  updateTaskStatus: (id: string, status: TaskStatus) => Promise<void>;
  deleteTask: (id: string) => Promise<void>;
  setSelectedTask: (task: Task | null) => void;
}

export const useTasksStore = create<TasksState>((set, get) => ({
  tasks: [],
  loading: false,
  error: null,
  selectedTask: null,
  _fetchPromise: null,

  fetchTasks: async () => {
    // Dedup: reuse in-flight request
    const existing = get()._fetchPromise;
    if (existing) {
      return existing;
    }

    set({ loading: true, error: null });

    const promise = (async () => {
      try {
        const tasks = await tasksApi.list();
        set({ tasks, loading: false, _fetchPromise: null });
      } catch (error) {
        set({ error: (error as Error).message, loading: false, _fetchPromise: null });
      }
    })();

    set({ _fetchPromise: promise });
    return promise;
  },

  fetchTask: async (id: string) => {
    set({ loading: true, error: null });
    try {
      const task = await tasksApi.get(id);
      set({ selectedTask: task, loading: false });
    } catch (error) {
      set({ error: (error as Error).message, loading: false });
    }
  },

  createTask: async (task: Partial<Task>) => {
    const created = await tasksApi.create(task);
    set((state) => ({ tasks: [...state.tasks, created] }));
    return created;
  },

  updateTask: async (id: string, updates: Partial<Task>) => {
    const updated = await tasksApi.update(id, updates);
    set((state) => ({
      tasks: state.tasks.map((t) => (t.id === id ? updated : t)),
      selectedTask: state.selectedTask?.id === id ? updated : state.selectedTask,
    }));
  },

  updateTaskStatus: async (id: string, status: TaskStatus) => {
    const updated = await tasksApi.updateStatus(id, status);
    set((state) => ({
      tasks: state.tasks.map((t) => (t.id === id ? updated : t)),
    }));
  },

  deleteTask: async (id: string) => {
    await tasksApi.delete(id);
    set((state) => ({
      tasks: state.tasks.filter((t) => t.id !== id),
      selectedTask: state.selectedTask?.id === id ? null : state.selectedTask,
    }));
  },

  setSelectedTask: (task) => set({ selectedTask: task }),
}));
