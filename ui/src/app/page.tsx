'use client';

import { useEffect, useMemo } from 'react';
import { useAgentsStore } from '@/stores/agents';
import { useTasksStore } from '@/stores/tasks';
import { useEventsStore } from '@/stores/events';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { 
  Bot, 
  ListTodo, 
  CheckCircle2, 
  XCircle,
  Activity,
  Clock,
  TrendingUp
} from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';
import { calculateProgress } from '@/lib/task-utils';
import { getAgentStatusColor } from '@/lib/status-utils';

/** Event icon lookup map - created once, never recreated */
const EVENT_ICON_MAP: Record<string, React.ReactNode> = {
  task_completed: <CheckCircle2 className="h-4 w-4 text-green-400" />,
  task_created: <ListTodo className="h-4 w-4 text-blue-400" />,
  task_started: <Activity className="h-4 w-4 text-purple-400" />,
  agent_spawned: <Bot className="h-4 w-4 text-cyan-400" />,
  phase_completed: <CheckCircle2 className="h-4 w-4 text-emerald-400" />,
  verification_passed: <CheckCircle2 className="h-4 w-4 text-green-400" />,
  verification_failed: <XCircle className="h-4 w-4 text-red-400" />,
};
const DEFAULT_EVENT_ICON = <Activity className="h-4 w-4 text-slate-400" />;

function getEventIcon(type: string): React.ReactNode {
  return EVENT_ICON_MAP[type] || DEFAULT_EVENT_ICON;
}

export default function DashboardPage() {
  const agents = useAgentsStore((s) => s.agents);
  const fetchAgents = useAgentsStore((s) => s.fetchAgents);
  const tasks = useTasksStore((s) => s.tasks);
  const fetchTasks = useTasksStore((s) => s.fetchTasks);
  const events = useEventsStore((s) => s.events);
  const fetchEvents = useEventsStore((s) => s.fetchEvents);

  useEffect(() => {
    fetchAgents();
    fetchTasks();
    fetchEvents({ limit: 10 });
  }, [fetchAgents, fetchTasks, fetchEvents]);

  // Memoize all derived stats in a single pass to avoid repeated iterations
  const stats = useMemo(() => {
    const today = new Date();
    today.setHours(0, 0, 0, 0);

    let activeAgents = 0;
    let queuedTasks = 0;
    let completedToday = 0;
    let failedTasks = 0;

    // Single pass over agents
    for (const a of agents) {
      if (a.status === 'working') activeAgents++;
    }

    // Single pass over tasks instead of 4 separate .filter() calls
    for (const t of tasks) {
      if (t.status === 'queued' || t.status === 'backlog' || t.status === 'planning') {
        queuedTasks++;
      }
      if (t.status === 'done' && new Date(t.updated_at) >= today) {
        completedToday++;
      }
      if (
        t.status === 'failed' ||
        t.phases?.some(p => p.status === 'failed') ||
        t.stories?.some(s => !s.passes)
      ) {
        failedTasks++;
      }
    }

    return { activeAgents, queuedTasks, completedToday, failedTasks };
  }, [agents, tasks]);

  // Memoize working agents list
  const workingAgents = useMemo(
    () => agents.filter(a => a.status === 'working'),
    [agents]
  );

  // Memoize recent events slice
  const recentEvents = useMemo(
    () => events.slice(0, 10),
    [events]
  );

  return (
    <div className="space-y-4 md:space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-2xl md:text-3xl font-bold text-white">Mission Control</h1>
        <p className="mt-1 md:mt-2 text-sm md:text-base text-[#8b949e]">
          Real-time overview of your AI agent operations
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-2 gap-3 md:gap-6 lg:grid-cols-4">
        <Card className="border-[#30363d] bg-[#161b22]">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-[#c9d1d9]">
              Active Agents
            </CardTitle>
            <Bot className="h-4 w-4 text-[#8b949e]" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-white">{stats.activeAgents}</div>
            <p className="text-xs text-[#8b949e]">
              {agents.length} total agents
            </p>
          </CardContent>
        </Card>

        <Card className="border-[#30363d] bg-[#161b22]">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-[#c9d1d9]">
              Tasks in Queue
            </CardTitle>
            <ListTodo className="h-4 w-4 text-[#8b949e]" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-white">{stats.queuedTasks}</div>
            <p className="text-xs text-[#8b949e]">
              Waiting for assignment
            </p>
          </CardContent>
        </Card>

        <Card className="border-[#30363d] bg-[#161b22]">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-[#c9d1d9]">
              Completed Today
            </CardTitle>
            <CheckCircle2 className="h-4 w-4 text-[#8b949e]" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-white">{stats.completedToday}</div>
            <p className="text-xs text-green-400 flex items-center gap-1">
              <TrendingUp className="h-3 w-3" />
              On track
            </p>
          </CardContent>
        </Card>

        <Card className="border-[#30363d] bg-[#161b22]">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-[#c9d1d9]">
              Failed Tasks
            </CardTitle>
            <XCircle className="h-4 w-4 text-[#8b949e]" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-white">{stats.failedTasks}</div>
            <p className="text-xs text-[#8b949e]">
              Require attention
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Two Column Layout */}
      <div className="grid gap-6 lg:grid-cols-2">
        {/* Working Agents */}
        <Card className="border-[#30363d] bg-[#161b22]">
          <CardHeader>
            <CardTitle className="text-white">Currently Working</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {workingAgents.length === 0 ? (
              <p className="text-sm text-[#8b949e]">No agents currently working</p>
            ) : (
              workingAgents.map(agent => {
                const currentTask = tasks.find(t => t.id === agent.current_task_id);
                const progress = currentTask ? calculateProgress(currentTask) : 0;
                
                return (
                  <div key={agent.id} className="space-y-2 rounded-lg border border-[#30363d] p-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <div className={`h-2 w-2 rounded-full ${getAgentStatusColor(agent.status)}`} />
                        <div>
                          <div className="font-medium text-white">{agent.name}</div>
                          <div className="text-sm text-[#8b949e]">{agent.model}</div>
                        </div>
                      </div>
                      <Badge variant="outline" className="border-blue-600 text-blue-400">
                        {agent.status}
                      </Badge>
                    </div>
                    {currentTask && (
                      <div className="space-y-2">
                        <div className="text-sm text-[#c9d1d9]">{currentTask.title}</div>
                        <div className="space-y-1">
                          <div className="flex items-center justify-between text-xs">
                            <span className="text-[#8b949e]">Progress</span>
                            <span className="text-[#c9d1d9]">{progress}%</span>
                          </div>
                          <Progress value={progress} className="h-1.5" />
                        </div>
                      </div>
                    )}
                  </div>
                );
              })
            )}
          </CardContent>
        </Card>

        {/* Recent Activity */}
        <Card className="border-[#30363d] bg-[#161b22]">
          <CardHeader>
            <CardTitle className="text-white">Recent Activity</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {recentEvents.length === 0 ? (
                <p className="text-sm text-[#8b949e]">No recent activity</p>
              ) : (
                recentEvents.map(event => (
                  <div key={event.id} className="flex gap-3">
                    <div className="mt-0.5">{getEventIcon(event.type)}</div>
                    <div className="flex-1 space-y-1">
                      <div className="text-sm text-[#c9d1d9]">{event.message}</div>
                      <div className="flex items-center gap-2 text-xs text-[#6e7681]">
                        <Clock className="h-3 w-3" />
                        {formatDistanceToNow(new Date(event.created_at), { addSuffix: true })}
                      </div>
                    </div>
                  </div>
                ))
              )}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
