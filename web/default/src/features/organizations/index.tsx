import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { Building2, Copy, KeyRound, MailPlus, Plus } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Main } from '@/components/layout'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { NativeSelect, NativeSelectOption } from '@/components/ui/native-select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  formatQuota,
  parseQuotaFromDollars,
  quotaUnitsToDollars,
} from '@/lib/format'
import { useAuthStore } from '@/stores/auth-store'

import {
  acceptOrganizationInvitation,
  createOrganization,
  createOrganizationToken,
  getMyOrganizationInvitations,
  getOrganization,
  getOrganizationInvitations,
  getOrganizationMembers,
  getOrganizationTokens,
  getOrganizationUsage,
  inviteOrganizationMember,
  getOrganizations,
  getOrganizationWalletTransactions,
  transferToOrganization,
  updateOrganizationMember,
  removeOrganizationMember,
  revokeOrganizationInvitation,
  revealOrganizationToken,
  updateOrganizationToken,
  deleteOrganizationToken,
  getOrganizationLogs,
  leaveOrganization,
  transferOrganizationOwnership,
} from './api'
import type { OrganizationMember } from './types'
import { useActiveOrganizationId, useWorkspaceStore } from './workspace-store'

export function Organizations() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const userId = useAuthStore((state) => state.auth.user?.id)
  const activeId = useActiveOrganizationId()
  const selectOrganization = useWorkspaceStore(
    (state) => state.selectOrganization
  )
  const selectorEntry = useRef(false)
  const [organizationName, setOrganizationName] = useState('')
  const [inviteEmail, setInviteEmail] = useState('')
  const [tokenName, setTokenName] = useState('')
  const [transferAmount, setTransferAmount] = useState('')
  const [logPage, setLogPage] = useState(1)
  const [confirmation, setConfirmation] = useState<{
    title: string
    description: string
    run: () => void
  } | null>(null)

  const organizations = useQuery({
    queryKey: ['organizations', userId],
    queryFn: getOrganizations,
    enabled: Boolean(userId),
  })
  const selectedId = organizations.isSuccess
    ? (organizations.data.find((entry) => entry.organization.id === activeId)
        ?.organization.id ?? null)
    : null
  useEffect(() => {
    if (selectorEntry.current) return
    selectorEntry.current = true
    if (userId && activeId !== null) selectOrganization(userId, null)
  }, [activeId, selectOrganization, userId])
  const organization = useQuery({
    queryKey: ['organization', userId, selectedId],
    queryFn: () => getOrganization(selectedId as number),
    enabled: selectedId !== null,
  })
  const membershipRole = organization.isSuccess
    ? organization.data.membership.role
    : undefined
  const canManage = membershipRole === 'owner' || membershipRole === 'admin'
  const members = useQuery({
    queryKey: ['organization-members', userId, selectedId],
    queryFn: () => getOrganizationMembers(selectedId as number),
    enabled: selectedId !== null && canManage,
  })
  const invitations = useQuery({
    queryKey: ['organization-invitations', userId, selectedId],
    queryFn: () => getOrganizationInvitations(selectedId as number),
    enabled: selectedId !== null && canManage,
  })
  const pendingInvitations = useQuery({
    queryKey: ['organization-invitations-pending', userId],
    queryFn: getMyOrganizationInvitations,
    enabled: Boolean(userId),
  })
  const tokens = useQuery({
    queryKey: ['organization-tokens', userId, selectedId, membershipRole],
    queryFn: () => getOrganizationTokens(selectedId as number),
    enabled: selectedId !== null && membershipRole !== undefined,
  })
  const usage = useQuery({
    queryKey: ['organization-usage', userId, selectedId, membershipRole],
    queryFn: () => getOrganizationUsage(selectedId as number),
    enabled: selectedId !== null && membershipRole !== undefined,
  })
  const transactions = useQuery({
    queryKey: ['organization-wallet-transactions', userId, selectedId],
    queryFn: () => getOrganizationWalletTransactions(selectedId as number),
    enabled: selectedId !== null && canManage,
  })

  const logs = useQuery({
    queryKey: [
      'organization-logs',
      userId,
      selectedId,
      membershipRole,
      logPage,
    ],
    queryFn: () => getOrganizationLogs(selectedId as number, logPage),
    enabled: selectedId !== null && membershipRole !== undefined,
  })
  const refresh = () => {
    void queryClient.invalidateQueries({ queryKey: ['organizations'] })
    void queryClient.invalidateQueries({ queryKey: ['organization'] })
    void queryClient.invalidateQueries({
      predicate: (query) =>
        String(query.queryKey[0]).startsWith('organization-'),
    })
  }
  const action = (run: () => Promise<unknown>, success: string) => {
    void run()
      .then(() => {
        toast.success(t(success))
        refresh()
      })
      .catch((error: unknown) =>
        toast.error(
          error instanceof Error ? error.message : t('Request failed')
        )
      )
  }
  const create = () => {
    if (!organizationName.trim()) return
    action(
      () => createOrganization(organizationName.trim()),
      'Organization created'
    )
    setOrganizationName('')
  }
  const invite = () => {
    if (!selectedId || !inviteEmail.trim()) return
    action(
      () => inviteOrganizationMember(selectedId, inviteEmail.trim(), 'member'),
      'Invitation sent'
    )
    setInviteEmail('')
  }
  const createToken = () => {
    if (!selectedId || !tokenName.trim()) return
    action(
      () => createOrganizationToken(selectedId, tokenName.trim()),
      'Organization key created'
    )
    setTokenName('')
  }
  const transfer = () => {
    if (!selectedId) return
    const amount = Number(transferAmount)
    const quota = parseQuotaFromDollars(amount)
    if (
      !Number.isFinite(amount) ||
      amount <= 0 ||
      !Number.isSafeInteger(quota) ||
      quota <= 0
    ) {
      return
    }
    action(
      () => transferToOrganization(selectedId, quota),
      'Organization balance updated'
    )
    setTransferAmount('')
  }
  const copyKey = async (tokenId: number) => {
    if (!selectedId) return
    try {
      const result = await revealOrganizationToken(selectedId, tokenId)
      if (!result?.key) return
      const key = result.key.startsWith('sk-') ? result.key : `sk-${result.key}`
      await navigator.clipboard.writeText(key)
      toast.success(t('Copied'))
    } catch (error) {
      toast.error(error instanceof Error ? error.message : t('Request failed'))
    }
  }

  return (
    <Main>
      <div className='min-h-0 flex-1 overflow-auto px-3 py-3 sm:px-4 sm:py-6'>
        <div className='mx-auto flex w-full max-w-7xl flex-col gap-4 sm:gap-6'>
          <div className='flex flex-wrap items-start justify-between gap-3'>
            <div>
              <h1 className='text-xl font-semibold'>{t('Organizations')}</h1>
              <p className='text-muted-foreground text-sm'>
                {t(
                  'Manage shared team workspaces, members, and organization API keys.'
                )}
              </p>
            </div>
            {activeId === null && (
              <div className='flex items-end gap-2'>
                <Field>
                  <FieldLabel htmlFor='organizationName'>
                    {t('New organization')}
                  </FieldLabel>
                  <Input
                    id='organizationName'
                    value={organizationName}
                    onChange={(event) =>
                      setOrganizationName(event.target.value)
                    }
                    placeholder={t('Organization name')}
                  />
                </Field>
                <Button onClick={create} disabled={!organizationName.trim()}>
                  <Plus data-icon='inline-start' />
                  {t('Create')}
                </Button>
              </div>
            )}
          </div>

          {activeId === null &&
            pendingInvitations.data &&
            pendingInvitations.data.length > 0 && (
              <Card>
                <CardHeader>
                  <CardTitle>{t('Pending invitations')}</CardTitle>
                  <CardDescription>
                    {t('Accept an invitation to join an organization.')}
                  </CardDescription>
                </CardHeader>
                <CardContent className='flex flex-col gap-2'>
                  {pendingInvitations.data.map((invitation) => (
                    <div
                      key={invitation.id}
                      className='flex flex-wrap items-center justify-between gap-3 rounded-lg border p-3'
                    >
                      <span>
                        {invitation.organization_name || invitation.email}
                      </span>
                      <Button
                        size='sm'
                        onClick={() =>
                          action(
                            () => acceptOrganizationInvitation(invitation.id),
                            'Invitation accepted'
                          )
                        }
                      >
                        {t('Accept')}
                      </Button>
                    </div>
                  ))}
                </CardContent>
              </Card>
            )}

          {activeId === null &&
            organizations.data &&
            organizations.data.length > 0 && (
              <div className='flex flex-wrap gap-2'>
                {organizations.data.map((entry) => (
                  <Button
                    key={entry.organization.id}
                    variant={
                      selectedId === entry.organization.id
                        ? 'default'
                        : 'outline'
                    }
                    onClick={() =>
                      userId &&
                      selectOrganization(userId, entry.organization.id)
                    }
                  >
                    <Building2 data-icon='inline-start' />
                    {entry.organization.name}
                  </Button>
                ))}
              </div>
            )}

          {!selectedId ? (
            <Card>
              <CardContent className='text-muted-foreground py-12 text-center text-sm'>
                {t('Create or join an organization to get started.')}
              </CardContent>
            </Card>
          ) : (
            <Tabs defaultValue='overview' className='flex flex-col gap-4'>
              <TabsList className='w-full justify-start overflow-x-auto'>
                <TabsTrigger value='overview'>{t('Overview')}</TabsTrigger>
                {canManage && (
                  <TabsTrigger value='members'>{t('Members')}</TabsTrigger>
                )}
                {canManage && (
                  <TabsTrigger value='wallet'>{t('Wallet')}</TabsTrigger>
                )}
                <TabsTrigger value='keys'>
                  {t('Organization API Keys')}
                </TabsTrigger>
                <TabsTrigger value='logs'>{t('Usage Logs')}</TabsTrigger>
              </TabsList>
              {organization.data?.membership.role !== 'owner' && (
                <div className='flex justify-end'>
                  <Button
                    variant='outline'
                    size='sm'
                    onClick={() =>
                      setConfirmation({
                        title: t('Leave organization'),
                        description: t(
                          'Your organization keys will be disabled. Historical usage stays with the organization.'
                        ),
                        run: () =>
                          action(async () => {
                            if (!selectedId) return
                            await leaveOrganization(selectedId)
                            if (userId) selectOrganization(userId, null)
                          }, 'Organization left'),
                      })
                    }
                  >
                    {t('Leave organization')}
                  </Button>
                </div>
              )}
              <TabsContent
                value='overview'
                className='grid gap-4 md:grid-cols-3'
              >
                {canManage && (
                  <Card>
                    <CardHeader>
                      <CardTitle>{t('Organization balance')}</CardTitle>
                    </CardHeader>
                    <CardContent className='text-2xl font-semibold'>
                      {formatQuota(organization.data?.organization.quota ?? 0)}
                    </CardContent>
                  </Card>
                )}
                <Card>
                  <CardHeader>
                    <CardTitle>{t('This month')}</CardTitle>
                  </CardHeader>
                  <CardContent className='text-2xl font-semibold'>
                    {formatQuota(
                      usage.data?.members?.reduce(
                        (total, member) => total + member.used_quota,
                        0
                      ) ??
                        usage.data?.monthly_used_quota ??
                        organization.data?.monthly_used_quota ??
                        0
                    )}
                  </CardContent>
                </Card>
                {canManage && (
                  <Card>
                    <CardHeader>
                      <CardTitle>{t('Requests')}</CardTitle>
                    </CardHeader>
                    <CardContent className='text-2xl font-semibold'>
                      {usage.data?.request_count ??
                        organization.data?.organization.request_count ??
                        0}
                    </CardContent>
                  </Card>
                )}
                {!canManage && (
                  <Card>
                    <CardHeader>
                      <CardTitle>{t('Monthly limit')}</CardTitle>
                    </CardHeader>
                    <CardContent>
                      {organization.data?.membership.monthly_quota_limit
                        ? formatQuota(
                            organization.data.membership.monthly_quota_limit
                          )
                        : t('Unlimited')}
                    </CardContent>
                  </Card>
                )}
                {canManage && (
                  <Card className='md:col-span-3'>
                    <CardHeader>
                      <CardTitle>{t('Invite members')}</CardTitle>
                      <CardDescription>
                        {t(
                          'Members share the organization balance according to their role and monthly limit.'
                        )}
                      </CardDescription>
                    </CardHeader>
                    <CardContent>
                      <FieldGroup className='sm:flex-row sm:items-end'>
                        <Field>
                          <FieldLabel htmlFor='inviteEmail'>
                            {t('Member email')}
                          </FieldLabel>
                          <Input
                            id='inviteEmail'
                            type='email'
                            value={inviteEmail}
                            onChange={(event) =>
                              setInviteEmail(event.target.value)
                            }
                          />
                        </Field>
                        <Button onClick={invite} disabled={!inviteEmail.trim()}>
                          <MailPlus data-icon='inline-start' />
                          {t('Send invitation')}
                        </Button>
                      </FieldGroup>
                    </CardContent>
                  </Card>
                )}
              </TabsContent>
              {canManage && (
                <TabsContent value='members'>
                  <Card>
                    <CardHeader>
                      <CardTitle>{t('Members')}</CardTitle>
                    </CardHeader>
                    <CardContent className='flex flex-col gap-2'>
                      {(members.data ?? []).map((member) => (
                        <div
                          key={member.user_id}
                          className='flex flex-wrap items-center justify-between gap-3 border-b py-3 last:border-0'
                        >
                          <div>
                            <div className='font-medium'>
                              {member.display_name || member.username}
                            </div>
                            <div className='text-muted-foreground text-xs'>
                              {member.email}
                            </div>
                          </div>
                          <div className='flex items-center gap-2'>
                            <Badge variant='outline'>{t(member.role)}</Badge>
                            <span className='text-muted-foreground text-xs'>
                              {formatQuota(member.monthly_used_quota)}
                            </span>
                            {member.role !== 'owner' &&
                              (organization.data?.membership.role === 'owner' ||
                                member.role === 'member') && (
                                <MemberControls
                                  key={`${member.user_id}-${member.monthly_quota_limit}-${member.role}`}
                                  member={member}
                                  owner={
                                    organization.data?.membership.role ===
                                    'owner'
                                  }
                                  onSave={(data) =>
                                    selectedId &&
                                    action(
                                      () =>
                                        updateOrganizationMember(
                                          selectedId,
                                          member.user_id,
                                          data
                                        ),
                                      'Member updated'
                                    )
                                  }
                                  onRemove={() =>
                                    setConfirmation({
                                      title: t('Remove member'),
                                      description: t(
                                        'The member will lose organization access and their organization keys will be disabled.'
                                      ),
                                      run: () =>
                                        selectedId &&
                                        action(
                                          () =>
                                            removeOrganizationMember(
                                              selectedId,
                                              member.user_id
                                            ),
                                          'Member removed'
                                        ),
                                    })
                                  }
                                  onTransfer={
                                    organization.data?.membership.role ===
                                    'owner'
                                      ? () =>
                                          setConfirmation({
                                            title: t('Transfer ownership'),
                                            description: t(
                                              'This member will become the owner. Your role will change to administrator.'
                                            ),
                                            run: () =>
                                              selectedId &&
                                              action(
                                                () =>
                                                  transferOrganizationOwnership(
                                                    selectedId,
                                                    member.user_id
                                                  ),
                                                'Ownership transferred'
                                              ),
                                          })
                                      : undefined
                                  }
                                />
                              )}
                          </div>
                        </div>
                      ))}
                      {(invitations.data ?? [])
                        .filter((invitation) => invitation.status === 'pending')
                        .map((invitation) => (
                          <div
                            key={`invitation-${invitation.id}`}
                            className='text-muted-foreground flex items-center gap-2 border-b py-3 text-sm last:border-0'
                          >
                            <MailPlus className='size-4' />
                            {invitation.email}
                            <Badge variant='outline'>{t('Pending')}</Badge>
                            <Button
                              variant='ghost'
                              size='sm'
                              onClick={() =>
                                selectedId &&
                                action(
                                  () =>
                                    revokeOrganizationInvitation(
                                      selectedId,
                                      invitation.id
                                    ),
                                  'Invitation revoked'
                                )
                              }
                            >
                              {t('Revoke')}
                            </Button>
                          </div>
                        ))}
                    </CardContent>
                  </Card>
                </TabsContent>
              )}
              {canManage && (
                <TabsContent value='wallet'>
                  <Card>
                    <CardHeader>
                      <CardTitle>{t('Organization wallet')}</CardTitle>
                      <CardDescription>
                        {t(
                          'Transfer from your personal balance to fund this organization.'
                        )}
                      </CardDescription>
                    </CardHeader>
                    <CardContent className='flex flex-col gap-4'>
                      <Button render={<Link to='/wallet' />}>
                        {t('Add funds')}
                      </Button>
                      <FieldGroup className='sm:flex-row sm:items-end'>
                        <Field>
                          <FieldLabel htmlFor='transferAmount'>
                            {t('Transfer amount')}
                          </FieldLabel>
                          <Input
                            id='transferAmount'
                            type='number'
                            min='0'
                            step='any'
                            value={transferAmount}
                            onChange={(event) =>
                              setTransferAmount(event.target.value)
                            }
                          />
                        </Field>
                        <Button
                          onClick={transfer}
                          disabled={!(Number(transferAmount) > 0)}
                        >
                          {t('Transfer to organization')}
                        </Button>
                      </FieldGroup>
                      {(transactions.data ?? []).map((transaction) => (
                        <div
                          key={transaction.id}
                          className='flex justify-between gap-3 border-b py-2 text-sm last:border-0'
                        >
                          <div className='min-w-0 flex-1'>
                            <div className='flex flex-wrap gap-x-3 gap-y-1'>
                              <span>
                                {new Date(
                                  transaction.created_at * 1000
                                ).toLocaleString()}
                              </span>
                              <span>
                                {transaction.user_name || t('System')}
                              </span>
                              <span>
                                {(
                                  {
                                    transfer_in: t('Transfer in'),
                                    topup: t('Add funds'),
                                  } as Record<string, string>
                                )[transaction.type] ?? transaction.type}
                              </span>
                            </div>
                            {transaction.description && (
                              <div className='text-muted-foreground truncate text-xs'>
                                {transaction.description ===
                                'Personal wallet transfer'
                                  ? t('Transfer from personal balance')
                                  : transaction.description}
                              </div>
                            )}
                          </div>
                          <div className='text-right'>
                            <div>{formatQuota(transaction.amount)}</div>
                            <div className='text-muted-foreground text-xs'>
                              {t('Balance')}:{' '}
                              {formatQuota(transaction.balance_after)}
                            </div>
                          </div>
                        </div>
                      ))}
                    </CardContent>
                  </Card>
                </TabsContent>
              )}
              <TabsContent value='keys'>
                <Card>
                  <CardHeader>
                    <CardTitle>{t('Organization API Keys')}</CardTitle>
                    <CardDescription>
                      {t('Keys created here charge the organization balance.')}
                    </CardDescription>
                  </CardHeader>
                  <CardContent className='flex flex-col gap-3'>
                    <FieldGroup className='sm:flex-row sm:items-end'>
                      <Field>
                        <FieldLabel htmlFor='tokenName'>
                          {t('Key name')}
                        </FieldLabel>
                        <Input
                          id='tokenName'
                          value={tokenName}
                          onChange={(event) => setTokenName(event.target.value)}
                        />
                      </Field>
                      <Button
                        onClick={createToken}
                        disabled={!tokenName.trim()}
                      >
                        <KeyRound data-icon='inline-start' />
                        {t('Create key')}
                      </Button>
                    </FieldGroup>
                    {(tokens.data ?? []).map((token) => (
                      <div
                        key={token.id}
                        className='flex flex-wrap items-center justify-between gap-3 rounded-lg border p-3'
                      >
                        <div>
                          <div className='font-medium'>{token.name}</div>
                          <div className='mb-1 flex flex-wrap gap-1'>
                            <Badge variant='secondary'>
                              {t('Organization key')}
                            </Badge>
                            <Badge
                              variant={
                                token.status === 1 ? 'default' : 'outline'
                              }
                            >
                              {token.status === 1
                                ? t('Enabled')
                                : t('Disabled')}
                            </Badge>
                          </div>
                          <div className='text-muted-foreground font-mono text-xs'>
                            {token.key}
                          </div>
                          <div className='text-muted-foreground text-xs'>
                            {t('Created by')}:{' '}
                            {token.creator_name || t('Unknown')}
                          </div>
                          <div className='text-muted-foreground text-xs'>
                            {organization.data?.organization.name} ·{' '}
                            {t('Last used')}:{' '}
                            {token.accessed_time > 0
                              ? new Date(
                                  token.accessed_time * 1000
                                ).toLocaleString()
                              : t('Never used')}
                          </div>
                          <div className='text-muted-foreground text-xs'>
                            {t('Created at')}:{' '}
                            {new Date(
                              token.created_time * 1000
                            ).toLocaleString()}{' '}
                            ·{' '}
                            {token.expired_time <= 0
                              ? t('Never expires')
                              : new Date(
                                  token.expired_time * 1000
                                ).toLocaleString()}
                          </div>
                        </div>
                        <Button
                          variant='ghost'
                          size='icon'
                          aria-label={t('Copy')}
                          onClick={() => void copyKey(token.id)}
                        >
                          <Copy />
                        </Button>
                        <Button
                          variant='outline'
                          size='sm'
                          onClick={() =>
                            selectedId &&
                            action(
                              () =>
                                updateOrganizationToken(
                                  selectedId,
                                  token.id,
                                  token.status === 1 ? 2 : 1
                                ),
                              'Key updated'
                            )
                          }
                        >
                          {token.status === 1 ? t('Disable') : t('Enable')}
                        </Button>
                        <Button
                          variant='ghost'
                          size='sm'
                          onClick={() =>
                            setConfirmation({
                              title: t('Delete key'),
                              description: t(
                                'This key will stop working immediately.'
                              ),
                              run: () =>
                                selectedId &&
                                action(
                                  () =>
                                    deleteOrganizationToken(
                                      selectedId,
                                      token.id
                                    ),
                                  'Key deleted'
                                ),
                            })
                          }
                        >
                          {t('Delete')}
                        </Button>
                      </div>
                    ))}
                  </CardContent>
                </Card>
              </TabsContent>
              <TabsContent value='logs'>
                <Card>
                  <CardHeader>
                    <CardTitle>{t('Usage Logs')}</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead>{t('Time')}</TableHead>
                          <TableHead>{t('Member')}</TableHead>
                          <TableHead>{t('Model')}</TableHead>
                          <TableHead>{t('API Key')}</TableHead>
                          <TableHead>{t('Tokens')}</TableHead>
                          <TableHead>{t('Cost')}</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {(logs.data?.items ?? []).map((log) => (
                          <TableRow key={log.id}>
                            <TableCell>
                              {new Date(log.created_at * 1000).toLocaleString()}
                            </TableCell>
                            <TableCell>{log.username}</TableCell>
                            <TableCell>{log.model_name}</TableCell>
                            <TableCell>{log.token_name}</TableCell>
                            <TableCell>
                              {log.prompt_tokens + log.completion_tokens}
                            </TableCell>
                            <TableCell>{formatQuota(log.quota)}</TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                    <div className='mt-3 flex items-center justify-end gap-2'>
                      <Button
                        variant='outline'
                        size='sm'
                        disabled={logPage <= 1}
                        onClick={() => setLogPage((page) => page - 1)}
                      >
                        {t('Previous')}
                      </Button>
                      <span className='text-xs'>{logPage}</span>
                      <Button
                        variant='outline'
                        size='sm'
                        disabled={logPage * 20 >= (logs.data?.total ?? 0)}
                        onClick={() => setLogPage((page) => page + 1)}
                      >
                        {t('Next')}
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              </TabsContent>
            </Tabs>
          )}
        </div>
      </div>
      <AlertDialog
        open={confirmation !== null}
        onOpenChange={(open) => !open && setConfirmation(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{confirmation?.title}</AlertDialogTitle>
            <AlertDialogDescription>
              {confirmation?.description}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t('Cancel')}</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                confirmation?.run()
                setConfirmation(null)
              }}
            >
              {t('Confirm')}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </Main>
  )
}

function MemberControls(props: {
  member: OrganizationMember
  owner: boolean
  onSave: (data: {
    role: 'admin' | 'member'
    monthly_quota_limit: number
  }) => void
  onRemove: () => void
  onTransfer?: () => void
}) {
  const { t } = useTranslation()
  const [role, setRole] = useState<'admin' | 'member'>(
    props.member.role === 'admin' ? 'admin' : 'member'
  )
  const [limit, setLimit] = useState(
    String(quotaUnitsToDollars(props.member.monthly_quota_limit))
  )
  return (
    <div className='flex flex-wrap items-center gap-2'>
      {props.owner && (
        <NativeSelect
          aria-label={t('Role')}
          value={role}
          onChange={(event) =>
            setRole(event.target.value as 'admin' | 'member')
          }
        >
          <NativeSelectOption value='member'>{t('member')}</NativeSelectOption>
          <NativeSelectOption value='admin'>{t('admin')}</NativeSelectOption>
        </NativeSelect>
      )}
      <Input
        type='number'
        min='0'
        className='w-28'
        aria-label={t('Monthly limit')}
        value={limit}
        onChange={(event) => setLimit(event.target.value)}
      />
      <Button
        variant='outline'
        size='sm'
        onClick={() => {
          const quota = parseQuotaFromDollars(Number(limit))
          if (Number.isSafeInteger(quota) && quota >= 0) {
            props.onSave({ role, monthly_quota_limit: quota })
          }
        }}
      >
        {t('Save')}
      </Button>
      <Button variant='ghost' size='sm' onClick={props.onRemove}>
        {t('Remove')}
      </Button>
      {props.onTransfer && (
        <Button variant='ghost' size='sm' onClick={props.onTransfer}>
          {t('Transfer ownership')}
        </Button>
      )}
    </div>
  )
}
