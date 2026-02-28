'use client';

import React, { useState, useEffect, useRef, useCallback } from 'react';
import { chatApi } from '@/services/api';
import type { ChatSession, ChatMessage } from '@/types';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Card } from '@/components/ui/card';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Loader2, Send, MessageSquare, StopCircle } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';

interface AgentChatProps {
  agentId: string;
}

export function AgentChat({ agentId }: AgentChatProps) {
  const [sessions, setSessions] = useState<ChatSession[]>([]);
  const [currentSession, setCurrentSession] = useState<ChatSession | null>(null);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [inputValue, setInputValue] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [isSending, setIsSending] = useState(false);
  const [isStarting, setIsStarting] = useState(false);
  const scrollRef = useRef<HTMLDivElement>(null);
  
  /** Track mounted state to prevent state updates after unmount */
  const isMountedRef = useRef(true);
  /** Track active poll timer for cleanup */
  const pollTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  /** Use ref for messages in polling to avoid stale closures */
  const messagesRef = useRef<ChatMessage[]>([]);

  // Keep messagesRef in sync
  useEffect(() => {
    messagesRef.current = messages;
  }, [messages]);

  // Cleanup on unmount
  useEffect(() => {
    isMountedRef.current = true;
    return () => {
      isMountedRef.current = false;
      if (pollTimerRef.current) {
        clearTimeout(pollTimerRef.current);
        pollTimerRef.current = null;
      }
    };
  }, []);

  // Load sessions on mount
  useEffect(() => {
    loadSessions();
  }, [agentId]); // eslint-disable-line react-hooks/exhaustive-deps

  // Load messages when session changes
  useEffect(() => {
    if (currentSession) {
      loadMessages(currentSession.id);
    }
  }, [currentSession?.id]); // eslint-disable-line react-hooks/exhaustive-deps

  // Auto-scroll to bottom when messages change
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages]);

  const loadSessions = useCallback(async () => {
    try {
      setIsLoading(true);
      const data = await chatApi.listSessions(agentId);
      if (!isMountedRef.current) return;
      setSessions(data);
      
      // Auto-select active session if exists
      const activeSession = data.find((s: ChatSession) => s.status === 'active');
      if (activeSession) {
        setCurrentSession(activeSession);
      }
    } catch (error) {
      console.error('[AgentChat] Failed to load sessions:', error);
    } finally {
      if (isMountedRef.current) setIsLoading(false);
    }
  }, [agentId]);

  const loadMessages = useCallback(async (sessionId: string) => {
    try {
      setIsLoading(true);
      const data = await chatApi.getMessages(agentId, sessionId);
      if (!isMountedRef.current) return;
      setMessages(data);
    } catch (error) {
      console.error('[AgentChat] Failed to load messages:', error);
    } finally {
      if (isMountedRef.current) setIsLoading(false);
    }
  }, [agentId]);

  const startNewSession = useCallback(async () => {
    try {
      setIsStarting(true);
      const newSession = await chatApi.startSession(agentId);
      if (!isMountedRef.current) return;
      setSessions(prev => [newSession, ...prev]);
      setCurrentSession(newSession);
      setMessages([]);
    } catch (error) {
      console.error('[AgentChat] Failed to start session:', error);
    } finally {
      if (isMountedRef.current) setIsStarting(false);
    }
  }, [agentId]);

  const endCurrentSession = useCallback(async () => {
    if (!currentSession) return;
    if (!confirm('Are you sure you want to end this session?')) return;

    try {
      await chatApi.endSession(agentId, currentSession.id);
      if (!isMountedRef.current) return;
      await loadSessions();
      setCurrentSession(null);
      setMessages([]);
    } catch (error) {
      console.error('[AgentChat] Failed to end session:', error);
    }
  }, [agentId, currentSession, loadSessions]);

  /**
   * Polls for agent responses using messagesRef to avoid stale closure.
   * Properly cleans up on unmount via isMountedRef.
   */
  const pollForResponse = useCallback(async (sessionId: string, attempts = 0) => {
    if (!isMountedRef.current) return;
    if (attempts > 15) { // Max 30 seconds (15 * 2s)
      if (isMountedRef.current) setIsSending(false);
      return;
    }

    try {
      const newMessages = await chatApi.getMessages(agentId, sessionId);
      if (!isMountedRef.current) return;

      // Use ref to get current messages (avoids stale closure)
      const currentMessages = messagesRef.current;
      const hasNewAgentMessage = newMessages.some(
        (m: ChatMessage) => m.role === 'agent' && !currentMessages.find(existing => existing.id === m.id)
      );

      if (hasNewAgentMessage) {
        setMessages(newMessages);
        setIsSending(false);
      } else {
        // Wait 2 seconds and poll again
        pollTimerRef.current = setTimeout(() => pollForResponse(sessionId, attempts + 1), 2000);
      }
    } catch (error) {
      console.error('[AgentChat] Failed to poll messages:', error);
      if (isMountedRef.current) setIsSending(false);
    }
  }, [agentId]);

  const sendMessage = useCallback(async () => {
    if (!currentSession || !inputValue.trim() || isSending) return;

    const content = inputValue.trim();
    const userMessage: ChatMessage = {
      id: `temp-${Date.now()}`,
      session_id: currentSession.id,
      role: 'user',
      content: content,
      created_at: new Date().toISOString(),
    };

    setMessages(prev => [...prev, userMessage]);
    setInputValue('');
    setIsSending(true);

    try {
      await chatApi.sendMessage(agentId, currentSession.id, content);
      // Start polling for agent response
      pollForResponse(currentSession.id);
    } catch (error) {
      console.error('[AgentChat] Failed to send message:', error);
      // Remove temp message on error
      setMessages(prev => prev.filter(m => m.id !== userMessage.id));
      setIsSending(false);
    }
  }, [agentId, currentSession, inputValue, isSending, pollForResponse]);

  const handleKeyPress = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  }, [sendMessage]);

  const handleSessionChange = useCallback((sessionId: string) => {
    const session = sessions.find(s => s.id === sessionId);
    if (session) {
      setCurrentSession(session);
    }
  }, [sessions]);

  return (
    <div className="flex flex-col h-full space-y-4">
      {/* Header with session selector and controls */}
      <div className="flex items-center gap-2">
        <Select
          value={currentSession?.id || ''}
          onValueChange={handleSessionChange}
          disabled={isLoading || sessions.length === 0}
        >
          <SelectTrigger className="flex-1 bg-slate-900 border-slate-800">
            <SelectValue placeholder="Select a session..." />
          </SelectTrigger>
          <SelectContent className="bg-slate-900 border-slate-800">
            {sessions.map(session => (
              <SelectItem key={session.id} value={session.id}>
                <div className="flex items-center gap-2">
                  <span className={`h-2 w-2 rounded-full ${
                    session.status === 'active' ? 'bg-green-500' : 'bg-slate-500'
                  }`} />
                  <span>
                    {formatDistanceToNow(new Date(session.started_at), { addSuffix: true })}
                  </span>
                  <span className="text-slate-500">({session.message_count} msgs)</span>
                </div>
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        {currentSession?.status === 'active' ? (
          <Button
            variant="destructive"
            size="sm"
            onClick={endCurrentSession}
            disabled={isSending}
          >
            <StopCircle className="h-4 w-4 mr-1" />
            End
          </Button>
        ) : (
          <Button
            variant="default"
            size="sm"
            onClick={startNewSession}
            disabled={isStarting}
          >
            {isStarting ? (
              <Loader2 className="h-4 w-4 mr-1 animate-spin" />
            ) : (
              <MessageSquare className="h-4 w-4 mr-1" />
            )}
            Start
          </Button>
        )}
      </div>

      {/* Messages area */}
      <Card className="flex-1 border-slate-800 bg-slate-900 overflow-hidden flex flex-col">
        <ScrollArea className="flex-1 p-4">
          {isLoading && messages.length === 0 ? (
            <div className="flex items-center justify-center h-full text-slate-400">
              <Loader2 className="h-6 w-6 animate-spin" />
            </div>
          ) : messages.length === 0 ? (
            <div className="flex items-center justify-center h-full text-slate-400">
              {currentSession ? 'No messages yet. Start chatting!' : 'Start a session to begin chatting'}
            </div>
          ) : (
            <div className="space-y-4">
              {messages.map((message, index) => {
                const isUser = message.role === 'user';
                const showSessionDivider = index === 0 || 
                  messages[index - 1].session_id !== message.session_id;

                return (
                  <React.Fragment key={message.id}>
                    {showSessionDivider && index > 0 && (
                      <div className="flex items-center gap-2 py-2">
                        <div className="flex-1 border-t border-slate-700" />
                        <span className="text-xs text-slate-500">New Session</span>
                        <div className="flex-1 border-t border-slate-700" />
                      </div>
                    )}
                    
                    <div className={`flex ${isUser ? 'justify-end' : 'justify-start'}`}>
                      <div
                        className={`max-w-[80%] rounded-lg px-4 py-2 ${
                          isUser
                            ? 'bg-blue-600 text-white'
                            : 'bg-slate-800 text-slate-100'
                        }`}
                      >
                        <div className="text-sm whitespace-pre-wrap break-words">
                          {message.content}
                        </div>
                        <div
                          className={`text-xs mt-1 ${
                            isUser ? 'text-blue-200' : 'text-slate-500'
                          }`}
                        >
                          {formatDistanceToNow(new Date(message.created_at), {
                            addSuffix: true,
                          })}
                        </div>
                      </div>
                    </div>
                  </React.Fragment>
                );
              })}
              <div ref={scrollRef} />
            </div>
          )}
        </ScrollArea>
      </Card>

      {/* Input area */}
      <div className="flex items-center gap-2">
        <Input
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          onKeyDown={handleKeyPress}
          placeholder={
            currentSession?.status === 'active'
              ? 'Type a message...'
              : 'Start a session to send messages'
          }
          disabled={!currentSession || currentSession.status !== 'active' || isSending}
          className="flex-1 bg-slate-900 border-slate-800 text-white placeholder:text-slate-500"
        />
        <Button
          onClick={sendMessage}
          disabled={!currentSession || currentSession.status !== 'active' || !inputValue.trim() || isSending}
          size="icon"
        >
          {isSending ? (
            <Loader2 className="h-4 w-4 animate-spin" />
          ) : (
            <Send className="h-4 w-4" />
          )}
        </Button>
      </div>
    </div>
  );
}
