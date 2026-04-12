import { useState, useRef } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Upload, Download, ChevronDown, ChevronUp, AlertTriangle, CheckCircle, X, FileText } from 'lucide-react'
import { Button } from '../../components/ui/Button'
import { userApi, type AlertChannel, type ImportChannelRow, type ImportResult } from '../../api/user'

// ─── CSV helpers ──────────────────────────────────────────────────────────────

const TEMPLATE_CSV = `name,type,value,enabled
My Email Alert,email,alerts@example.com,true
Team Slack,slack,https://hooks.slack.com/services/REPLACE,true
Discord Alerts,discord,https://discord.com/api/webhooks/REPLACE,false`

function downloadBlob(content: string, filename: string, mime: string) {
  const blob = new Blob([content], { type: mime })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

function channelsToCSV(channels: AlertChannel[]): string {
  const rows = channels.map(ch => {
    const value = ch.type === 'email' ? ch.config.email : ch.config.webhook_url
    const name = ch.name.includes(',') ? `"${ch.name}"` : ch.name
    return `${name},${ch.type},${value ?? ''},${ch.is_enabled}`
  })
  return ['name,type,value,enabled', ...rows].join('\n')
}

function channelsToJSON(channels: AlertChannel[]): string {
  const rows = channels.map(ch => ({
    name: ch.name,
    type: ch.type,
    value: ch.type === 'email' ? ch.config.email : ch.config.webhook_url,
    enabled: ch.is_enabled,
  }))
  return JSON.stringify(rows, null, 2)
}

// ─── CSV parser ───────────────────────────────────────────────────────────────

function parseCSVLine(line: string): string[] {
  const result: string[] = []
  let current = ''
  let inQuotes = false
  for (let i = 0; i < line.length; i++) {
    const ch = line[i]
    if (ch === '"') {
      inQuotes = !inQuotes
    } else if (ch === ',' && !inQuotes) {
      result.push(current.trim())
      current = ''
    } else {
      current += ch
    }
  }
  result.push(current.trim())
  return result
}

function parseFile(text: string, filename: string): ImportChannelRow[] | string {
  const ext = filename.split('.').pop()?.toLowerCase()

  if (ext === 'json') {
    try {
      const parsed = JSON.parse(text)
      if (!Array.isArray(parsed)) return 'JSON must be an array of channel objects'
      return parsed.map((r: any) => ({
        name:    String(r.name ?? ''),
        type:    String(r.type ?? ''),
        value:   String(r.value ?? ''),
        enabled: r.enabled !== false,
      }))
    } catch {
      return 'Invalid JSON — could not parse file'
    }
  }

  if (ext === 'csv') {
    const lines = text.split('\n').map(l => l.trim()).filter(Boolean)
    if (lines.length < 2) return 'CSV must have a header row and at least one data row'
    const header = parseCSVLine(lines[0].toLowerCase())
    const idx = {
      name:    header.indexOf('name'),
      type:    header.indexOf('type'),
      value:   header.indexOf('value'),
      enabled: header.indexOf('enabled'),
    }
    if (idx.name === -1 || idx.type === -1 || idx.value === -1) {
      return 'CSV is missing required columns: name, type, value'
    }
    return lines.slice(1).map(line => {
      const cols = parseCSVLine(line)
      return {
        name:    cols[idx.name]    ?? '',
        type:    cols[idx.type]    ?? '',
        value:   cols[idx.value]   ?? '',
        enabled: idx.enabled === -1 ? true : cols[idx.enabled]?.toLowerCase() !== 'false',
      }
    })
  }

  return 'Unsupported file type — use .json or .csv'
}

// ─── Per-row validation (mirrors backend) ────────────────────────────────────

function validateRow(row: ImportChannelRow): string | null {
  if (!row.name) return 'name is required'
  if (!['email', 'slack', 'discord'].includes(row.type)) return `unsupported type "${row.type}"`
  if (row.type === 'email') {
    if (!row.value || !row.value.includes('@')) return 'invalid email address'
  } else {
    if (!row.value || !row.value.startsWith('https://')) return 'invalid webhook URL'
  }
  return null
}

function conflictKey(row: ImportChannelRow) {
  return `${row.type}:${row.value}`
}

// ─── Preview row status ───────────────────────────────────────────────────────

type RowStatus = 'new' | 'conflict' | 'invalid'

interface PreviewRow extends ImportChannelRow {
  rowNum: number
  status: RowStatus
  error?: string
}

function buildPreview(rows: ImportChannelRow[], existing: AlertChannel[]): PreviewRow[] {
  const existingKeys = new Set(
    existing.map(ch => `${ch.type}:${ch.type === 'email' ? ch.config.email : ch.config.webhook_url}`)
  )
  return rows.map((row, i) => {
    const error = validateRow(row)
    if (error) return { ...row, rowNum: i + 1, status: 'invalid', error }
    if (existingKeys.has(conflictKey(row))) return { ...row, rowNum: i + 1, status: 'conflict' }
    return { ...row, rowNum: i + 1, status: 'new' }
  })
}

// ─── Docs panel ───────────────────────────────────────────────────────────────

function DocsPanel() {
  const [open, setOpen] = useState(false)
  return (
    <div className="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
      <button
        onClick={() => setOpen(o => !o)}
        className="w-full flex items-center justify-between px-4 py-3 bg-gray-50 dark:bg-gray-800/50 text-sm font-medium text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
      >
        <span className="flex items-center gap-2"><FileText size={14} /> How to format your file</span>
        {open ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
      </button>

      {open && (
        <div className="px-4 py-4 space-y-4 text-sm text-gray-600 dark:text-gray-400 bg-white dark:bg-gray-900">
          <div>
            <p className="font-semibold text-gray-700 dark:text-gray-300 mb-1">CSV format</p>
            <pre className="bg-gray-50 dark:bg-gray-800 rounded p-3 text-xs overflow-x-auto leading-relaxed">{`name,type,value,enabled
My Email Alert,email,you@example.com,true
Team Slack,slack,https://hooks.slack.com/...,true
Discord Alerts,discord,https://discord.com/api/webhooks/...,false`}</pre>
          </div>

          <div>
            <p className="font-semibold text-gray-700 dark:text-gray-300 mb-1">JSON format</p>
            <pre className="bg-gray-50 dark:bg-gray-800 rounded p-3 text-xs overflow-x-auto leading-relaxed">{`[
  { "name": "My Email", "type": "email", "value": "you@example.com", "enabled": true },
  { "name": "Slack",    "type": "slack", "value": "https://hooks.slack.com/...", "enabled": true }
]`}</pre>
          </div>

          <ul className="space-y-1 text-xs list-disc list-inside">
            <li><code className="font-mono">type</code> must be <code className="font-mono">email</code>, <code className="font-mono">slack</code>, or <code className="font-mono">discord</code></li>
            <li><code className="font-mono">value</code> is the email address for email, webhook URL for Slack/Discord</li>
            <li><code className="font-mono">enabled</code> is optional — defaults to <code className="font-mono">true</code></li>
            <li>Names with commas must be wrapped in double quotes in CSV</li>
          </ul>

          <div className="flex items-start gap-2 text-xs bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded p-2.5">
            <AlertTriangle size={13} className="text-amber-500 shrink-0 mt-0.5" />
            <span className="text-amber-700 dark:text-amber-400">Your file contains webhook URLs — treat it like a password. Don't share it or commit it to version control.</span>
          </div>

          <button
            onClick={() => downloadBlob(TEMPLATE_CSV, 'channels-template.csv', 'text/csv')}
            className="text-xs text-indigo-600 dark:text-indigo-400 hover:underline flex items-center gap-1"
          >
            <Download size={12} /> Download template CSV
          </button>
        </div>
      )}
    </div>
  )
}

// ─── Main modal ───────────────────────────────────────────────────────────────

interface Props {
  channels: AlertChannel[]
  onClose: () => void
  defaultTab?: 'import' | 'export'
}

export function ImportExportModal({ channels, onClose, defaultTab = 'import' }: Props) {
  const queryClient = useQueryClient()
  const [tab, setTab] = useState<'import' | 'export'>(defaultTab)

  // import state
  const fileRef = useRef<HTMLInputElement>(null)
  const [parseError, setParseError] = useState<string | null>(null)
  const [preview, setPreview] = useState<PreviewRow[] | null>(null)
  const [onConflict, setOnConflict] = useState<'skip' | 'overwrite'>('skip')
  const [result, setResult] = useState<ImportResult | null>(null)

  const conflictCount = preview?.filter(r => r.status === 'conflict').length ?? 0
  const invalidCount  = preview?.filter(r => r.status === 'invalid').length ?? 0
  const validCount    = preview?.filter(r => r.status !== 'invalid').length ?? 0

  const importMutation = useMutation({
    mutationFn: () => {
      const rows = preview!.filter(r => r.status !== 'invalid').map(({ name, type, value, enabled }) => ({ name, type, value, enabled }))
      return userApi.importAlertChannels(rows, onConflict).then(r => r.data!)
    },
    onSuccess: (data) => {
      setResult(data)
      queryClient.invalidateQueries({ queryKey: ['alert-channels'] })
    },
  })

  function handleFile(file: File) {
    setParseError(null)
    setPreview(null)
    setResult(null)
    const reader = new FileReader()
    reader.onload = e => {
      const text = e.target?.result as string
      const parsed = parseFile(text, file.name)
      if (typeof parsed === 'string') {
        setParseError(parsed)
      } else {
        setPreview(buildPreview(parsed, channels))
      }
    }
    reader.readAsText(file)
  }

  const statusBadge = (s: RowStatus) => {
    if (s === 'new')      return <span className="text-xs px-1.5 py-0.5 rounded bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400">New</span>
    if (s === 'conflict') return <span className="text-xs px-1.5 py-0.5 rounded bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400">Conflict</span>
    return <span className="text-xs px-1.5 py-0.5 rounded bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400">Invalid</span>
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="w-full max-w-2xl bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-xl max-h-[90vh] flex flex-col">

        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100 dark:border-gray-800 shrink-0">
          <div className="flex gap-1 bg-gray-100 dark:bg-gray-800 rounded-lg p-1">
            {(['import', 'export'] as const).map(t => (
              <button
                key={t}
                onClick={() => setTab(t)}
                className={`px-4 py-1.5 rounded-md text-sm font-medium transition-colors capitalize ${
                  tab === t
                    ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-white shadow-sm'
                    : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
                }`}
              >
                {t}
              </button>
            ))}
          </div>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200">
            <X size={18} />
          </button>
        </div>

        {/* Body */}
        <div className="flex-1 overflow-y-auto p-6 space-y-4">

          {/* ── Export tab ── */}
          {tab === 'export' && (
            <div className="space-y-4">
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Download all {channels.length} alert channel{channels.length !== 1 ? 's' : ''} (including disabled ones).
                You can use this file to back up your channels or import them into another account.
              </p>

              <div className="grid grid-cols-2 gap-3">
                <button
                  onClick={() => downloadBlob(channelsToCSV(channels), 'alert-channels.csv', 'text/csv')}
                  disabled={channels.length === 0}
                  className="flex flex-col items-center gap-2 p-5 border-2 border-dashed border-gray-200 dark:border-gray-700 rounded-xl hover:border-indigo-400 hover:bg-indigo-50/50 dark:hover:bg-indigo-900/10 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
                >
                  <Download size={22} className="text-indigo-500" />
                  <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Export as CSV</span>
                  <span className="text-xs text-gray-400">Open in Excel, Google Sheets</span>
                </button>
                <button
                  onClick={() => downloadBlob(channelsToJSON(channels), 'alert-channels.json', 'application/json')}
                  disabled={channels.length === 0}
                  className="flex flex-col items-center gap-2 p-5 border-2 border-dashed border-gray-200 dark:border-gray-700 rounded-xl hover:border-indigo-400 hover:bg-indigo-50/50 dark:hover:bg-indigo-900/10 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
                >
                  <Download size={22} className="text-indigo-500" />
                  <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Export as JSON</span>
                  <span className="text-xs text-gray-400">Round-trip import ready</span>
                </button>
              </div>

              {channels.length === 0 && (
                <p className="text-center text-sm text-gray-400 py-4">No channels to export yet.</p>
              )}

              <DocsPanel />
            </div>
          )}

          {/* ── Import tab ── */}
          {tab === 'import' && (
            <div className="space-y-4">
              {!result ? (
                <>
                  {/* File drop zone */}
                  <div
                    className="border-2 border-dashed border-gray-200 dark:border-gray-700 rounded-xl p-8 text-center cursor-pointer hover:border-indigo-400 hover:bg-indigo-50/30 dark:hover:bg-indigo-900/10 transition-colors"
                    onClick={() => fileRef.current?.click()}
                    onDragOver={e => e.preventDefault()}
                    onDrop={e => { e.preventDefault(); const f = e.dataTransfer.files[0]; if (f) handleFile(f) }}
                  >
                    <Upload size={24} className="mx-auto mb-2 text-gray-400" />
                    <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Drop a file here or click to browse</p>
                    <p className="text-xs text-gray-400 mt-1">Supports .csv and .json</p>
                    <input ref={fileRef} type="file" accept=".csv,.json" className="hidden" onChange={e => { const f = e.target.files?.[0]; if (f) handleFile(f) }} />
                  </div>

                  {parseError && (
                    <div className="flex items-start gap-2 text-sm text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg px-4 py-3">
                      <AlertTriangle size={15} className="shrink-0 mt-0.5" />
                      {parseError}
                    </div>
                  )}

                  {/* Preview table */}
                  {preview && preview.length > 0 && (
                    <div className="space-y-3">
                      <div className="flex items-center gap-3 text-xs text-gray-500 dark:text-gray-400">
                        <span className="text-green-600 dark:text-green-400 font-medium">{preview.filter(r => r.status === 'new').length} new</span>
                        {conflictCount > 0 && <span className="text-amber-600 dark:text-amber-400 font-medium">{conflictCount} conflict{conflictCount > 1 ? 's' : ''}</span>}
                        {invalidCount > 0  && <span className="text-red-600 dark:text-red-400 font-medium">{invalidCount} invalid</span>}
                      </div>

                      <div className="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
                        <div className="grid grid-cols-[2fr_1fr_2fr_auto] gap-3 px-4 py-2 bg-gray-50 dark:bg-gray-800/50 text-xs font-medium text-gray-400 uppercase tracking-wide">
                          <span>Name</span><span>Type</span><span>Value</span><span>Status</span>
                        </div>
                        <div className="divide-y divide-gray-100 dark:divide-gray-800 max-h-48 overflow-y-auto">
                          {preview.map(row => (
                            <div key={row.rowNum} className={`grid grid-cols-[2fr_1fr_2fr_auto] gap-3 items-center px-4 py-2.5 text-sm ${row.status === 'invalid' ? 'bg-red-50/50 dark:bg-red-900/10' : ''}`}>
                              <span className="truncate text-gray-800 dark:text-gray-200">{row.name || <span className="text-gray-400 italic">empty</span>}</span>
                              <span className="text-gray-500 dark:text-gray-400">{row.type}</span>
                              <span className="truncate text-gray-400 text-xs" title={row.value}>{row.value || '—'}</span>
                              <div className="flex flex-col items-end gap-0.5">
                                {statusBadge(row.status)}
                                {row.error && <span className="text-xs text-red-500">{row.error}</span>}
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>

                      {/* Conflict option */}
                      {conflictCount > 0 && (
                        <div className="space-y-2">
                          <p className="text-sm font-medium text-gray-700 dark:text-gray-300">
                            {conflictCount} channel{conflictCount > 1 ? 's' : ''} already exist{conflictCount === 1 ? 's' : ''}. What should happen?
                          </p>
                          <label className="flex items-start gap-2.5 cursor-pointer group">
                            <input type="radio" name="conflict" value="skip" checked={onConflict === 'skip'} onChange={() => setOnConflict('skip')} className="mt-0.5 text-indigo-600" />
                            <div>
                              <p className="text-sm text-gray-700 dark:text-gray-300 font-medium">Skip conflicts</p>
                              <p className="text-xs text-gray-400">Keep existing channels unchanged, only import new ones</p>
                            </div>
                          </label>
                          <label className="flex items-start gap-2.5 cursor-pointer group">
                            <input type="radio" name="conflict" value="overwrite" checked={onConflict === 'overwrite'} onChange={() => setOnConflict('overwrite')} className="mt-0.5 text-indigo-600" />
                            <div>
                              <p className="text-sm text-gray-700 dark:text-gray-300 font-medium">Overwrite conflicts</p>
                              <p className="text-xs text-gray-400">Update the name and enabled state of existing channels</p>
                            </div>
                          </label>
                        </div>
                      )}
                    </div>
                  )}

                  <DocsPanel />
                </>
              ) : (
                /* Result screen */
                <div className="space-y-4">
                  <div className="flex items-center gap-3">
                    <CheckCircle size={22} className="text-green-500 shrink-0" />
                    <p className="text-base font-semibold text-gray-900 dark:text-white">Import complete</p>
                  </div>

                  <div className="grid grid-cols-3 gap-3">
                    {[
                      { label: 'Imported',    value: result.imported,    color: 'text-green-600 dark:text-green-400' },
                      { label: 'Overwritten', value: result.overwritten, color: 'text-blue-600 dark:text-blue-400'  },
                      { label: 'Skipped',     value: result.skipped,     color: 'text-amber-600 dark:text-amber-400'},
                    ].map(({ label, value, color }) => (
                      <div key={label} className="text-center bg-gray-50 dark:bg-gray-800 rounded-xl py-4">
                        <p className={`text-2xl font-bold ${color}`}>{value}</p>
                        <p className="text-xs text-gray-400 mt-0.5">{label}</p>
                      </div>
                    ))}
                  </div>

                  {result.errors.length > 0 && (
                    <div className="space-y-1.5">
                      <p className="text-sm font-medium text-red-600 dark:text-red-400 flex items-center gap-1.5">
                        <AlertTriangle size={14} /> {result.errors.length} row{result.errors.length > 1 ? 's' : ''} had errors
                      </p>
                      <div className="border border-red-200 dark:border-red-800 rounded-lg divide-y divide-red-100 dark:divide-red-900 max-h-40 overflow-y-auto">
                        {result.errors.map(e => (
                          <div key={e.row} className="px-3 py-2 text-xs flex items-center gap-3">
                            <span className="text-gray-400 shrink-0">Row {e.row}</span>
                            <span className="font-medium text-gray-700 dark:text-gray-300 truncate">{e.name || '(empty)'}</span>
                            <span className="text-red-500 ml-auto shrink-0">{e.reason}</span>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              )}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex justify-end gap-3 px-6 py-4 border-t border-gray-100 dark:border-gray-800 shrink-0">
          {tab === 'import' && !result && preview && validCount > 0 && (
            <Button
              onClick={() => importMutation.mutate()}
              loading={importMutation.isPending}
            >
              Import {validCount} channel{validCount !== 1 ? 's' : ''}
            </Button>
          )}
          <Button variant="secondary" onClick={onClose}>
            {result ? 'Close' : 'Cancel'}
          </Button>
        </div>
      </div>
    </div>
  )
}
