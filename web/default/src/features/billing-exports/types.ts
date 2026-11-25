export type BillingExportStatus = 'pending' | 'running' | 'succeeded' | 'failed'

export type BillingExportLanguage = 'zh' | 'en'

export type BillingExportLevel = 'zip' | 'l0' | 'l1' | 'l2' | 'l3'

export type BillingExportFilters = {
  startMonth: string
  endMonth: string
  language: BillingExportLanguage
  organizationId: number | null
  userId: number | null
}

export type BillingExport = {
  task_id: string
  user_id?: number
  organization_id?: number
  start_month: string
  end_month: string
  language?: BillingExportLanguage
  created_at: number
  expires_at: number
  status: BillingExportStatus
  error?: string
}

export type BillingExportResponse = {
  success: boolean
  message?: string
  data: BillingExport[]
}
