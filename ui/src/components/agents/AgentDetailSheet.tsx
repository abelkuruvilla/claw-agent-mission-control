'use client';

import React, { useState } from 'react';
import { useAgentsStore, useTasksStore } from '@/stores';
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Textarea } from '@/components/ui/textarea';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Edit, Trash2, Bot, Save, X } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';
import { getAgentStatusColor } from '@/lib/status-utils';
import type { Agent } from '@/types';
import { AgentChat } from './AgentChat';

interface AgentDetailSheetProps {
  agent: Agent | null;
  isOpen: boolean;
  onClose: () => void;
  onEdit?: () => void;
}

export function AgentDetailSheet({ agent, isOpen, onClose, onEdit }: AgentDetailSheetProps) {
  const updateAgent = useAgentsStore((s) => s.updateAgent);
  const deleteAgent = useAgentsStore((s) => s.deleteAgent);
  const tasks = useTasksStore((s) => s.tasks);
  const [editingConfig, setEditingConfig] = useState<string | null>(null);
  const [configContent, setConfigContent] = useState('');

  if (!agent) return null;

  const agentTasks = tasks.filter(t => t.agent_id === agent.id);
  const currentTask = agent.current_task_id 
    ? tasks.find(t => t.id === agent.current_task_id)
    : null;

  const handleDelete = async () => {
    if (confirm(`Are you sure you want to delete ${agent.name}?`)) {
      await deleteAgent(agent.id);
      onClose();
    }
  };

  const handleEditConfig = (field: string, content: string) => {
    setEditingConfig(field);
    setConfigContent(content || '');
  };

  const handleSaveConfig = async () => {
    if (!editingConfig) return;
    await updateAgent(agent.id, { [editingConfig]: configContent });
    setEditingConfig(null);
    setConfigContent('');
  };

  const configFields = [
    { key: 'soul_md', label: 'SOUL.md' },
    { key: 'identity_md', label: 'IDENTITY.md' },
    { key: 'agents_md', label: 'AGENTS.md' },
    { key: 'user_md', label: 'USER.md' },
    { key: 'tools_md', label: 'TOOLS.md' },
    { key: 'memory_md', label: 'MEMORY.md' },
  ];

  return (
    <Sheet open={isOpen} onOpenChange={onClose}>
      <SheetContent className="w-full sm:w-[600px] sm:max-w-[600px] bg-[#161b22] border-[#30363d] overflow-hidden flex flex-col">
        <SheetHeader>
          <div className="flex items-center gap-3">
            <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-purple-600">
              <Bot className="h-6 w-6 text-white" />
            </div>
            <div className="flex-1">
              <SheetTitle className="text-white flex items-center gap-2">
                {agent.name}
                <div className={`h-2 w-2 rounded-full ${getAgentStatusColor(agent.status)}`} />
              </SheetTitle>
              <SheetDescription className="text-slate-400">
                {agent.description}
              </SheetDescription>
            </div>
          </div>
          <div className="flex gap-2 pt-2">
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
          <Tabs defaultValue="overview" className="w-full">
            <TabsList className="bg-slate-800 w-full">
              <TabsTrigger value="overview" className="flex-1">Overview</TabsTrigger>
              <TabsTrigger value="config" className="flex-1">Config</TabsTrigger>
              <TabsTrigger value="tasks" className="flex-1">Tasks</TabsTrigger>
              <TabsTrigger value="chat" className="flex-1">Chat</TabsTrigger>
            </TabsList>

            <TabsContent value="overview" className="space-y-4 mt-4">
              <div className="grid grid-cols-2 gap-4">
                <Card className="border-slate-800 bg-slate-900">
                  <CardContent className="pt-4">
                    <div className="text-sm text-slate-400">Status</div>
                    <Badge className="mt-1 capitalize">{agent.status}</Badge>
                  </CardContent>
                </Card>
                <Card className="border-slate-800 bg-slate-900">
                  <CardContent className="pt-4">
                    <div className="text-sm text-slate-400">Model</div>
                    <div className="text-white mt-1 text-sm truncate">{agent.model}</div>
                  </CardContent>
                </Card>
              </div>

              {currentTask && (
                <Card className="border-slate-800 bg-slate-900">
                  <CardHeader className="pb-2">
                    <CardTitle className="text-sm text-white">Current Task</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="text-white">{currentTask.title}</div>
                    <Badge className="mt-2" variant="outline">{currentTask.status}</Badge>
                  </CardContent>
                </Card>
              )}

              <Card className="border-slate-800 bg-slate-900">
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm text-white">Details</CardTitle>
                </CardHeader>
                <CardContent className="space-y-2 text-sm">
                  <div>
                    <span className="text-slate-400">Workspace: </span>
                    <span className="text-slate-300 font-mono">{agent.workspace_path || 'N/A'}</span>
                  </div>
                  <div>
                    <span className="text-slate-400">Created: </span>
                    <span className="text-slate-300">
                      {formatDistanceToNow(new Date(agent.created_at), { addSuffix: true })}
                    </span>
                  </div>
                  <div>
                    <span className="text-slate-400">Tasks: </span>
                    <span className="text-slate-300">{agentTasks.length}</span>
                  </div>
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="config" className="space-y-4 mt-4">
              {configFields.map(({ key, label }) => {
                const content = (agent as unknown as Record<string, unknown>)[key] as string | undefined;
                const isEditing = editingConfig === key;

                return (
                  <Card key={key} className="border-slate-800 bg-slate-900">
                    <CardHeader className="pb-2 flex flex-row items-center justify-between">
                      <CardTitle className="text-sm text-white">{label}</CardTitle>
                      {!isEditing && content && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleEditConfig(key, content)}
                        >
                          <Edit className="h-3 w-3" />
                        </Button>
                      )}
                    </CardHeader>
                    <CardContent>
                      {isEditing ? (
                        <div className="space-y-2">
                          <Textarea
                            value={configContent}
                            onChange={(e) => setConfigContent(e.target.value)}
                            className="min-h-[150px] font-mono text-xs bg-slate-950 border-slate-800"
                          />
                          <div className="flex gap-2">
                            <Button size="sm" onClick={handleSaveConfig}>
                              <Save className="h-3 w-3 mr-1" /> Save
                            </Button>
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={() => setEditingConfig(null)}
                            >
                              <X className="h-3 w-3 mr-1" /> Cancel
                            </Button>
                          </div>
                        </div>
                      ) : content ? (
                        <pre className="text-xs text-slate-300 whitespace-pre-wrap font-mono max-h-[200px] overflow-auto">
                          {content.slice(0, 500)}{content.length > 500 ? '...' : ''}
                        </pre>
                      ) : (
                        <p className="text-xs text-slate-500">Not configured</p>
                      )}
                    </CardContent>
                  </Card>
                );
              })}
            </TabsContent>

            <TabsContent value="tasks" className="space-y-4 mt-4">
              {agentTasks.length === 0 ? (
                <p className="text-slate-400 text-center py-8">No tasks assigned</p>
              ) : (
                agentTasks.map(task => (
                  <Card key={task.id} className="border-slate-800 bg-slate-900">
                    <CardContent className="pt-4">
                      <div className="font-medium text-white">{task.title}</div>
                      <div className="flex items-center gap-2 mt-2">
                        <Badge variant="secondary">{task.status}</Badge>
                        {task.stories_total && task.stories_total > 0 && (
                          <Badge variant="outline">
                            {task.stories_passed || 0}/{task.stories_total} stories
                          </Badge>
                        )}
                      </div>
                    </CardContent>
                  </Card>
                ))
              )}
            </TabsContent>

            <TabsContent value="chat" className="mt-4 h-[calc(100vh-280px)]">
              <AgentChat agentId={agent.id} />
            </TabsContent>
          </Tabs>
        </ScrollArea>
      </SheetContent>
    </Sheet>
  );
}
