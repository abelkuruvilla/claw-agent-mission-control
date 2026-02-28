'use client';

import { useEffect, useState } from 'react';
import { useSettingsStore } from '@/stores/settings';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { CheckCircle2, AlertCircle, Loader2, Info } from 'lucide-react';
import type { Settings } from '@/types';

export default function SettingsPage() {
  const { settings, fetchSettings, updateSettings, testConnection, loading } = useSettingsStore();
  const [formData, setFormData] = useState<Partial<Settings>>({});
  const [testingConnection, setTestingConnection] = useState(false);
  const [connectionStatus, setConnectionStatus] = useState<'idle' | 'success' | 'error'>('idle');
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    fetchSettings();
  }, [fetchSettings]);

  useEffect(() => {
    if (settings) {
      setFormData(settings);
    }
  }, [settings]);

  const handleSave = async () => {
    if (!formData) return;
    
    setSaving(true);
    try {
      await updateSettings(formData);
      setConnectionStatus('idle');
    } catch (error) {
      console.error('Failed to save settings:', error);
    } finally {
      setSaving(false);
    }
  };

  const handleTestConnection = async () => {
    setTestingConnection(true);
    setConnectionStatus('idle');
    
    try {
      const result = await testConnection();
      setConnectionStatus(result ? 'success' : 'error');
    } catch (error) {
      setConnectionStatus('error');
    } finally {
      setTestingConnection(false);
    }
  };

  if (loading && !settings) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 text-[#8b949e] animate-spin" />
      </div>
    );
  }

  return (
    <div className="space-y-6 max-w-4xl">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-white">Settings</h1>
        <p className="mt-2 text-[#8b949e]">
          Configure system settings and execution defaults
        </p>
      </div>

      {/* Connection Settings */}
      <Card className="border-[#30363d] bg-[#161b22]">
        <CardHeader>
          <CardTitle className="text-white">OpenClaw Gateway</CardTitle>
          <CardDescription className="text-[#8b949e]">
            Configure connection to your OpenClaw backend
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="gateway-url" className="text-[#c9d1d9]">Gateway URL</Label>
            <Input
              id="gateway-url"
              value={formData.openclaw_gateway_url || ''}
              onChange={(e) => setFormData({
                ...formData,
                openclaw_gateway_url: e.target.value
              })}
              placeholder="http://localhost:3000"
              className="bg-[#0d1117] border-[#30363d] text-white placeholder:text-[#6e7681]"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="auth-token" className="text-[#c9d1d9]">
              Authentication Token
              <span className="text-xs text-[#6e7681] ml-2">(optional)</span>
            </Label>
            <Input
              id="auth-token"
              type="password"
              value={formData.openclaw_gateway_token || ''}
              onChange={(e) => setFormData({
                ...formData,
                openclaw_gateway_token: e.target.value
              })}
              placeholder="Enter your auth token"
              className="bg-[#0d1117] border-[#30363d] text-white placeholder:text-[#6e7681]"
            />
          </div>
          <div className="flex items-center gap-3">
            <Button
              onClick={handleTestConnection}
              disabled={testingConnection}
              variant="outline"
              className="gap-2 bg-[#161b22] border-[#30363d] text-white hover:bg-[#1f2937]"
            >
              {testingConnection ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Testing...
                </>
              ) : (
                'Test Connection'
              )}
            </Button>
            {connectionStatus === 'success' && (
              <div className="flex items-center gap-2 text-sm text-green-400">
                <CheckCircle2 className="h-4 w-4" />
                Connected
              </div>
            )}
            {connectionStatus === 'error' && (
              <div className="flex items-center gap-2 text-sm text-red-400">
                <AlertCircle className="h-4 w-4" />
                Connection failed
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Project Defaults */}
      <Card className="border-[#30363d] bg-[#161b22]">
        <CardHeader>
          <CardTitle className="text-white">Project Defaults</CardTitle>
          <CardDescription className="text-[#8b949e]">
            Default settings for new projects
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="default-project-directory" className="text-[#c9d1d9]">
              Default Project Directory
            </Label>
            <Input
              id="default-project-directory"
              value={formData.default_project_directory || ''}
              onChange={(e) => setFormData({
                ...formData,
                default_project_directory: e.target.value
              })}
              placeholder="~/projects"
              className="bg-[#0d1117] border-[#30363d] text-white placeholder:text-[#6e7681]"
            />
            <p className="text-xs text-[#6e7681]">
              Base directory where new projects will be created. Can be overridden per project.
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Execution Defaults */}
      <Card className="border-[#30363d] bg-[#161b22]">
        <CardHeader>
          <CardTitle className="text-white">Execution Defaults</CardTitle>
          <CardDescription className="text-[#8b949e]">
            Default settings for task execution
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Protocol explanation */}
          <div className="p-3 bg-purple-500/10 border border-purple-500/30 rounded-lg flex gap-3">
            <Info className="h-5 w-5 text-purple-400 flex-shrink-0 mt-0.5" />
            <div className="text-sm text-purple-200">
              <p className="font-medium mb-1">Task Execution Protocols</p>
              <p className="text-purple-300/80">
                All tasks use <span className="text-purple-400 font-medium">GSD</span> for planning 
                (research, requirements, roadmap) and <span className="text-purple-400 font-medium">Ralph Loop</span> for 
                execution (iterate on stories until complete). Configure each protocol below.
              </p>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="default-model" className="text-[#c9d1d9]">Default Model</Label>
            <Select
              value={formData.default_model}
              onValueChange={(value) => setFormData({
                ...formData,
                default_model: value
              })}
            >
              <SelectTrigger className="bg-[#0d1117] border-[#30363d] text-white">
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="bg-[#0d1117] border-[#30363d]">
                <SelectItem value="anthropic/claude-sonnet-4-5">Claude Sonnet 4.5</SelectItem>
                <SelectItem value="anthropic/claude-opus-4-5">Claude Opus 4.5</SelectItem>
                <SelectItem value="anthropic/claude-haiku-4-5">Claude Haiku 4.5</SelectItem>
                <SelectItem value="openai/gpt-4-turbo">GPT-4 Turbo</SelectItem>
                <SelectItem value="openai/gpt-4">GPT-4</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label htmlFor="max-parallel" className="text-[#c9d1d9]">Max Parallel Executions</Label>
            <Input
              id="max-parallel"
              type="number"
              min="1"
              max="10"
              value={formData.max_parallel_executions || 1}
              onChange={(e) => setFormData({
                ...formData,
                max_parallel_executions: parseInt(e.target.value) || 1
              })}
              className="bg-[#0d1117] border-[#30363d] text-white"
            />
          </div>
        </CardContent>
      </Card>

      {/* GSD Settings */}
      <Card className="border-[#30363d] bg-[#161b22]">
        <CardHeader>
          <CardTitle className="text-white">GSD Settings (Planning Phase)</CardTitle>
          <CardDescription className="text-[#8b949e]">
            Configuration for the GSD planning protocol — used when orchestrating new tasks
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="gsd-depth" className="text-[#c9d1d9]">Planning Depth</Label>
            <Select
              value={formData.gsd_depth}
              onValueChange={(value) => setFormData({
                ...formData,
                gsd_depth: value
              })}
            >
              <SelectTrigger className="bg-[#0d1117] border-[#30363d] text-white">
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="bg-[#0d1117] border-[#30363d]">
                <SelectItem value="quick">Quick — Minimal research, fast planning</SelectItem>
                <SelectItem value="standard">Standard — Balanced research and planning</SelectItem>
                <SelectItem value="comprehensive">Comprehensive — Deep research, thorough planning</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label htmlFor="gsd-mode" className="text-[#c9d1d9]">Execution Mode</Label>
            <Select
              value={formData.gsd_mode}
              onValueChange={(value) => setFormData({
                ...formData,
                gsd_mode: value
              })}
            >
              <SelectTrigger className="bg-[#0d1117] border-[#30363d] text-white">
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="bg-[#0d1117] border-[#30363d]">
                <SelectItem value="interactive">Interactive — Confirm at each step</SelectItem>
                <SelectItem value="yolo">YOLO — Auto-approve all steps</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <Separator className="bg-[#30363d]" />
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label htmlFor="research-enabled" className="text-[#c9d1d9]">Enable Research</Label>
              <p className="text-sm text-[#8b949e]">
                Research the domain before planning
              </p>
            </div>
            <Switch
              id="research-enabled"
              checked={formData.gsd_research_enabled || false}
              onCheckedChange={(checked) => setFormData({
                ...formData,
                gsd_research_enabled: checked
              })}
            />
          </div>
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label htmlFor="plan-check-enabled" className="text-[#c9d1d9]">Enable Plan Check</Label>
              <p className="text-sm text-[#8b949e]">
                Verify plans before execution
              </p>
            </div>
            <Switch
              id="plan-check-enabled"
              checked={formData.gsd_plan_check_enabled || false}
              onCheckedChange={(checked) => setFormData({
                ...formData,
                gsd_plan_check_enabled: checked
              })}
            />
          </div>
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label htmlFor="verifier-enabled" className="text-[#c9d1d9]">Enable Verifier</Label>
              <p className="text-sm text-[#8b949e]">
                Verify deliverables after execution
              </p>
            </div>
            <Switch
              id="verifier-enabled"
              checked={formData.gsd_verifier_enabled || false}
              onCheckedChange={(checked) => setFormData({
                ...formData,
                gsd_verifier_enabled: checked
              })}
            />
          </div>
        </CardContent>
      </Card>

      {/* Ralph Settings */}
      <Card className="border-[#30363d] bg-[#161b22]">
        <CardHeader>
          <CardTitle className="text-white">Ralph Settings (Execution Phase)</CardTitle>
          <CardDescription className="text-[#8b949e]">
            Configuration for the Ralph Loop execution protocol — used when agents work on stories
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="max-iterations" className="text-[#c9d1d9]">Max Iterations per Story</Label>
            <Input
              id="max-iterations"
              type="number"
              min="1"
              max="50"
              value={formData.ralph_max_iterations || 10}
              onChange={(e) => setFormData({
                ...formData,
                ralph_max_iterations: parseInt(e.target.value) || 10
              })}
              className="bg-[#0d1117] border-[#30363d] text-white"
            />
            <p className="text-xs text-[#6e7681]">
              Maximum number of retry attempts for each story before marking as failed
            </p>
          </div>
          <Separator className="bg-[#30363d]" />
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label htmlFor="auto-commit" className="text-[#c9d1d9]">Auto-commit on Pass</Label>
              <p className="text-sm text-[#8b949e]">
                Automatically commit changes when a story passes
              </p>
            </div>
            <Switch
              id="auto-commit"
              checked={formData.ralph_auto_commit || false}
              onCheckedChange={(checked) => setFormData({
                ...formData,
                ralph_auto_commit: checked
              })}
            />
          </div>
        </CardContent>
      </Card>

      {/* Save Button */}
      <div className="flex justify-end gap-3">
        <Button 
          onClick={() => setFormData(settings || {})} 
          variant="outline"
          className="bg-[#161b22] border-[#30363d] text-white hover:bg-[#1f2937]"
        >
          Reset
        </Button>
        <Button 
          onClick={handleSave} 
          disabled={saving}
          className="gap-2"
        >
          {saving ? (
            <>
              <Loader2 className="h-4 w-4 animate-spin" />
              Saving...
            </>
          ) : (
            <>
              <CheckCircle2 className="h-4 w-4" />
              Save Settings
            </>
          )}
        </Button>
      </div>
    </div>
  );
}
