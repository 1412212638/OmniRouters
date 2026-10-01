import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { PublicHeader } from '@/components/layout'
import { Skeleton } from '@/components/ui/skeleton'
import { useSystemConfig } from '@/hooks/use-system-config'
import { cn } from '@/lib/utils'

import { AuthMascot } from './components/auth-mascot'

type AuthLayoutProps = {
  children: React.ReactNode
  variant?: 'default' | 'modern' | 'sign-in'
}

const loginHeroImage = 'https://cos.frostai.cn/omnirouters/login_en.webp'

const authClassicStyles = `
.auth-classic-shell {
  --auth-bg: radial-gradient(circle at 18% 16%, rgba(255,255,255,.78) 0, rgba(255,255,255,0) 28%), linear-gradient(90deg, #e7e3ef 0%, #e7e3ef 50%, #ffffff 50%, #ffffff 100%);
  --auth-left-aura: radial-gradient(circle at 20% 18%, rgba(255,255,255,.5) 0, rgba(255,255,255,0) 26%), radial-gradient(circle at 34% 78%, rgba(255,255,255,.24) 0, rgba(255,255,255,0) 24%);
  --auth-panel-bg: rgba(255,255,255,.82);
  --auth-panel-border: rgba(255,255,255,.74);
  --auth-panel-shadow: 0 24px 70px rgba(40,36,60,.12), inset 0 1px 0 rgba(255,255,255,.82);
  --auth-brand: #161827;
  --auth-title: #111326;
  --auth-subtitle: #767386;
  --auth-label: #27243a;
  --auth-link: #5f35cf;
  --auth-link-hover: #4521a5;
  --auth-primary-bg: #111111;
  --auth-primary-text: #ffffff;
  --auth-input-bg: rgba(255,255,255,.88);
  --auth-input-border: #e7e4ee;
  --auth-input-focus: rgba(95,53,207,.28);
  --auth-floor-shadow: linear-gradient(90deg, transparent 0%, rgba(35,31,42,.12) 18%, rgba(35,31,42,.16) 55%, transparent 100%);
  --auth-purple-shadow: 0 18px 45px rgba(83,49,217,.22);
  --auth-black-shadow: 0 16px 36px rgba(18,18,20,.20);
  --auth-orange-shadow: 0 22px 45px rgba(255,142,90,.20);
  --auth-yellow-shadow: 0 20px 44px rgba(224,208,54,.22);
  min-height: 100svh;
  overflow: hidden;
  background: var(--auth-bg);
  color: var(--auth-title);
}

.auth-classic-shell *,
.auth-classic-shell *::before,
.auth-classic-shell *::after {
  box-sizing: border-box;
}

.auth-classic-grid {
  display: grid;
  min-height: 100svh;
  grid-template-columns: minmax(0, 50%) minmax(430px, 50%);
}

.auth-classic-hero {
  position: relative;
  display: flex;
  min-height: 100svh;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.auth-classic-hero::before {
  content: '';
  position: absolute;
  inset: 0;
  background: var(--auth-left-aura);
}

.auth-classic-form-column {
  display: flex;
  min-height: 100svh;
  width: 100%;
  align-items: center;
  justify-content: center;
  padding: 40px;
}

.auth-classic-panel-wrap {
  width: min(420px, calc(100vw - 80px));
  max-width: 100%;
}

.auth-classic-brand {
  margin-bottom: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: var(--auth-brand);
  text-decoration: none;
}

.auth-classic-logo {
  height: 40px;
  width: 40px;
  border-radius: 16px;
  object-fit: cover;
  box-shadow: 0 10px 28px rgba(55,45,85,.12);
}

.auth-classic-brand-name {
  color: var(--auth-brand);
  font-size: 24px;
  font-weight: 800;
  letter-spacing: 0;
}

.auth-classic-card {
  width: 100%;
  border: 1px solid var(--auth-panel-border);
  border-radius: 28px;
  background: var(--auth-panel-bg);
  box-shadow: var(--auth-panel-shadow);
  padding: 32px 36px;
  backdrop-filter: blur(18px);
}

.auth-classic-card h2 {
  margin: 0;
  color: var(--auth-title);
  text-align: center !important;
  font-size: 30px;
  font-weight: 900;
  letter-spacing: 0 !important;
}

.auth-classic-card p {
  color: var(--auth-subtitle);
  letter-spacing: 0;
}

.auth-classic-card label {
  color: var(--auth-label);
  font-weight: 650;
  letter-spacing: 0;
}

.auth-classic-card a {
  color: var(--auth-link);
  text-decoration-color: rgba(95,53,207,.34);
}

.auth-classic-card a:hover {
  color: var(--auth-link-hover);
}

.auth-classic-card input {
  min-height: 44px;
  border-color: var(--auth-input-border);
  border-radius: 14px;
  background: var(--auth-input-bg);
  color: var(--auth-title);
  box-shadow: inset 0 1px 0 rgba(255,255,255,.78);
}

.auth-classic-card input:focus-visible {
  border-color: rgba(95,53,207,.52);
  box-shadow: 0 0 0 4px var(--auth-input-focus);
}

.auth-classic-card button {
  min-height: 42px;
  border-radius: 14px;
  letter-spacing: 0;
}

.auth-classic-card form > button:not([type='button']) {
  min-height: 46px;
  border: 0;
  background: var(--auth-primary-bg);
  color: var(--auth-primary-text);
  box-shadow: 0 10px 22px rgba(17,17,17,.16);
  transition:
    background-color 180ms ease,
    box-shadow 180ms ease;
}

.auth-classic-card form > button:not([type='button']):hover:not(:disabled) {
  background: #222222;
  box-shadow: 0 12px 24px rgba(17,17,17,.18);
}

.auth-classic-card .tracking-tight {
  letter-spacing: 0 !important;
}

.auth-classic-shell .auth-floor-shadow {
  background: var(--auth-floor-shadow);
}

.auth-classic-shell .auth-characters-wrapper {
  display: flex;
  align-items: flex-end;
  justify-content: center;
}

.auth-classic-shell .auth-characters-scene {
  position: absolute;
  bottom: 36px;
  left: 50%;
  width: 480px;
  height: 360px;
  transform: translateX(-50%);
}

.auth-classic-shell .auth-character {
  position: absolute;
  bottom: 0;
  transform-origin: bottom center;
  transition:
    height 700ms ease-in-out,
    transform 700ms ease-in-out;
  will-change: transform, height;
}

.auth-classic-shell .auth-char-purple {
  left: 60px;
  width: 170px;
  height: 370px;
  z-index: 1;
  border-radius: 10px 10px 0 0;
  background: #6c3ff5;
}

.auth-classic-shell .auth-char-black {
  left: 220px;
  width: 115px;
  height: 290px;
  z-index: 2;
  border-radius: 8px 8px 0 0;
  background: #2d2d2d;
}

.auth-classic-shell .auth-char-orange {
  left: 0;
  width: 230px;
  height: 190px;
  z-index: 3;
  border-radius: 115px 115px 0 0;
  background: #ff9b6b;
}

.auth-classic-shell .auth-char-yellow {
  left: 290px;
  width: 135px;
  height: 215px;
  z-index: 4;
  border-radius: 68px 68px 0 0;
  background: #e8d754;
}

.auth-classic-shell .auth-eyes {
  position: absolute;
  display: flex;
  transition:
    left 700ms ease-in-out,
    top 700ms ease-in-out;
}

.auth-classic-shell .auth-eyeball {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: 50%;
  background: #ffffff;
  transition: height 150ms ease;
}

.auth-classic-shell .auth-pupil,
.auth-classic-shell .auth-bare-pupil {
  display: inline-block;
  border-radius: 50%;
  background: #2d2d2d;
}

.auth-classic-shell .auth-pupil {
  transition: transform 100ms ease-out;
}

.auth-classic-shell .auth-bare-pupil {
  width: 12px;
  height: 12px;
  transition: transform 700ms ease-in-out;
}

.auth-classic-shell .auth-yellow-mouth {
  position: absolute;
  width: 50px;
  height: 4px;
  border-radius: 2px;
  background: #2d2d2d;
  transition:
    left 700ms ease-in-out,
    top 700ms ease-in-out,
    transform 700ms ease-in-out;
}

.auth-classic-shell .auth-orange-mouth {
  position: absolute;
  left: 90px;
  top: 120px;
  width: 28px;
  height: 14px;
  border: 3px solid #2d2d2d;
  border-top: 0;
  border-radius: 0 0 14px 14px;
  opacity: 0;
}

.auth-classic-shell .auth-purple-body {
  box-shadow: var(--auth-purple-shadow);
}

.auth-classic-shell .auth-black-body {
  box-shadow: var(--auth-black-shadow);
}

.auth-classic-shell .auth-orange-body {
  box-shadow: var(--auth-orange-shadow);
}

.auth-classic-shell .auth-yellow-body {
  box-shadow: var(--auth-yellow-shadow);
}

.auth-login-shell {
  --auth-bg: var(--background);
  --auth-panel-bg: var(--background);
  --auth-panel-border: var(--border);
  --auth-panel-shadow: none;
  --auth-brand: var(--foreground);
  --auth-title: var(--foreground);
  --auth-subtitle: var(--muted-foreground);
  --auth-label: var(--foreground);
  --auth-link: var(--primary);
  --auth-link-hover: var(--primary);
  --auth-primary-bg: var(--foreground);
  --auth-primary-text: var(--background);
  --auth-input-bg: var(--background);
  --auth-input-border: var(--border);
  --auth-input-focus: color-mix(in oklch, var(--primary) 28%, transparent);
  background: var(--background);
  color: var(--foreground);
}

.auth-login-shell .auth-classic-grid {
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
}

.auth-login-shell .auth-classic-hero {
  align-items: stretch;
  background: var(--background);
}

.auth-login-shell .auth-classic-hero::before {
  display: none;
}

.auth-login-hero-image {
  display: block;
  height: 100%;
  width: 100%;
  object-fit: cover;
  object-position: center;
}

.auth-login-shell .auth-classic-form-column {
  position: relative;
  background: var(--background);
  padding: 72px clamp(32px, 6vw, 92px);
}

.auth-login-shell .auth-classic-panel-wrap {
  width: min(420px, 100%);
}

.auth-login-shell .auth-classic-brand {
  display: none;
}

.auth-login-shell .auth-classic-card {
  border: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
  padding: 0;
  backdrop-filter: none;
}

.auth-login-shell .auth-classic-card h2 {
  text-align: left !important;
  font-size: 32px;
  font-weight: 750;
}

.auth-login-shell .auth-classic-card input {
  min-height: 46px;
  border-radius: 10px;
  background: var(--background);
}

.auth-login-shell .auth-classic-card form > button:not([type='button']) {
  min-height: 46px;
  border-radius: 10px;
}

.auth-login-shell .auth-classic-card form > button:not([type='button']):hover:not(:disabled) {
  background: color-mix(
    in oklch,
    var(--auth-primary-bg) 88%,
    var(--auth-primary-text)
  );
}

.auth-login-shell .auth-terms-footer {
  position: absolute;
  right: clamp(32px, 6vw, 92px);
  bottom: 28px;
  left: clamp(32px, 6vw, 92px);
  margin: 0;
  text-align: center;
}

@media (max-width: 1023px) {
  .auth-classic-shell:not(.auth-login-shell) {
    overflow-y: auto;
    overflow-x: hidden;
    background: linear-gradient(145deg, #e8e3f2 0%, #ffffff 56%);
  }

  .auth-classic-grid {
    display: block;
  }

  .auth-classic-hero {
    display: none;
  }

  .auth-classic-form-column {
    min-height: 100svh;
    padding: 28px 0;
  }

  .auth-classic-panel-wrap {
    width: min(420px, calc(100vw - 32px));
  }

  .auth-login-shell .auth-classic-grid {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
    min-height: 100svh;
  }

  .auth-login-shell .auth-classic-hero {
    display: flex;
    min-height: 100svh;
    max-height: none;
  }

  .auth-login-shell .auth-classic-form-column {
    min-height: 100svh;
    padding: 72px 40px;
  }

  .auth-login-shell .auth-terms-footer {
    right: 40px;
    left: 40px;
  }
}

@media (max-width: 767px) {
  .auth-login-shell .auth-classic-grid {
    display: block;
  }

  .auth-login-shell .auth-classic-hero {
    display: none;
  }

  .auth-login-shell .auth-classic-form-column {
    min-height: 100svh;
    align-items: flex-start;
    overflow-y: auto;
    padding: clamp(104px, 18svh, 168px) 20px 92px;
  }

  .auth-login-shell .auth-classic-panel-wrap {
    width: min(390px, 100%);
  }

  .auth-login-shell .auth-terms-footer {
    right: 20px;
    bottom: max(20px, env(safe-area-inset-bottom));
    left: 20px;
  }

  .auth-classic-panel-wrap {
    width: min(358px, calc(100vw - 32px));
  }

  .auth-classic-card {
    border-radius: 24px;
    padding: 28px 24px;
  }

  .auth-classic-brand {
    margin-bottom: 24px;
  }

  .auth-classic-card h2 {
    font-size: 26px;
  }

  .auth-login-shell .auth-classic-card h2 {
    font-size: 28px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .auth-character,
  .auth-char-purple,
  .auth-char-black,
  .auth-char-orange,
  .auth-char-yellow,
  .auth-eyes,
  .auth-pupil,
  .auth-bare-pupil {
    animation: none !important;
    transition: none !important;
  }
}
`

export function AuthLayout({ children, variant = 'default' }: AuthLayoutProps) {
  const { t } = useTranslation()
  const { systemName, logo, loading } = useSystemConfig()
  const isModern = variant === 'modern' || variant === 'sign-in'

  return (
    <div className={cn('auth-classic-shell', isModern && 'auth-login-shell')}>
      <style>{authClassicStyles}</style>
      <PublicHeader
        showAuthButtons={false}
        showBrand
        showNotifications={false}
        showNavigation
      />
      <div className='auth-classic-grid'>
        <section
          className={cn('auth-classic-hero', isModern && 'auth-login-hero')}
          aria-hidden={isModern ? undefined : true}
        >
          {isModern ? (
            <img
              src={loginHeroImage}
              alt=''
              className='auth-login-hero-image'
            />
          ) : (
            <AuthMascot />
          )}
        </section>

        <section className='auth-classic-form-column'>
          <div className='auth-classic-panel-wrap'>
            <Link to='/' className='auth-classic-brand hover:opacity-85'>
              <div className='relative h-10 w-10'>
                {loading ? (
                  <Skeleton className='absolute inset-0 rounded-2xl' />
                ) : (
                  <img
                    src={logo}
                    alt={t('Logo')}
                    className='auth-classic-logo'
                  />
                )}
              </div>
              {loading ? (
                <Skeleton className='h-7 w-36' />
              ) : (
                <div className='auth-classic-brand-name'>{systemName}</div>
              )}
            </Link>

            <div className='auth-classic-card'>{children}</div>
          </div>
        </section>
      </div>
    </div>
  )
}
