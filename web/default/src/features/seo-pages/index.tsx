import {
  ArrowRight,
  Check,
  CircleDollarSign,
  Code2,
  KeyRound,
  Layers3,
  Route,
} from 'lucide-react'
import type { ReactNode } from 'react'

import { CopyButton } from '@/components/copy-button'
import { PublicLayout } from '@/components/layout'
import { Footer } from '@/components/layout/components/footer'
import { Button } from '@/components/ui/button'
import type { PublicBillingInfo } from '@/features/auth/types'
import { useStatus } from '@/hooks/use-status'

const OPENAI_PYTHON_EXAMPLE = `from openai import OpenAI

client = OpenAI(
    base_url="__API_ORIGIN__/v1",
    api_key="YOUR_API_KEY",
)

completion = client.chat.completions.create(
    model="YOUR_ENABLED_MODEL",
    messages=[
        {"role": "user", "content": "Summarize this text in one sentence."}
    ],
)

print(completion.choices[0].message.content)`

const MULTI_PROVIDER_CURL_EXAMPLE = `curl __API_ORIGIN__/v1/chat/completions \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "YOUR_ENABLED_MODEL",
    "messages": [{"role": "user", "content": "Hello"}]
  }'`

function getApiBaseUrl() {
  return typeof window === 'undefined' ? '/v1' : `${window.location.origin}/v1`
}

function withSiteOrigin(example: string) {
  const origin = typeof window === 'undefined' ? '' : window.location.origin
  return example.replaceAll('__API_ORIGIN__', origin)
}

function formatMinimumTopup(billing: PublicBillingInfo) {
  const amount = new Intl.NumberFormat().format(billing.minimum_topup ?? 0)
  if (billing.currency_display_type === 'TOKENS') return `${amount} tokens`
  return `${billing.currency_symbol ?? ''}${amount}`
}

function getBillingFacts(billing?: PublicBillingInfo) {
  if (!billing) return []

  const facts: string[] = []
  if (billing.registration_credit_enabled !== undefined) {
    facts.push(
      billing.registration_credit_enabled
        ? 'Registration includes default initial credit'
        : 'No default registration credit'
    )
  }
  if (billing.minimum_topup_enabled && (billing.minimum_topup ?? 0) > 0) {
    facts.push(`Minimum top-up: ${formatMinimumTopup(billing)}`)
  }
  return facts
}

function getBillingDetails(billing?: PublicBillingInfo) {
  const facts = getBillingFacts(billing)
  if (facts.length === 0) {
    return "Review the account page for current registration and recharge terms. Usage is charged according to each model's current price."
  }
  return `${facts.join('. ')}. Usage is charged according to each model’s current price.`
}

function PageFrame({ children }: { children: ReactNode }) {
  return (
    <PublicLayout showMainContainer={false}>
      {children}
      <Footer />
    </PublicLayout>
  )
}

function CodePanel({ label, code }: { label: string; code: string }) {
  return (
    <div className='overflow-hidden rounded-lg border border-zinc-800 bg-zinc-950 text-zinc-100 shadow-sm'>
      <div className='flex min-h-11 items-center justify-between gap-3 border-b border-zinc-800 px-3 sm:px-4'>
        <div className='flex min-w-0 items-center gap-2 text-xs text-zinc-300'>
          <span className='size-2 shrink-0 rounded-full bg-emerald-400' />
          <span className='truncate'>{label}</span>
        </div>
        <CopyButton
          value={code}
          className='text-zinc-300 hover:bg-zinc-800 hover:text-white'
          tooltip='Copy code example'
          successTooltip='Code copied'
          aria-label='Copy code example'
        />
      </div>
      <pre className='max-h-[28rem] overflow-x-auto p-4 text-xs leading-6 sm:text-sm'>
        <code>{code}</code>
      </pre>
    </div>
  )
}

function ActionLink({
  href,
  children,
  secondary = false,
}: {
  href: string
  children: ReactNode
  secondary?: boolean
}) {
  return (
    <Button
      variant={secondary ? 'outline' : 'default'}
      render={<a href={href} />}
    >
      {children}
      <ArrowRight data-icon='inline-end' />
    </Button>
  )
}

function Eyebrow({ children }: { children: ReactNode }) {
  return (
    <p className='text-muted-foreground mb-3 flex items-center gap-2 text-xs font-semibold uppercase'>
      <span className='h-px w-6 bg-emerald-600 dark:bg-emerald-400' />
      {children}
    </p>
  )
}

function PageHero({
  eyebrow,
  title,
  description,
  codeLabel,
  code,
  billing,
}: {
  eyebrow: string
  title: string
  description: string
  codeLabel: string
  code: string
  billing?: PublicBillingInfo
}) {
  const billingFacts = getBillingFacts(billing)

  return (
    <section className='border-border/70 border-b'>
      <div className='mx-auto grid w-full max-w-6xl items-center gap-9 px-4 pt-24 pb-12 sm:px-6 sm:pt-28 sm:pb-16 lg:grid-cols-[1fr_1.05fr] lg:gap-14'>
        <div className='max-w-2xl'>
          <Eyebrow>{eyebrow}</Eyebrow>
          <h1 className='max-w-[18ch] text-4xl leading-tight font-semibold sm:text-5xl'>
            {title}
          </h1>
          <p className='text-muted-foreground mt-5 max-w-xl text-base leading-7 sm:text-lg'>
            {description}
          </p>
          <div className='mt-7 flex flex-wrap items-center gap-3'>
            <ActionLink href='/sign-up'>Create an account</ActionLink>
            <ActionLink href='/pricing' secondary>
              Browse model pricing
            </ActionLink>
          </div>
          {billingFacts.length > 0 && (
            <p className='text-muted-foreground mt-5 flex flex-wrap items-center gap-x-2 gap-y-1 text-sm'>
              <CircleDollarSign className='size-4 shrink-0 text-emerald-600 dark:text-emerald-400' />
              {billingFacts.map((fact, index) => (
                <span key={fact} className='inline-flex items-center gap-2'>
                  {index > 0 && <span aria-hidden='true'>·</span>}
                  <span>{fact}</span>
                </span>
              ))}
            </p>
          )}
        </div>
        <CodePanel label={codeLabel} code={code} />
      </div>
    </section>
  )
}

function Steps({ items }: { items: { title: string; description: string }[] }) {
  return (
    <div className='grid gap-0 md:grid-cols-3'>
      {items.map((item, index) => (
        <article
          key={item.title}
          className='border-border/70 border-t py-5 md:px-5 md:not-first:border-l md:first:pl-0 md:last:pr-0'
        >
          <p className='text-muted-foreground mb-3 text-xs font-medium'>
            0{index + 1}
          </p>
          <h3 className='text-base font-semibold'>{item.title}</h3>
          <p className='text-muted-foreground mt-2 max-w-sm text-sm leading-6'>
            {item.description}
          </p>
        </article>
      ))}
    </div>
  )
}

function Faq({ items }: { items: { question: string; answer: string }[] }) {
  return (
    <div className='divide-border/70 divide-y border-y'>
      {items.map((item) => (
        <details key={item.question} className='group py-4'>
          <summary className='flex cursor-pointer list-none items-center justify-between gap-4 text-sm font-medium [&::-webkit-details-marker]:hidden'>
            {item.question}
            <span className='text-muted-foreground text-lg leading-none transition-transform group-open:rotate-45'>
              +
            </span>
          </summary>
          <p className='text-muted-foreground max-w-3xl pt-3 text-sm leading-6'>
            {item.answer}
          </p>
        </details>
      ))}
    </div>
  )
}

function PageCta({
  title,
  description,
}: {
  title: string
  description: string
}) {
  return (
    <section className='bg-muted/40 border-border/70 border-y'>
      <div className='mx-auto flex max-w-6xl flex-col gap-5 px-4 py-9 sm:px-6 sm:py-11 md:flex-row md:items-center md:justify-between'>
        <div className='max-w-2xl'>
          <h2 className='text-xl font-semibold'>{title}</h2>
          <p className='text-muted-foreground mt-2 text-sm leading-6'>
            {description}
          </p>
        </div>
        <div className='flex shrink-0 flex-wrap gap-3'>
          <ActionLink href='/pricing' secondary>
            Check pricing
          </ActionLink>
          <ActionLink href='/sign-up'>Get started</ActionLink>
        </div>
      </div>
    </section>
  )
}

export function OpenAICompatibleApiPage() {
  const { status } = useStatus()
  const billing = status?.public_billing
  const apiBaseUrl = getApiBaseUrl()
  const systemName =
    typeof status?.system_name === 'string'
      ? status.system_name
      : 'this service'

  return (
    <PageFrame>
      <main>
        <PageHero
          eyebrow='OpenAI-compatible API'
          title='Use the OpenAI SDK with multiple AI models'
          description='Keep the OpenAI client and chat-completions workflow you already know. Set a compatible base URL, use your API key, and choose an enabled model from the live catalog.'
          codeLabel='Python · OpenAI SDK'
          code={withSiteOrigin(OPENAI_PYTHON_EXAMPLE)}
          billing={billing}
        />

        <section className='mx-auto max-w-6xl px-4 py-11 sm:px-6 sm:py-14'>
          <div className='mb-7 max-w-2xl'>
            <Eyebrow>Quick start</Eyebrow>
            <h2 className='text-2xl font-semibold sm:text-3xl'>
              Three values connect your app
            </h2>
            <p className='text-muted-foreground mt-3 text-sm leading-6'>
              The SDK request shape stays familiar. The endpoint and model ID
              tell the gateway where to send the request.
            </p>
          </div>
          <Steps
            items={[
              {
                title: 'Create an API key',
                description:
                  'Create a key in your account and keep it in a server-side environment variable.',
              },
              {
                title: 'Set the base URL',
                description: `Use ${apiBaseUrl} as the base URL for OpenAI-compatible endpoints.`,
              },
              {
                title: 'Choose an enabled model',
                description:
                  'Copy the exact model ID from the current catalog and confirm it supports the endpoint you need.',
              },
            ]}
          />
        </section>

        <section className='border-border/70 border-y'>
          <div className='mx-auto grid max-w-6xl gap-8 px-4 py-11 sm:px-6 sm:py-14 md:grid-cols-[0.85fr_1.15fr] md:gap-14'>
            <div>
              <Eyebrow>Compatibility</Eyebrow>
              <h2 className='text-2xl font-semibold sm:text-3xl'>
                OpenAI-compatible does not mean every model is identical
              </h2>
              <p className='text-muted-foreground mt-4 text-sm leading-6'>
                The gateway accepts supported OpenAI-style requests, while model
                capabilities and optional parameters still depend on the
                selected model and upstream. Check the model page before relying
                on a feature in production.
              </p>
            </div>
            <div className='divide-border/70 divide-y border-y'>
              <div className='flex gap-4 py-4'>
                <Code2 className='mt-0.5 size-4 shrink-0 text-emerald-600 dark:text-emerald-400' />
                <div>
                  <h3 className='text-sm font-semibold'>Endpoint</h3>
                  <p className='text-muted-foreground mt-1 text-sm leading-6'>
                    Chat requests use <code>POST /v1/chat/completions</code>.
                    Other API formats and endpoints are listed in the API
                    documentation.
                  </p>
                </div>
              </div>
              <div className='flex gap-4 py-4'>
                <Layers3 className='mt-0.5 size-4 shrink-0 text-sky-600 dark:text-sky-400' />
                <div>
                  <h3 className='text-sm font-semibold'>Model ID</h3>
                  <p className='text-muted-foreground mt-1 text-sm leading-6'>
                    Model availability can change. Use the exact ID shown in the
                    current model catalog rather than assuming a provider's
                    public model name is enabled.
                  </p>
                </div>
              </div>
              <div className='flex gap-4 py-4'>
                <KeyRound className='mt-0.5 size-4 shrink-0 text-amber-600 dark:text-amber-400' />
                <div>
                  <h3 className='text-sm font-semibold'>Authentication</h3>
                  <p className='text-muted-foreground mt-1 text-sm leading-6'>
                    Send your account API key as a bearer token. Do not expose
                    keys in browser code or public repositories.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section className='mx-auto max-w-6xl px-4 py-11 sm:px-6 sm:py-14'>
          <div className='mb-6 max-w-2xl'>
            <Eyebrow>Before your first request</Eyebrow>
            <h2 className='text-2xl font-semibold sm:text-3xl'>
              Know the pricing and account requirements
            </h2>
          </div>
          <div className='grid gap-6 md:grid-cols-2'>
            <div className='border-border/70 border-t pt-4'>
              <CircleDollarSign className='mb-3 size-5 text-emerald-600 dark:text-emerald-400' />
              <h3 className='font-semibold'>Account and recharge details</h3>
              <p className='text-muted-foreground mt-2 text-sm leading-6'>
                {getBillingDetails(billing)}
              </p>
            </div>
            <div className='border-border/70 border-t pt-4'>
              <Check className='mb-3 size-5 text-sky-600 dark:text-sky-400' />
              <h3 className='font-semibold'>Confirm the live model listing</h3>
              <p className='text-muted-foreground mt-2 text-sm leading-6'>
                Review model pricing, supported endpoints, and any listed limits
                before switching production traffic.
              </p>
            </div>
          </div>
          <div className='mt-8 flex flex-wrap gap-x-6 gap-y-2 text-sm'>
            <a
              className='text-primary underline-offset-4 hover:underline'
              href='/pricing'
            >
              Compare model pricing
            </a>
            <a
              className='text-primary underline-offset-4 hover:underline'
              href='/multi-provider-ai-api'
            >
              Read about multi-provider API access
            </a>
          </div>
        </section>

        <section className='mx-auto max-w-6xl px-4 pb-11 sm:px-6 sm:pb-14'>
          <div className='mb-6 max-w-2xl'>
            <Eyebrow>Common questions</Eyebrow>
            <h2 className='text-2xl font-semibold sm:text-3xl'>
              OpenAI-compatible API FAQ
            </h2>
          </div>
          <Faq
            items={[
              {
                question: 'Can I keep using the OpenAI Python SDK?',
                answer: `Yes. Set base_url to ${apiBaseUrl}, provide your ${systemName} API key, and use a model ID currently listed in the catalog.`,
              },
              {
                question:
                  'Does every model support the same OpenAI parameters?',
                answer:
                  'No. The compatible request format is shared, but model capabilities and optional parameters depend on the selected model and upstream. Check its current model details before production use.',
              },
            ]}
          />
        </section>

        <PageCta
          title='Ready to test an OpenAI-compatible connection?'
          description='Create an account, review available models, then make a small request with your own API key and model ID.'
        />
      </main>
    </PageFrame>
  )
}

export function MultiProviderAiApiPage() {
  const { status } = useStatus()
  const billing = status?.public_billing

  return (
    <PageFrame>
      <main>
        <PageHero
          eyebrow='Multi-provider AI API'
          title='One API gateway for multiple AI providers'
          description='Access supported models through one account and API key. Choose the compatible API format and exact model ID shown in the current catalog; endpoint support is model-specific.'
          codeLabel='cURL · OpenAI-compatible chat'
          code={withSiteOrigin(MULTI_PROVIDER_CURL_EXAMPLE)}
          billing={billing}
        />

        <section className='mx-auto max-w-6xl px-4 py-11 sm:px-6 sm:py-14'>
          <div className='mb-7 max-w-2xl'>
            <Eyebrow>Compatible API formats</Eyebrow>
            <h2 className='text-2xl font-semibold sm:text-3xl'>
              Match the request format to the model
            </h2>
            <p className='text-muted-foreground mt-3 text-sm leading-6'>
              A shared gateway does not make provider protocols interchangeable.
              Use the endpoint type listed for the model you intend to call.
            </p>
          </div>

          <div className='overflow-x-auto'>
            <table className='w-full min-w-[560px] border-collapse text-left text-sm'>
              <thead>
                <tr className='border-border/70 text-muted-foreground border-y'>
                  <th className='py-3 pr-5 font-medium'>API format</th>
                  <th className='py-3 pr-5 font-medium'>Example endpoint</th>
                  <th className='py-3 font-medium'>Use when</th>
                </tr>
              </thead>
              <tbody>
                <tr className='border-border/70 border-b'>
                  <th className='py-4 pr-5 font-medium'>OpenAI-compatible</th>
                  <td className='py-4 pr-5 font-mono text-xs'>
                    /v1/chat/completions
                  </td>
                  <td className='text-muted-foreground py-4'>
                    Your client uses the OpenAI SDK or chat-completions format.
                  </td>
                </tr>
                <tr className='border-border/70 border-b'>
                  <th className='py-4 pr-5 font-medium'>
                    Anthropic-compatible
                  </th>
                  <td className='py-4 pr-5 font-mono text-xs'>/v1/messages</td>
                  <td className='text-muted-foreground py-4'>
                    The selected model lists the Anthropic Messages endpoint.
                  </td>
                </tr>
                <tr className='border-border/70 border-b'>
                  <th className='py-4 pr-5 font-medium'>Gemini-compatible</th>
                  <td className='py-4 pr-5 font-mono text-xs'>
                    /v1beta/models/&#123;model&#125;:generateContent
                  </td>
                  <td className='text-muted-foreground py-4'>
                    The selected model lists the Gemini generate-content format.
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <section className='border-border/70 border-y'>
          <div className='mx-auto grid max-w-6xl gap-10 px-4 py-11 sm:px-6 sm:py-14 md:grid-cols-[0.85fr_1.15fr] md:gap-14'>
            <div>
              <Eyebrow>Request routing</Eyebrow>
              <h2 className='text-2xl font-semibold sm:text-3xl'>
                Select a model explicitly
              </h2>
              <p className='text-muted-foreground mt-4 text-sm leading-6'>
                Use the model ID from the catalog in your request. The gateway
                routes that request according to the configured model and
                channel. Availability, compatible formats, and features can
                differ between models.
              </p>
            </div>
            <div className='space-y-5'>
              <div className='border-border/70 flex gap-4 border-t pt-4'>
                <Route className='mt-0.5 size-4 shrink-0 text-emerald-600 dark:text-emerald-400' />
                <div>
                  <h3 className='text-sm font-semibold'>One account key</h3>
                  <p className='text-muted-foreground mt-1 text-sm leading-6'>
                    Authenticate requests with an API key from your account and
                    keep that key on your server.
                  </p>
                </div>
              </div>
              <div className='border-border/70 flex gap-4 border-t pt-4'>
                <Layers3 className='mt-0.5 size-4 shrink-0 text-sky-600 dark:text-sky-400' />
                <div>
                  <h3 className='text-sm font-semibold'>
                    Current model catalog
                  </h3>
                  <p className='text-muted-foreground mt-1 text-sm leading-6'>
                    Use the live catalog to check model names, supported
                    endpoints, and current prices before you integrate.
                  </p>
                </div>
              </div>
              <div className='border-border/70 flex gap-4 border-t pt-4'>
                <Code2 className='mt-0.5 size-4 shrink-0 text-amber-600 dark:text-amber-400' />
                <div>
                  <h3 className='text-sm font-semibold'>
                    Protocol-specific clients
                  </h3>
                  <p className='text-muted-foreground mt-1 text-sm leading-6'>
                    Keep the request format expected by each API. Do not send
                    Anthropic or Gemini payloads to an OpenAI endpoint unless
                    the model documentation explicitly supports that path.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section className='mx-auto max-w-6xl px-4 py-11 sm:px-6 sm:py-14'>
          <div className='mb-6 max-w-2xl'>
            <Eyebrow>Account and pricing</Eyebrow>
            <h2 className='text-2xl font-semibold sm:text-3xl'>
              See the cost before you route production traffic
            </h2>
            <p className='text-muted-foreground mt-3 text-sm leading-6'>
              {getBillingDetails(billing)} Compare prices and supported
              endpoints in the catalog before choosing a model.
            </p>
          </div>
          <div className='flex flex-wrap items-center gap-x-6 gap-y-2 text-sm'>
            <a
              className='text-primary underline-offset-4 hover:underline'
              href='/pricing'
            >
              Open model pricing
            </a>
            <a
              className='text-primary underline-offset-4 hover:underline'
              href='/openai-compatible-api'
            >
              Use an OpenAI-compatible SDK
            </a>
          </div>
        </section>

        <section className='mx-auto max-w-6xl px-4 pb-11 sm:px-6 sm:pb-14'>
          <div className='mb-6 max-w-2xl'>
            <Eyebrow>Common questions</Eyebrow>
            <h2 className='text-2xl font-semibold sm:text-3xl'>
              Multi-provider API FAQ
            </h2>
          </div>
          <Faq
            items={[
              {
                question: 'Is one request format supported by every provider?',
                answer:
                  'No. Use the compatible endpoint type listed for the model. OpenAI, Anthropic, and Gemini formats have different request and response shapes.',
              },
              {
                question: 'How do I compare models before choosing one?',
                answer:
                  'Open the model pricing page to compare current usage prices, endpoint support, and the exact model ID. Availability can change, so verify it before switching traffic.',
              },
            ]}
          />
        </section>

        <PageCta
          title='Choose a model and start with one request'
          description='Check its endpoint support and current price, create an API key, and test against the model ID listed in the catalog.'
        />
      </main>
    </PageFrame>
  )
}
