'use client';
import { useState, useEffect } from 'react';

const QUICK_PICKS = [
  { label: '+15m', minutes: 15 },
  { label: '+30m', minutes: 30 },
  { label: '+1h',  minutes: 60 },
  { label: '+2h',  minutes: 120 },
  { label: '+4h',  minutes: 240 },
  { label: '+1 day', minutes: 1440 },
];

interface SchedulePickerProps {
  value: string;
  onChange: (iso: string) => void;
  minDate?: Date;
  label?: string;
  className?: string;
}

function toDateTimeInputs(iso: string): { date: string; time: string } {
  const d = iso ? new Date(iso) : new Date();
  const date = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
  const rawMin = d.getMinutes();
  const roundedMin = rawMin % 5 === 0 ? rawMin : Math.ceil(rawMin / 5) * 5;
  const hh = String(d.getHours()).padStart(2, '0');
  const mm = String(roundedMin % 60).padStart(2, '0');
  return { date, time: `${hh}:${mm}` };
}

function combineToISO(date: string, time: string): string {
  if (!date || !time) return '';
  return new Date(`${date}T${time}`).toISOString();
}

function formatPreview(iso: string): { text: string; isPast: boolean } {
  if (!iso) return { text: '', isPast: false };
  const target = new Date(iso);
  const now = new Date();
  const diffMs = target.getTime() - now.getTime();
  if (diffMs < 0) return { text: 'This time is in the past', isPast: true };

  const diffMin = Math.round(diffMs / 60000);
  let relative: string;
  if (diffMin < 60) {
    relative = `In ${diffMin} minute${diffMin !== 1 ? 's' : ''}`;
  } else if (diffMin < 1440) {
    const hrs = Math.round(diffMin / 60);
    relative = `In ${hrs} hour${hrs !== 1 ? 's' : ''}`;
  } else {
    const days = Math.round(diffMin / 1440);
    relative = `In ${days} day${days !== 1 ? 's' : ''}`;
  }

  const formatted = target.toLocaleString(undefined, {
    weekday: 'short', month: 'short', day: 'numeric',
    hour: 'numeric', minute: '2-digit',
  });

  return { text: `${relative} — ${formatted}`, isPast: false };
}

export function SchedulePicker({ value, onChange, label, className }: SchedulePickerProps) {
  const getInitial = () => {
    if (value) return toDateTimeInputs(value);
    const d = new Date(Date.now() + 60 * 60 * 1000);
    return toDateTimeInputs(d.toISOString());
  };

  const [inputs, setInputs] = useState<{ date: string; time: string }>(getInitial);
  const [activeChip, setActiveChip] = useState<number | null>(60);

  // Emit initial value on mount if value was empty
  useEffect(() => {
    if (!value) {
      const d = new Date(Date.now() + 60 * 60 * 1000);
      onChange(d.toISOString());
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleChip = (minutes: number) => {
    const d = new Date(Date.now() + minutes * 60 * 1000);
    const next = toDateTimeInputs(d.toISOString());
    setInputs(next);
    setActiveChip(minutes);
    onChange(combineToISO(next.date, next.time));
  };

  const handleDateChange = (newDate: string) => {
    setInputs(prev => ({ ...prev, date: newDate }));
    setActiveChip(null);
    onChange(combineToISO(newDate, inputs.time));
  };

  const handleTimeChange = (newTime: string) => {
    setInputs(prev => ({ ...prev, time: newTime }));
    setActiveChip(null);
    onChange(combineToISO(inputs.date, newTime));
  };

  const iso = combineToISO(inputs.date, inputs.time);
  const preview = formatPreview(iso);
  // Use local date for min attribute (not UTC, to match user's timezone)
  const today = new Date();
  const localMinDate = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`;

  return (
    <div className={`space-y-2 ${className ?? ''}`}>
      {label && <p className="text-xs font-medium text-slate-400">{label}</p>}

      {/* Quick-select chips */}
      <div className="flex flex-wrap gap-1">
        {QUICK_PICKS.map(({ label: chipLabel, minutes }) => (
          <button
            key={minutes}
            type="button"
            onClick={() => handleChip(minutes)}
            className={`px-2 py-0.5 rounded text-xs font-medium transition-colors ${
              activeChip === minutes
                ? 'bg-blue-600 text-white'
                : 'bg-[#21262d] text-slate-300 hover:bg-[#30363d] border border-[#30363d]'
            }`}
          >
            {chipLabel}
          </button>
        ))}
      </div>

      {/* Date + Time inputs */}
      <div className="flex gap-2">
        <input
          type="date"
          value={inputs.date}
          min={localMinDate}
          onChange={e => handleDateChange(e.target.value)}
          className="flex-1 rounded border border-[#30363d] bg-[#0d1117] px-2 py-1 text-xs text-white [color-scheme:dark] focus:outline-none focus:border-blue-500"
        />
        <input
          type="time"
          value={inputs.time}
          onChange={e => handleTimeChange(e.target.value)}
          className="w-24 rounded border border-[#30363d] bg-[#0d1117] px-2 py-1 text-xs text-white [color-scheme:dark] focus:outline-none focus:border-blue-500"
        />
      </div>

      {/* Preview */}
      {iso && (
        <p className={`text-xs ${preview.isPast ? 'text-red-400' : 'text-slate-400'}`}>
          {preview.isPast ? '⚠ ' : '⏰ '}{preview.text}
        </p>
      )}
    </div>
  );
}
