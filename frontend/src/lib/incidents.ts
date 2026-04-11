import type { IncidentStatus } from '../api/incidents'

export const STATUS_LABEL: Record<IncidentStatus, string> = {
  investigating: 'Investigating',
  identified:    'Identified',
  monitoring:    'Monitoring',
  resolved:      'Resolved',
}

// Pill badge — used inline next to incident names
export const STATUS_COLOR: Record<IncidentStatus, string> = {
  investigating: 'bg-red-50 text-red-700 border border-red-200 dark:bg-red-900/20 dark:text-red-400 dark:border-red-800/50',
  identified:    'bg-orange-50 text-orange-700 border border-orange-200 dark:bg-orange-900/20 dark:text-orange-400 dark:border-orange-800/50',
  monitoring:    'bg-yellow-50 text-yellow-700 border border-yellow-200 dark:bg-yellow-900/20 dark:text-yellow-400 dark:border-yellow-800/50',
  resolved:      'bg-green-50 text-green-700 border border-green-200 dark:bg-green-900/20 dark:text-green-400 dark:border-green-800/50',
}

// Colour dot for timeline / table
export const STATUS_DOT: Record<IncidentStatus, string> = {
  investigating: 'bg-red-500',
  identified:    'bg-orange-500',
  monitoring:    'bg-yellow-500',
  resolved:      'bg-green-500',
}
