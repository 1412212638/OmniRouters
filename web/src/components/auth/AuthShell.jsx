/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useMemo, useState } from 'react';

const clamp = (value, min, max) => Math.min(Math.max(value, min), max);

const Eye = ({ x, y, light = true, size = 18 }) => {
  const pupilSize = Math.max(5, Math.round(size * 0.38));
  const travel = Math.max(2, Math.round(size * 0.16));

  return (
    <span
      style={{
        width: size,
        height: size,
        borderRadius: '50%',
        background: light ? '#ffffff' : 'transparent',
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <span
        style={{
          width: pupilSize,
          height: pupilSize,
          borderRadius: '50%',
          background: '#25252a',
          transform: `translate(${x * travel}px, ${y * travel}px)`,
          transition: 'transform 180ms cubic-bezier(.2,.8,.2,1)',
        }}
      />
    </span>
  );
};

const MascotBody = ({
  className = '',
  style,
  eyes = 'light',
  face = 'eyes',
  motion,
  sensitivity = 1,
  children,
}) => {
  const rotate = motion.x * 4 * sensitivity;
  const translateX = motion.x * 8 * sensitivity;
  const translateY = motion.y * 5 * sensitivity;
  const eyeX = clamp(motion.x * 1.8, -1, 1);
  const eyeY = clamp(motion.y * 1.8, -1, 1);

  return (
    <div
      className={className}
      style={{
        ...style,
        transform: `translate(${translateX}px, ${translateY}px) rotate(${rotate}deg)`,
        transition: motion.active
          ? 'transform 120ms cubic-bezier(.2,.8,.2,1)'
          : 'transform 520ms cubic-bezier(.2,.8,.2,1)',
        transformOrigin: '50% 100%',
      }}
    >
      {face === 'eyes' && (
        <div
          style={{
            display: 'flex',
            gap: 20,
            position: 'absolute',
            left: '50%',
            top: 44,
            transform: 'translateX(-50%)',
          }}
        >
          <Eye x={eyeX} y={eyeY} light={eyes === 'light'} />
          <Eye x={eyeX} y={eyeY} light={eyes === 'light'} />
        </div>
      )}
      {children}
    </div>
  );
};

const AuthMascot = () => {
  const prefersReducedMotion = useMemo(() => {
    if (typeof window === 'undefined') return false;
    return window.matchMedia?.('(prefers-reduced-motion: reduce)')?.matches;
  }, []);
  const [motion, setMotion] = useState({ x: 0, y: 0, active: false });

  const handlePointerMove = (event) => {
    if (prefersReducedMotion) return;
    const rect = event.currentTarget.getBoundingClientRect();
    const x = clamp((event.clientX - rect.left) / rect.width - 0.5, -0.5, 0.5);
    const y = clamp((event.clientY - rect.top) / rect.height - 0.5, -0.5, 0.5);
    setMotion({ x: x * 2, y: y * 2, active: true });
  };

  const handlePointerLeave = () => {
    setMotion({ x: 0, y: 0, active: false });
  };

  return (
    <div
      className='relative h-[420px] w-[520px] max-w-[78vw]'
      onPointerMove={handlePointerMove}
      onPointerLeave={handlePointerLeave}
      aria-hidden='true'
    >
      <div
        className='absolute bottom-0 left-16 h-4 w-[420px]'
        style={{
          background:
            'linear-gradient(90deg, transparent 0%, rgba(35,31,42,.12) 18%, rgba(35,31,42,.16) 55%, transparent 100%)',
          filter: 'blur(10px)',
        }}
      />

      <MascotBody
        motion={motion}
        sensitivity={1.08}
        className='absolute bottom-10 left-24 h-[300px] w-[150px]'
        style={{
          background: 'linear-gradient(180deg, #7542ff 0%, #6533ef 100%)',
          borderRadius: '18px 18px 10px 10px',
          boxShadow: '0 18px 45px rgba(83,49,217,.22)',
        }}
      />

      <MascotBody
        motion={motion}
        sensitivity={0.82}
        className='absolute bottom-10 left-[270px] h-[240px] w-[100px]'
        style={{
          background: '#29292b',
          borderRadius: '12px 12px 4px 4px',
          boxShadow: '0 16px 36px rgba(18,18,20,.2)',
        }}
      />

      <MascotBody
        motion={motion}
        sensitivity={0.52}
        eyes='dark'
        className='absolute bottom-10 left-8 h-[175px] w-[230px]'
        style={{
          background: 'linear-gradient(180deg, #ff9c6a 0%, #ff9261 100%)',
          borderRadius: '130px 130px 0 0',
          boxShadow: '0 22px 45px rgba(255,142,90,.2)',
        }}
      />

      <MascotBody
        motion={motion}
        sensitivity={0.66}
        eyes='dark'
        className='absolute bottom-10 left-[335px] h-[190px] w-[125px]'
        style={{
          background: 'linear-gradient(180deg, #efe153 0%, #e6d844 100%)',
          borderRadius: '84px 84px 0 0',
          boxShadow: '0 20px 44px rgba(224,208,54,.22)',
        }}
      >
        <span
          style={{
            position: 'absolute',
            left: '50%',
            top: 92,
            width: 48,
            height: 4,
            borderRadius: 8,
            background: '#25252a',
            transform: `translateX(-50%) translate(${motion.x * 3}px, ${motion.y * 1.5}px)`,
            transition: 'transform 180ms cubic-bezier(.2,.8,.2,1)',
          }}
        />
      </MascotBody>
    </div>
  );
};

const AuthShell = ({ logo, systemName, title, subtitle, children }) => {
  return (
    <div
      className='min-h-screen w-full overflow-hidden'
      style={{
        background:
          'radial-gradient(circle at 18% 16%, rgba(255,255,255,.78) 0, rgba(255,255,255,0) 28%), linear-gradient(115deg, #cfcbd9 0%, #e7e3ef 43%, #ffffff 43%, #ffffff 100%)',
      }}
    >
      <div className='grid min-h-screen grid-cols-1 lg:grid-cols-[56%_44%]'>
        <section className='relative hidden min-h-screen items-center justify-center overflow-hidden lg:flex'>
          <div
            className='absolute inset-0'
            style={{
              background:
                'radial-gradient(circle at 20% 18%, rgba(255,255,255,.5) 0, rgba(255,255,255,0) 26%), radial-gradient(circle at 34% 78%, rgba(255,255,255,.24) 0, rgba(255,255,255,0) 24%)',
            }}
          />
          <AuthMascot />
        </section>

        <section className='flex min-h-screen items-center justify-center px-6 py-10 sm:px-10'>
          <div className='w-full max-w-[420px]'>
            <div className='mb-8 flex items-center justify-center gap-3'>
              <img
                src={logo}
                alt='Logo'
                className='h-10 w-10 rounded-2xl object-cover shadow-sm'
              />
              <div className='text-2xl font-bold tracking-tight text-[#161827]'>
                {systemName}
              </div>
            </div>

            <div
              className='rounded-[28px] border border-white/70 bg-white/80 px-7 py-8 shadow-[0_24px_70px_rgba(40,36,60,.12)] backdrop-blur-xl sm:px-9'
              style={{
                boxShadow:
                  '0 24px 70px rgba(40,36,60,.12), inset 0 1px 0 rgba(255,255,255,.82)',
              }}
            >
              <div className='mb-7 text-center'>
                <div className='mx-auto mb-4 h-5 w-5 text-[#18182b]'>
                  <svg viewBox='0 0 24 24' fill='none'>
                    <path
                      d='M12 2.5l1.35 6.15L19.5 10l-6.15 1.35L12 17.5l-1.35-6.15L4.5 10l6.15-1.35L12 2.5z'
                      fill='currentColor'
                    />
                  </svg>
                </div>
                <h1 className='m-0 text-3xl font-black tracking-tight text-[#111326]'>
                  {title}
                </h1>
                {subtitle && (
                  <p className='mt-2 text-sm text-[#7d7b89]'>{subtitle}</p>
                )}
              </div>

              {children}
            </div>
          </div>
        </section>
      </div>
    </div>
  );
};

export default AuthShell;
