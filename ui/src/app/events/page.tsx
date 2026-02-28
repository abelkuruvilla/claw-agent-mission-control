'use client';

import { useEffect, useState, useMemo, useCallback } from 'react';
import { useEventsStore } from '@/stores/events';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  CheckCircle2,
  XCircle,
  ListTodo,
  Activity,
  Bot,
  Clock,
  Search,
  Filter
} from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';
import { getEventBadgeVariant, formatEventType } from '@/lib/status-utils';
import type { EventType } from '@/types';

/** Event icon lookup map - created once, module-level */
const EVENT_ICON_MAP: Record<string, React.ReactNode> = {
  task_completed: <CheckCircle2 className="h-5 w-5 text-green-400" />,
  task_created: <ListTodo className="h-5 w-5 text-blue-400" />,
  task_started: <Activity className="h-5 w-5 text-purple-400" />,
  task_assigned: <ListTodo className="h-5 w-5 text-cyan-400" />,
  agent_spawned: <Bot className="h-5 w-5 text-cyan-400" />,
  phase_completed: <CheckCircle2 className="h-5 w-5 text-emerald-400" />,
  phase_started: <Activity className="h-5 w-5 text-blue-400" />,
  story_passed: <CheckCircle2 className="h-5 w-5 text-green-400" />,
  story_failed: <XCircle className="h-5 w-5 text-red-400" />,
  verification_passed: <CheckCircle2 className="h-5 w-5 text-green-400" />,
  verification_failed: <XCircle className="h-5 w-5 text-red-400" />,
  commit_created: <CheckCircle2 className="h-5 w-5 text-blue-400" />,
  execution_error: <XCircle className="h-5 w-5 text-red-400" />,
};
const DEFAULT_EVENT_ICON = <Activity className="h-5 w-5 text-[#8b949e]" />;

function getEventIcon(type: EventType): React.ReactNode {
  return EVENT_ICON_MAP[type] || DEFAULT_EVENT_ICON;
}

export default function EventsPage() {
  const events = useEventsStore((s) => s.events);
  const fetchEvents = useEventsStore((s) => s.fetchEvents);
  const loading = useEventsStore((s) => s.loading);

  const [searchQuery, setSearchQuery] = useState('');
  const [filterType, setFilterType] = useState<string>('all');
  const [displayLimit, setDisplayLimit] = useState(20);

  useEffect(() => {
    fetchEvents();
  }, [fetchEvents]);

  // Memoize unique event types - only recompute when events array changes
  const uniqueEventTypes = useMemo(
    () => Array.from(new Set(events.map(e => e.type))),
    [events]
  );

  // Memoize the full filtered result (before slicing) for count comparison
  const allFilteredEvents = useMemo(() => {
    const lowerSearch = searchQuery.toLowerCase();
    return events.filter(event => {
      const matchesSearch = !searchQuery ||
        event.message.toLowerCase().includes(lowerSearch) ||
        event.details?.toLowerCase().includes(lowerSearch);
      const matchesFilter = filterType === 'all' || event.type === filterType;
      return matchesSearch && matchesFilter;
    });
  }, [events, searchQuery, filterType]);

  // Slice the filtered events for display
  const filteredEvents = useMemo(
    () => allFilteredEvents.slice(0, displayLimit),
    [allFilteredEvents, displayLimit]
  );

  // Memoize stats - single pass over events array
  const stats = useMemo(() => {
    let completed = 0;
    let created = 0;
    let spawned = 0;
    for (const e of events) {
      if (e.type === 'task_completed') completed++;
      else if (e.type === 'task_created') created++;
      else if (e.type === 'agent_spawned') spawned++;
    }
    return { total: events.length, completed, created, spawned };
  }, [events]);

  const handleLoadMore = useCallback(() => {
    setDisplayLimit(prev => prev + 20);
  }, []);

  const hasMore = filteredEvents.length < allFilteredEvents.length;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-white">Activity Log</h1>
        <p className="mt-2 text-[#8b949e]">
          Complete history of all system events and actions
        </p>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-3 sm:gap-4">
        <div className="relative flex-1 min-w-0">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[#8b949e]" />
          <Input
            placeholder="Search events..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10 bg-[#161b22] border-[#30363d] text-white placeholder:text-[#6e7681] w-full"
          />
        </div>
        <div className="w-full sm:w-64 flex-shrink-0">
          <Select value={filterType} onValueChange={setFilterType}>
            <SelectTrigger className="bg-[#161b22] border-[#30363d] text-white w-full">
              <Filter className="h-4 w-4 mr-2" />
              <SelectValue />
            </SelectTrigger>
            <SelectContent className="bg-[#0d1117] border-[#30363d]">
              <SelectItem value="all">All Events</SelectItem>
              {uniqueEventTypes.map(type => (
                <SelectItem key={type} value={type}>
                  {formatEventType(type)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Events List */}
      {loading && events.length === 0 ? (
        <Card className="border-[#30363d] bg-[#161b22]">
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Activity className="h-12 w-12 text-[#6e7681] mb-4 animate-pulse" />
            <p className="text-[#8b949e] text-center">Loading events...</p>
          </CardContent>
        </Card>
      ) : filteredEvents.length === 0 ? (
        <Card className="border-[#30363d] bg-[#161b22]">
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Activity className="h-12 w-12 text-[#6e7681] mb-4" />
            <p className="text-[#8b949e] text-center">
              {searchQuery || filterType !== 'all' 
                ? 'No events found matching your filters' 
                : 'No events yet'}
            </p>
          </CardContent>
        </Card>
      ) : (
        <>
          <div className="space-y-3">
            {filteredEvents.map((event) => (
              <Card key={event.id} className="border-[#30363d] bg-[#161b22] hover:border-[#484f58] transition-colors overflow-hidden">
                <CardContent className="p-4">
                  <div className="flex gap-3 sm:gap-4">
                    {/* Icon */}
                    <div className="flex-shrink-0 mt-0.5">
                      {getEventIcon(event.type)}
                    </div>

                    {/* Content */}
                    <div className="flex-1 min-w-0">
                      <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-2 sm:gap-4">
                        <div className="flex-1 min-w-0">
                          <p className="text-[#c9d1d9] break-words">{event.message}</p>
                          {event.details && (
                            <p className="text-sm text-[#8b949e] mt-1 break-words">{event.details}</p>
                          )}
                          <div className="flex flex-wrap items-center gap-2 sm:gap-3 mt-2">
                            {event.agent_id && (
                              <div className="flex items-center gap-1.5 text-xs sm:text-sm text-[#8b949e]">
                                <Bot className="h-3.5 w-3.5" />
                                <span className="truncate max-w-[100px] sm:max-w-none">Agent {event.agent_id.slice(0, 8)}</span>
                              </div>
                            )}
                            {event.task_id && (
                              <div className="flex items-center gap-1.5 text-xs sm:text-sm text-[#8b949e]">
                                <ListTodo className="h-3.5 w-3.5" />
                                <span className="truncate max-w-[100px] sm:max-w-none">Task {event.task_id.slice(0, 8)}</span>
                              </div>
                            )}
                          </div>
                        </div>
                        <div className="flex flex-row sm:flex-col items-center sm:items-end gap-2 sm:gap-1 flex-shrink-0">
                          <Badge variant={getEventBadgeVariant(event.type)} className="text-xs whitespace-nowrap">
                            {formatEventType(event.type)}
                          </Badge>
                          <div className="flex items-center gap-1.5 text-xs text-[#6e7681]">
                            <Clock className="h-3 w-3 flex-shrink-0" />
                            <span className="whitespace-nowrap">
                              {formatDistanceToNow(new Date(event.created_at), { addSuffix: true })}
                            </span>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

          {/* Load More Button - uses pre-computed hasMore flag */}
          {hasMore && (
            <div className="flex justify-center">
              <Button 
                variant="outline" 
                onClick={handleLoadMore}
                className="bg-[#161b22] border-[#30363d] text-white hover:bg-[#1f2937]"
              >
                Load More
              </Button>
            </div>
          )}
        </>
      )}

      {/* Stats - uses pre-computed stats object */}
      <Card className="border-[#30363d] bg-[#161b22]">
        <CardContent className="p-6">
          <div className="grid gap-6 md:grid-cols-4">
            <div>
              <div className="text-2xl font-bold text-white">{stats.total}</div>
              <p className="text-sm text-[#8b949e]">Total Events</p>
            </div>
            <div>
              <div className="text-2xl font-bold text-white">{stats.completed}</div>
              <p className="text-sm text-[#8b949e]">Tasks Completed</p>
            </div>
            <div>
              <div className="text-2xl font-bold text-white">{stats.created}</div>
              <p className="text-sm text-[#8b949e]">Tasks Created</p>
            </div>
            <div>
              <div className="text-2xl font-bold text-white">{stats.spawned}</div>
              <p className="text-sm text-[#8b949e]">Agents Spawned</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
