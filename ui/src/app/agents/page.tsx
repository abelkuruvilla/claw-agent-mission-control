'use client';

import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { useAgentsStore, useTasksStore } from '@/stores';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Plus, Search, Bot, ChevronRight } from 'lucide-react';
import { AgentModal } from '@/components/agents/AgentModal';
import { AgentDetailSheet } from '@/components/agents/AgentDetailSheet';
import { getAgentStatusColor, getAgentBadgeVariant } from '@/lib/status-utils';
import type { Agent } from '@/types';

export default function AgentsPage() {
  const agents = useAgentsStore((s) => s.agents);
  const loading = useAgentsStore((s) => s.loading);
  const fetchAgents = useAgentsStore((s) => s.fetchAgents);
  const tasks = useTasksStore((s) => s.tasks);
  const fetchTasks = useTasksStore((s) => s.fetchTasks);

  const [searchQuery, setSearchQuery] = useState('');
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedAgent, setSelectedAgent] = useState<Agent | null>(null);

  useEffect(() => {
    fetchAgents();
    fetchTasks();
  }, [fetchAgents, fetchTasks]);

  // Memoize filtered agents - only recompute when agents or search changes
  const filteredAgents = useMemo(() => {
    if (!searchQuery) return agents;
    const lower = searchQuery.toLowerCase();
    return agents.filter(agent =>
      agent.name.toLowerCase().includes(lower) ||
      agent.description.toLowerCase().includes(lower)
    );
  }, [agents, searchQuery]);

  // Pre-build a task lookup map by ID for O(1) lookups instead of O(n) .find()
  const taskMap = useMemo(() => {
    const map = new Map<string, typeof tasks[0]>();
    for (const task of tasks) {
      map.set(task.id, task);
    }
    return map;
  }, [tasks]);

  const handleOpenModal = useCallback(() => setIsModalOpen(true), []);
  const handleCloseModal = useCallback(() => setIsModalOpen(false), []);
  const handleSelectAgent = useCallback((agent: Agent) => setSelectedAgent(agent), []);
  const handleCloseSheet = useCallback(() => setSelectedAgent(null), []);

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="text-slate-400">Loading agents...</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-[#f0f6fc]">Agents</h1>
          <p className="mt-2 text-[#8b949e]">
            Manage your AI agents and their configurations
          </p>
        </div>
        <Button onClick={handleOpenModal} className="gap-2">
          <Plus className="h-4 w-4" />
          New Agent
        </Button>
      </div>

      {/* Search */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[#8b949e]" />
        <Input
          placeholder="Search agents..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="pl-10 bg-[#0d1117] border-[#30363d] text-[#f0f6fc]"
        />
      </div>

      {/* Agents Grid */}
      {filteredAgents.length === 0 ? (
        <Card className="border-[#30363d] bg-[#161b22]">
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Bot className="h-12 w-12 text-[#30363d] mb-4" />
            <p className="text-[#8b949e] text-center">
              {searchQuery ? 'No agents found matching your search' : 'No agents yet'}
            </p>
            {!searchQuery && (
              <Button onClick={handleOpenModal} className="mt-4 gap-2">
                <Plus className="h-4 w-4" />
                Create Your First Agent
              </Button>
            )}
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {filteredAgents.map(agent => {
            // O(1) lookup via map instead of O(n) .find()
            const currentTask = agent.current_task_id 
              ? taskMap.get(agent.current_task_id) || null
              : null;
            
            return (
              <Card
                key={agent.id}
                className="border-[#30363d] bg-[#161b22] cursor-pointer hover:border-[#8b949e] transition-colors"
                onClick={() => handleSelectAgent(agent)}
              >
                <CardContent className="p-6">
                  <div className="space-y-4">
                    {/* Header */}
                    <div className="flex items-start justify-between">
                      <div className="flex items-center gap-3">
                        <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-purple-600">
                          <Bot className="h-5 w-5 text-white" />
                        </div>
                        <div>
                          <h3 className="font-semibold text-[#f0f6fc]">{agent.name}</h3>
                          <p className="text-xs text-[#8b949e]">{agent.model}</p>
                        </div>
                      </div>
                      <div className={`h-2 w-2 rounded-full ${getAgentStatusColor(agent.status)}`} />
                    </div>

                    {/* Description */}
                    <p className="text-sm text-[#8b949e] line-clamp-2">
                      {agent.description}
                    </p>

                    {/* Status Badge */}
                    <div className="flex items-center justify-between">
                      <Badge variant={getAgentBadgeVariant(agent.status)} className="capitalize">
                        {agent.status}
                      </Badge>
                      <ChevronRight className="h-4 w-4 text-[#8b949e]" />
                    </div>

                    {/* Current Task */}
                    {currentTask && (
                      <div className="rounded-lg border border-[#30363d] bg-[#0d1117] p-3">
                        <div className="text-xs font-medium text-[#8b949e] mb-1">
                          Current Task
                        </div>
                        <div className="text-sm text-[#f0f6fc] line-clamp-1">
                          {currentTask.title}
                        </div>
                      </div>
                    )}
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}

      {/* Create/Edit Modal */}
      <AgentModal
        isOpen={isModalOpen}
        onClose={handleCloseModal}
        agent={selectedAgent || undefined}
      />

      {/* Agent Detail Sheet */}
      <AgentDetailSheet
        agent={selectedAgent}
        isOpen={!!selectedAgent}
        onClose={handleCloseSheet}
        onEdit={() => {
          handleOpenModal();
        }}
      />
    </div>
  );
}
