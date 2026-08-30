import { api } from '@/lib/api'
import type { TaskPluginsResponse, TaskPlugin } from './types'

export const getTaskPlugins = async () =>
  (await api.get<TaskPluginsResponse>('/api/task-plugins')).data

export const updateTaskPlugin = async (plugin: TaskPlugin, enabled: boolean) =>
  (await api.patch<TaskPluginsResponse>(`/api/task-plugins/${encodeURIComponent(plugin.key)}`, { enabled })).data
