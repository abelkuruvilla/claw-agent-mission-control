'use client';

import React, { useState, useEffect } from 'react';
import { useAgentsStore } from '@/stores';
import { modelsApi, type ModelConfig } from '@/services/api';
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ChevronDown, ChevronRight } from 'lucide-react';
import type { Agent } from '@/types';

interface AgentModalProps {
  isOpen: boolean;
  onClose: () => void;
  agent?: Agent;
}

export const AgentModal: React.FC<AgentModalProps> = ({ isOpen, onClose, agent }) => {
  const createAgent = useAgentsStore((s) => s.createAgent);
  const updateAgent = useAgentsStore((s) => s.updateAgent);
  const [models, setModels] = useState<ModelConfig[]>([]);
  const [loadingModels, setLoadingModels] = useState(false);
  const [expandedSections, setExpandedSections] = useState<{[key: string]: boolean}>({
    soul: false,
    identity: false,
    agents: false,
    tools: false,
  });
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    model: '',
    mention_patterns: '',
    workspace_path: '',
    soul_md: '',
    identity_md: '',
    agents_md: '',
    tools_md: '',
  });

  // Fetch models from OpenClaw config
  useEffect(() => {
    const fetchModels = async () => {
      setLoadingModels(true);
      try {
        const configuredModels = await modelsApi.list();
        setModels(configuredModels);
        // Set default model if none selected
        if (configuredModels.length > 0 && !formData.model) {
          setFormData(prev => ({ ...prev, model: configuredModels[0].id }));
        }
      } catch (error) {
        console.error('Failed to fetch models:', error);
      } finally {
        setLoadingModels(false);
      }
    };
    
    if (isOpen) {
      fetchModels();
    }
  }, [isOpen]);

  useEffect(() => {
    const initialData = agent ? {
      name: agent.name,
      description: agent.description,
      model: agent.model,
      mention_patterns: Array.isArray(agent.mention_patterns) 
        ? agent.mention_patterns.join(', ') 
        : typeof agent.mention_patterns === 'string' 
          ? agent.mention_patterns 
          : '',
      workspace_path: agent.workspace_path || '',
      soul_md: agent.soul_md || '',
      identity_md: agent.identity_md || '',
      agents_md: agent.agents_md || '',
      tools_md: agent.tools_md || '',
    } : {
      name: '',
      description: '',
      model: models.length > 0 ? models[0].id : '',
      mention_patterns: '',
      workspace_path: '',
      soul_md: '',
      identity_md: '',
      agents_md: '',
      tools_md: '',
    };
    setFormData(initialData);
  }, [agent, isOpen, models]);

  const toggleSection = (section: string) => {
    setExpandedSections(prev => ({
      ...prev,
      [section]: !prev[section],
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    const agentData = {
      name: formData.name,
      description: formData.description,
      status: agent?.status || ('idle' as const),
      model: formData.model,
      mention_patterns: formData.mention_patterns
        .split(',')
        .map(p => p.trim())
        .filter(Boolean),
      workspace_path: formData.workspace_path || `/workspace/${formData.name.toLowerCase().replace(/\s+/g, '-')}`,
      soul_md: formData.soul_md,
      identity_md: formData.identity_md,
      agents_md: formData.agents_md,
      tools_md: formData.tools_md,
    };

    if (agent) {
      await updateAgent(agent.id, agentData);
    } else {
      await createAgent(agentData);
    }
    
    onClose();
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-3xl max-h-[90vh] overflow-y-auto bg-[#161b22] border-[#30363d]">
        <DialogHeader>
          <DialogTitle className="text-[#f0f6fc]">
            {agent ? 'Edit Agent' : 'Create New Agent'}
          </DialogTitle>
          <DialogDescription className="text-[#8b949e]">
            Configure your AI agent&apos;s settings and behavior
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-6">
          <Tabs defaultValue="basic" className="w-full">
            <TabsList className="grid w-full grid-cols-2 bg-[#0d1117]">
              <TabsTrigger value="basic" className="data-[state=active]:bg-[#161b22]">
                Basic Info
              </TabsTrigger>
              <TabsTrigger value="config" className="data-[state=active]:bg-[#161b22]">
                Configuration
              </TabsTrigger>
            </TabsList>

            <TabsContent value="basic" className="space-y-4 mt-4">
              <div className="space-y-2">
                <Label htmlFor="name" className="text-[#f0f6fc]">Agent Name</Label>
                <Input
                  id="name"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="e.g., CodeMaster"
                  required
                  className="bg-[#0d1117] border-[#30363d] text-[#f0f6fc]"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="description" className="text-[#f0f6fc]">Description</Label>
                <Textarea
                  id="description"
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  placeholder="Describe the agent's purpose and capabilities"
                  rows={3}
                  required
                  className="bg-[#0d1117] border-[#30363d] text-[#f0f6fc]"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="model" className="text-[#f0f6fc]">Model</Label>
                <Select
                  value={formData.model}
                  onValueChange={(value) => setFormData({ ...formData, model: value })}
                  disabled={loadingModels}
                >
                  <SelectTrigger className="bg-[#0d1117] border-[#30363d] text-[#f0f6fc]">
                    <SelectValue placeholder={loadingModels ? "Loading models..." : "Select a model"} />
                  </SelectTrigger>
                  <SelectContent className="bg-[#0d1117] border-[#30363d]">
                    {models.length === 0 && !loadingModels ? (
                      <SelectItem value="no-models" disabled>No models configured</SelectItem>
                    ) : (
                      models.map((model) => (
                        <SelectItem key={model.id} value={model.id}>
                          {model.alias ? `${model.alias} (${model.id})` : model.id}
                        </SelectItem>
                      ))
                    )}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="mentions" className="text-[#f0f6fc]">
                  Mention Patterns
                  <span className="text-xs text-[#8b949e] ml-2">(comma-separated)</span>
                </Label>
                <Input
                  id="mentions"
                  value={formData.mention_patterns}
                  onChange={(e) => setFormData({ ...formData, mention_patterns: e.target.value })}
                  placeholder="e.g., @agent, @myagent"
                  className="bg-[#0d1117] border-[#30363d] text-[#f0f6fc]"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="workspace" className="text-[#f0f6fc]">Workspace Path</Label>
                <Input
                  id="workspace"
                  value={formData.workspace_path}
                  onChange={(e) => setFormData({ ...formData, workspace_path: e.target.value })}
                  placeholder="/workspace/agent-name"
                  className="bg-[#0d1117] border-[#30363d] text-[#f0f6fc]"
                />
              </div>
            </TabsContent>

            <TabsContent value="config" className="space-y-4 mt-4">
              {/* SOUL.md */}
              <div className="space-y-2">
                <button
                  type="button"
                  onClick={() => toggleSection('soul')}
                  className="flex items-center gap-2 text-[#f0f6fc] hover:text-[#8b949e] transition-colors"
                >
                  {expandedSections.soul ? (
                    <ChevronDown className="h-4 w-4" />
                  ) : (
                    <ChevronRight className="h-4 w-4" />
                  )}
                  <Label className="cursor-pointer">SOUL.md</Label>
                </button>
                {expandedSections.soul && (
                  <Textarea
                    value={formData.soul_md}
                    onChange={(e) => setFormData({ ...formData, soul_md: e.target.value })}
                    placeholder="# Agent Soul&#10;&#10;Define the agent&apos;s core personality and purpose..."
                    rows={6}
                    className="font-mono text-sm bg-[#0d1117] border-[#30363d] text-[#f0f6fc]"
                  />
                )}
              </div>

              {/* IDENTITY.md */}
              <div className="space-y-2">
                <button
                  type="button"
                  onClick={() => toggleSection('identity')}
                  className="flex items-center gap-2 text-[#f0f6fc] hover:text-[#8b949e] transition-colors"
                >
                  {expandedSections.identity ? (
                    <ChevronDown className="h-4 w-4" />
                  ) : (
                    <ChevronRight className="h-4 w-4" />
                  )}
                  <Label className="cursor-pointer">IDENTITY.md</Label>
                </button>
                {expandedSections.identity && (
                  <Textarea
                    value={formData.identity_md}
                    onChange={(e) => setFormData({ ...formData, identity_md: e.target.value })}
                    placeholder="# Identity&#10;&#10;Define skills, expertise, and knowledge areas..."
                    rows={6}
                    className="font-mono text-sm bg-[#0d1117] border-[#30363d] text-[#f0f6fc]"
                  />
                )}
              </div>

              {/* AGENTS.md */}
              <div className="space-y-2">
                <button
                  type="button"
                  onClick={() => toggleSection('agents')}
                  className="flex items-center gap-2 text-[#f0f6fc] hover:text-[#8b949e] transition-colors"
                >
                  {expandedSections.agents ? (
                    <ChevronDown className="h-4 w-4" />
                  ) : (
                    <ChevronRight className="h-4 w-4" />
                  )}
                  <Label className="cursor-pointer">AGENTS.md</Label>
                </button>
                {expandedSections.agents && (
                  <Textarea
                    value={formData.agents_md}
                    onChange={(e) => setFormData({ ...formData, agents_md: e.target.value })}
                    placeholder="# Sub-agents&#10;&#10;List available sub-agents..."
                    rows={5}
                    className="font-mono text-sm bg-[#0d1117] border-[#30363d] text-[#f0f6fc]"
                  />
                )}
              </div>

              {/* TOOLS.md */}
              <div className="space-y-2">
                <button
                  type="button"
                  onClick={() => toggleSection('tools')}
                  className="flex items-center gap-2 text-[#f0f6fc] hover:text-[#8b949e] transition-colors"
                >
                  {expandedSections.tools ? (
                    <ChevronDown className="h-4 w-4" />
                  ) : (
                    <ChevronRight className="h-4 w-4" />
                  )}
                  <Label className="cursor-pointer">TOOLS.md</Label>
                </button>
                {expandedSections.tools && (
                  <Textarea
                    value={formData.tools_md}
                    onChange={(e) => setFormData({ ...formData, tools_md: e.target.value })}
                    placeholder="# Available Tools&#10;&#10;List tools and capabilities..."
                    rows={5}
                    className="font-mono text-sm bg-[#0d1117] border-[#30363d] text-[#f0f6fc]"
                  />
                )}
              </div>
            </TabsContent>
          </Tabs>

          <DialogFooter>
            <Button 
              type="button" 
              variant="outline" 
              onClick={onClose}
              className="border-[#30363d] text-[#f0f6fc] hover:bg-[#0d1117]"
            >
              Cancel
            </Button>
            <Button type="submit">
              {agent ? 'Update Agent' : 'Create Agent'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
};
