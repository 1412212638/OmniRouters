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

import React, { useEffect, useMemo, useRef, useState } from 'react';

const clamp = (value, min, max) => Math.min(Math.max(value, min), max);

const isPasswordInput = (input) => {
  if (!input || input.tagName !== 'INPUT') return false;
  const marker = `${input.type || ''} ${input.name || ''} ${input.id || ''} ${
    input.autocomplete || ''
  }`.toLowerCase();
  return marker.includes('password') || marker.includes('pwd');
};

const AuthMascot = () => {
  const prefersReducedMotion = useMemo(() => {
    if (typeof window === 'undefined') return false;
    return window.matchMedia?.('(prefers-reduced-motion: reduce)')?.matches;
  }, []);
  const [sceneState, setSceneState] = useState({
    mouseX: 0,
    mouseY: 0,
    isTyping: false,
    isLookingAtEachOther: false,
    isPurpleBlinking: false,
    isBlackBlinking: false,
    isPurplePeeking: false,
    isPasswordFocused: false,
    activePasswordVisible: false,
    activePasswordLength: 0,
  });
  const activePasswordInputRef = useRef(null);
  const purpleRef = useRef(null);
  const blackRef = useRef(null);
  const orangeRef = useRef(null);
  const yellowRef = useRef(null);
  const purpleEyeRef = useRef(null);
  const blackEyeRef = useRef(null);
  const orangePupilRef = useRef(null);
  const yellowPupilRef = useRef(null);
  const typingTimerRef = useRef(null);
  const peekTimerRef = useRef(null);
  const purpleBlinkTimerRef = useRef(null);
  const blackBlinkTimerRef = useRef(null);

  const setTyping = (typing) => {
    setSceneState((current) => ({
      ...current,
      isTyping: typing,
      isLookingAtEachOther: typing,
    }));
    if (typing) {
      window.clearTimeout(typingTimerRef.current);
      typingTimerRef.current = window.setTimeout(() => {
        setSceneState((current) => ({
          ...current,
          isLookingAtEachOther: false,
        }));
      }, 800);
    }
  };

  const syncPasswordState = (input) => {
    activePasswordInputRef.current = input;
    setSceneState((current) => ({
      ...current,
      isPasswordFocused: document.activeElement === input,
      activePasswordVisible: input?.type === 'text',
      activePasswordLength: input?.value?.length || 0,
    }));
  };

  const schedulePeek = () => {
    window.clearTimeout(peekTimerRef.current);
    const input = activePasswordInputRef.current;
    if (!input || input.type !== 'text' || !input.value.length) {
      setSceneState((current) => ({ ...current, isPurplePeeking: false }));
      return;
    }

    peekTimerRef.current = window.setTimeout(
      () => {
        const latestInput = activePasswordInputRef.current;
        if (!latestInput || latestInput.type !== 'text' || !latestInput.value) {
          setSceneState((current) => ({ ...current, isPurplePeeking: false }));
          return;
        }
        setSceneState((current) => ({ ...current, isPurplePeeking: true }));
        window.setTimeout(() => {
          setSceneState((current) => ({ ...current, isPurplePeeking: false }));
          schedulePeek();
        }, 800);
      },
      Math.random() * 3000 + 2000,
    );
  };

  useEffect(() => {
    if (typeof window === 'undefined') return undefined;

    setSceneState((current) => ({
      ...current,
      mouseX: window.innerWidth / 2,
      mouseY: window.innerHeight / 2,
    }));

    if (prefersReducedMotion) return undefined;

    const handleMouseMove = (event) => {
      setSceneState((current) => ({
        ...current,
        mouseX: event.clientX,
        mouseY: event.clientY,
      }));
    };

    const handleFocusIn = (event) => {
      const target = event.target;
      if (!target?.closest?.('.auth-shell') || target.tagName !== 'INPUT') {
        return;
      }
      if (isPasswordInput(target)) {
        syncPasswordState(target);
        window.setTimeout(() => syncPasswordState(target), 80);
        schedulePeek();
      } else {
        setTyping(true);
      }
    };

    const handleFocusOut = (event) => {
      const target = event.target;
      if (!target?.closest?.('.auth-shell') || target.tagName !== 'INPUT') {
        return;
      }
      if (isPasswordInput(target)) {
        window.setTimeout(() => {
          if (document.activeElement !== target) {
            setSceneState((current) => ({
              ...current,
              isPasswordFocused: false,
            }));
          }
        }, 60);
      } else {
        setSceneState((current) => ({
          ...current,
          isTyping: false,
          isLookingAtEachOther: false,
        }));
      }
    };

    const handleInput = (event) => {
      const target = event.target;
      if (!target?.closest?.('.auth-shell') || target.tagName !== 'INPUT') {
        return;
      }
      if (isPasswordInput(target)) {
        syncPasswordState(target);
        schedulePeek();
      } else {
        setTyping(true);
      }
    };

    const handleDocumentClick = () => {
      const input = activePasswordInputRef.current;
      if (input) window.setTimeout(() => syncPasswordState(input), 80);
    };

    const scheduleBlink = (key, timerRef) => {
      timerRef.current = window.setTimeout(
        () => {
          setSceneState((current) => ({ ...current, [key]: true }));
          window.setTimeout(() => {
            setSceneState((current) => ({ ...current, [key]: false }));
            scheduleBlink(key, timerRef);
          }, 150);
        },
        Math.random() * 4000 + 3000,
      );
    };

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('focusin', handleFocusIn);
    document.addEventListener('focusout', handleFocusOut);
    document.addEventListener('input', handleInput);
    document.addEventListener('click', handleDocumentClick);
    scheduleBlink('isPurpleBlinking', purpleBlinkTimerRef);
    scheduleBlink('isBlackBlinking', blackBlinkTimerRef);

    return () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('focusin', handleFocusIn);
      document.removeEventListener('focusout', handleFocusOut);
      document.removeEventListener('input', handleInput);
      document.removeEventListener('click', handleDocumentClick);
      window.clearTimeout(typingTimerRef.current);
      window.clearTimeout(peekTimerRef.current);
      window.clearTimeout(purpleBlinkTimerRef.current);
      window.clearTimeout(blackBlinkTimerRef.current);
    };
  }, [prefersReducedMotion]);

  const calculatePosition = (ref) => {
    const element = ref.current;
    if (!element) return { faceX: 0, faceY: 0, bodySkew: 0 };
    const rect = element.getBoundingClientRect();
    const centerX = rect.left + rect.width / 2;
    const centerY = rect.top + rect.height / 3;
    const dx = sceneState.mouseX - centerX;
    const dy = sceneState.mouseY - centerY;

    return {
      faceX: clamp(dx / 20, -15, 15),
      faceY: clamp(dy / 30, -10, 10),
      bodySkew: clamp(-dx / 120, -6, 6),
    };
  };

  const calculatePupilOffset = (ref, maxDistance) => {
    const element = ref.current;
    if (!element) return { x: 0, y: 0 };
    const rect = element.getBoundingClientRect();
    const centerX = rect.left + rect.width / 2;
    const centerY = rect.top + rect.height / 2;
    const dx = sceneState.mouseX - centerX;
    const dy = sceneState.mouseY - centerY;
    const distance = Math.min(Math.sqrt(dx * dx + dy * dy), maxDistance);
    const angle = Math.atan2(dy, dx);

    return {
      x: Math.cos(angle) * distance,
      y: Math.sin(angle) * distance,
    };
  };

  const purplePos = calculatePosition(purpleRef);
  const blackPos = calculatePosition(blackRef);
  const orangePos = calculatePosition(orangeRef);
  const yellowPos = calculatePosition(yellowRef);
  const isShowingPassword =
    sceneState.activePasswordLength > 0 && sceneState.activePasswordVisible;
  const isLookingAway =
    sceneState.isPasswordFocused && !sceneState.activePasswordVisible;

  const purplePupil = calculatePupilOffset(purpleEyeRef, 5);
  const blackPupil = calculatePupilOffset(blackEyeRef, 4);
  const orangePupil = calculatePupilOffset(orangePupilRef, 5);
  const yellowPupil = calculatePupilOffset(yellowPupilRef, 5);

  let purpleTransform = `skewX(${purplePos.bodySkew}deg)`;
  let purpleHeight = 370;
  let purpleEyes = {
    left: 45 + purplePos.faceX,
    top: 40 + purplePos.faceY,
    pupilX: purplePupil.x,
    pupilY: purplePupil.y,
  };

  if (isShowingPassword) {
    purpleTransform = 'skewX(0deg)';
    purpleEyes = {
      left: 20,
      top: 35,
      pupilX: sceneState.isPurplePeeking ? 4 : -4,
      pupilY: sceneState.isPurplePeeking ? 5 : -4,
    };
  } else if (isLookingAway) {
    purpleTransform = 'skewX(-14deg) translateX(-20px)';
    purpleHeight = 410;
    purpleEyes = { left: 20, top: 25, pupilX: -5, pupilY: -5 };
  } else if (sceneState.isTyping) {
    purpleTransform = `skewX(${purplePos.bodySkew - 12}deg) translateX(40px)`;
    purpleHeight = 410;
  }

  if (!isShowingPassword && !isLookingAway && sceneState.isLookingAtEachOther) {
    purpleEyes = { left: 55, top: 65, pupilX: 3, pupilY: 4 };
  }

  let blackTransform = `skewX(${blackPos.bodySkew}deg)`;
  let blackEyes = {
    left: 26 + blackPos.faceX,
    top: 32 + blackPos.faceY,
    pupilX: blackPupil.x,
    pupilY: blackPupil.y,
  };

  if (isShowingPassword) {
    blackTransform = 'skewX(0deg)';
    blackEyes = { left: 10, top: 28, pupilX: -4, pupilY: -4 };
  } else if (isLookingAway) {
    blackTransform = 'skewX(12deg) translateX(-10px)';
    blackEyes = { left: 10, top: 20, pupilX: -4, pupilY: -5 };
  } else if (sceneState.isLookingAtEachOther) {
    blackTransform = `skewX(${blackPos.bodySkew * 1.5 + 10}deg) translateX(20px)`;
    blackEyes = { left: 32, top: 12, pupilX: 0, pupilY: -4 };
  } else if (sceneState.isTyping) {
    blackTransform = `skewX(${blackPos.bodySkew * 1.5}deg)`;
  }

  let orangeEyes = {
    left: 82 + orangePos.faceX,
    top: 90 + orangePos.faceY,
    pupilX: orangePupil.x,
    pupilY: orangePupil.y,
  };
  if (isLookingAway) {
    orangeEyes = { left: 50, top: 75, pupilX: -5, pupilY: -5 };
  } else if (isShowingPassword) {
    orangeEyes = { left: 50, top: 85, pupilX: -5, pupilY: -4 };
  }

  let yellowEyes = {
    left: 52 + yellowPos.faceX,
    top: 40 + yellowPos.faceY,
    pupilX: yellowPupil.x,
    pupilY: yellowPupil.y,
  };
  let yellowMouth = {
    left: 40 + yellowPos.faceX,
    top: 88 + yellowPos.faceY,
  };
  if (isLookingAway) {
    yellowEyes = { left: 20, top: 30, pupilX: -5, pupilY: -5 };
    yellowMouth = { left: 15, top: 78 };
  } else if (isShowingPassword) {
    yellowEyes = { left: 20, top: 35, pupilX: -5, pupilY: -4 };
    yellowMouth = { left: 10, top: 88 };
  }

  return (
    <div
      className='auth-characters-wrapper relative h-[420px] w-[520px] max-w-[78vw]'
      aria-hidden='true'
    >
      <div
        className='auth-floor-shadow absolute bottom-0 left-16 h-4 w-[420px]'
        style={{
          filter: 'blur(10px)',
        }}
      />

      <div className='auth-characters-scene'>
        <div
          ref={purpleRef}
          className='auth-character auth-purple-body auth-char-purple'
          style={{
            height: purpleHeight,
            transform: purpleTransform,
          }}
        >
          <div
            className='auth-eyes'
            style={{
              left: purpleEyes.left,
              top: purpleEyes.top,
              gap: 28,
            }}
          >
            <span
              ref={purpleEyeRef}
              className='auth-eyeball'
              style={{
                width: 18,
                height: sceneState.isPurpleBlinking ? 2 : 18,
              }}
            >
              <span
                className='auth-pupil'
                style={{
                  width: 7,
                  height: 7,
                  transform: `translate(${purpleEyes.pupilX}px, ${purpleEyes.pupilY}px)`,
                }}
              />
            </span>
            <span
              className='auth-eyeball'
              style={{
                width: 18,
                height: sceneState.isPurpleBlinking ? 2 : 18,
              }}
            >
              <span
                className='auth-pupil'
                style={{
                  width: 7,
                  height: 7,
                  transform: `translate(${purpleEyes.pupilX}px, ${purpleEyes.pupilY}px)`,
                }}
              />
            </span>
          </div>
        </div>

        <div
          ref={blackRef}
          className='auth-character auth-black-body auth-char-black'
          style={{
            transform: blackTransform,
          }}
        >
          <div
            className='auth-eyes'
            style={{
              left: blackEyes.left,
              top: blackEyes.top,
              gap: 20,
            }}
          >
            <span
              ref={blackEyeRef}
              className='auth-eyeball'
              style={{
                width: 16,
                height: sceneState.isBlackBlinking ? 2 : 16,
              }}
            >
              <span
                className='auth-pupil'
                style={{
                  width: 6,
                  height: 6,
                  transform: `translate(${blackEyes.pupilX}px, ${blackEyes.pupilY}px)`,
                }}
              />
            </span>
            <span
              className='auth-eyeball'
              style={{
                width: 16,
                height: sceneState.isBlackBlinking ? 2 : 16,
              }}
            >
              <span
                className='auth-pupil'
                style={{
                  width: 6,
                  height: 6,
                  transform: `translate(${blackEyes.pupilX}px, ${blackEyes.pupilY}px)`,
                }}
              />
            </span>
          </div>
        </div>

        <div
          ref={orangeRef}
          className='auth-character auth-orange-body auth-char-orange'
          style={{
            transform: isShowingPassword
              ? 'skewX(0deg)'
              : `skewX(${orangePos.bodySkew}deg)`,
          }}
        >
          <div
            className='auth-eyes'
            style={{
              left: orangeEyes.left,
              top: orangeEyes.top,
              gap: 28,
            }}
          >
            <span
              ref={orangePupilRef}
              className='auth-bare-pupil'
              style={{
                transform: `translate(${orangeEyes.pupilX}px, ${orangeEyes.pupilY}px)`,
              }}
            />
            <span
              className='auth-bare-pupil'
              style={{
                transform: `translate(${orangeEyes.pupilX}px, ${orangeEyes.pupilY}px)`,
              }}
            />
          </div>
          <div className='auth-orange-mouth' />
        </div>

        <div
          ref={yellowRef}
          className='auth-character auth-yellow-body auth-char-yellow'
          style={{
            transform: isShowingPassword
              ? 'skewX(0deg)'
              : `skewX(${yellowPos.bodySkew}deg)`,
          }}
        >
          <div
            className='auth-eyes'
            style={{
              left: yellowEyes.left,
              top: yellowEyes.top,
              gap: 20,
            }}
          >
            <span
              ref={yellowPupilRef}
              className='auth-bare-pupil'
              style={{
                transform: `translate(${yellowEyes.pupilX}px, ${yellowEyes.pupilY}px)`,
              }}
            />
            <span
              className='auth-bare-pupil'
              style={{
                transform: `translate(${yellowEyes.pupilX}px, ${yellowEyes.pupilY}px)`,
              }}
            />
          </div>
          <div
            className='auth-yellow-mouth'
            style={{
              left: yellowMouth.left,
              top: yellowMouth.top,
            }}
          />
        </div>
      </div>
    </div>
  );
};

const authShellStyles = `
.auth-shell {
  --auth-bg: radial-gradient(circle at 18% 16%, rgba(255,255,255,.78) 0, rgba(255,255,255,0) 28%), linear-gradient(115deg, #cfcbd9 0%, #e7e3ef 43%, #ffffff 43%, #ffffff 100%);
  --auth-left-aura: radial-gradient(circle at 20% 18%, rgba(255,255,255,.5) 0, rgba(255,255,255,0) 26%), radial-gradient(circle at 34% 78%, rgba(255,255,255,.24) 0, rgba(255,255,255,0) 24%);
  --auth-panel-bg: rgba(255,255,255,.80);
  --auth-panel-border: rgba(255,255,255,.70);
  --auth-panel-shadow: 0 24px 70px rgba(40,36,60,.12), inset 0 1px 0 rgba(255,255,255,.82);
  --auth-brand: #161827;
  --auth-title: #111326;
  --auth-subtitle: #7d7b89;
  --auth-text: #25243a;
  --auth-muted: #6d6a7a;
  --auth-label: #27243a;
  --auth-oauth-bg: rgba(255,255,255,.80);
  --auth-oauth-bg-hover: #f7f4ff;
  --auth-oauth-border: #e8e6ef;
  --auth-oauth-border-hover: #d8d0f2;
  --auth-oauth-text: #202033;
  --auth-oauth-shadow: none;
  --auth-primary-bg: #19182c;
  --auth-primary-text: #ffffff;
  --auth-primary-hover: #272542;
  --auth-primary-hover-layer: #6c3ff5;
  --auth-primary-hover-layer-text: #ffffff;
  --auth-primary-disabled-bg: rgba(25,24,44,.72);
  --auth-primary-disabled-text: rgba(255,255,255,.78);
  --auth-link: #6c3ff5;
  --auth-link-hover: #5b21b6;
  --auth-input-bg: rgba(255,255,255,.72);
  --auth-input-bg-hover: rgba(255,255,255,.90);
  --auth-input-border: rgba(215,211,226,.92);
  --auth-input-border-focus: rgba(112,75,255,.52);
  --auth-input-placeholder: #aaa6b5;
  --auth-floor-shadow: linear-gradient(90deg, transparent 0%, rgba(35,31,42,.12) 18%, rgba(35,31,42,.16) 55%, transparent 100%);
  --auth-purple-shadow: 0 18px 45px rgba(83,49,217,.22);
  --auth-black-shadow: 0 16px 36px rgba(18,18,20,.20);
  --auth-orange-shadow: 0 22px 45px rgba(255,142,90,.20);
  --auth-yellow-shadow: 0 20px 44px rgba(224,208,54,.22);
  background: var(--auth-bg);
  color: var(--auth-text);
}

[theme-mode='dark'] .auth-shell,
.dark .auth-shell {
  --auth-bg: radial-gradient(circle at 18% 16%, rgba(129,109,255,.20) 0, rgba(129,109,255,0) 31%), radial-gradient(circle at 74% 76%, rgba(255,139,91,.10) 0, rgba(255,139,91,0) 27%), linear-gradient(115deg, #161623 0%, #1d1b2c 43%, #08090f 43%, #10111a 100%);
  --auth-left-aura: radial-gradient(circle at 20% 18%, rgba(164,148,255,.18) 0, rgba(164,148,255,0) 28%), radial-gradient(circle at 34% 78%, rgba(255,255,255,.07) 0, rgba(255,255,255,0) 24%);
  --auth-panel-bg: rgba(17,18,30,.80);
  --auth-panel-border: rgba(255,255,255,.11);
  --auth-panel-shadow: 0 24px 90px rgba(0,0,0,.42), inset 0 1px 0 rgba(255,255,255,.08);
  --auth-brand: #f7f4ff;
  --auth-title: #fffaff;
  --auth-subtitle: #aaa5bb;
  --auth-text: #eeeaf8;
  --auth-muted: #aaa5bb;
  --auth-label: #e7e2f5;
  --auth-oauth-bg: rgba(255,255,255,.065);
  --auth-oauth-bg-hover: rgba(255,255,255,.11);
  --auth-oauth-border: rgba(255,255,255,.11);
  --auth-oauth-border-hover: rgba(194,177,255,.38);
  --auth-oauth-text: #f3efff;
  --auth-oauth-shadow: inset 0 1px 0 rgba(255,255,255,.05);
  --auth-primary-bg: #f4f0ff;
  --auth-primary-text: #151423;
  --auth-primary-hover: #ffffff;
  --auth-primary-hover-layer: #6c3ff5;
  --auth-primary-hover-layer-text: #ffffff;
  --auth-primary-disabled-bg: rgba(255,255,255,.16);
  --auth-primary-disabled-text: rgba(255,255,255,.68);
  --auth-link: #c7b8ff;
  --auth-link-hover: #e4dcff;
  --auth-input-bg: rgba(255,255,255,.065);
  --auth-input-bg-hover: rgba(255,255,255,.09);
  --auth-input-border: rgba(255,255,255,.12);
  --auth-input-border-focus: rgba(185,167,255,.62);
  --auth-input-placeholder: #7f7a91;
  --auth-floor-shadow: linear-gradient(90deg, transparent 0%, rgba(0,0,0,.24) 18%, rgba(0,0,0,.36) 55%, transparent 100%);
  --auth-purple-shadow: 0 18px 52px rgba(82,57,211,.34);
  --auth-black-shadow: 0 16px 40px rgba(0,0,0,.32);
  --auth-orange-shadow: 0 22px 50px rgba(255,142,90,.18);
  --auth-yellow-shadow: 0 20px 48px rgba(224,208,54,.16);
}

.auth-shell .auth-left-aura {
  background: var(--auth-left-aura);
}

.auth-shell .auth-brand-name,
.auth-shell .auth-title,
.auth-shell .auth-subtitle,
.auth-shell .auth-muted,
.auth-shell .auth-link,
.auth-shell .auth-panel,
.auth-shell .auth-oauth-button,
.auth-shell .auth-primary-button,
.auth-shell .auth-telegram-wrap,
.auth-shell .semi-input-wrapper {
  transition:
    background-color 220ms ease,
    border-color 220ms ease,
    box-shadow 220ms ease,
    color 220ms ease;
}

.auth-shell .auth-panel {
  background: var(--auth-panel-bg);
  border-color: var(--auth-panel-border);
  box-shadow: var(--auth-panel-shadow);
}

.auth-shell .auth-brand-name {
  color: var(--auth-brand);
}

.auth-shell .auth-title {
  color: var(--auth-title);
}

.auth-shell .auth-subtitle {
  color: var(--auth-subtitle);
}

.auth-shell .auth-floor-shadow {
  background: var(--auth-floor-shadow);
}

.auth-shell .auth-characters-wrapper {
  display: flex;
  align-items: flex-end;
  justify-content: center;
}

.auth-shell .auth-characters-scene {
  position: absolute;
  bottom: 36px;
  left: 50%;
  width: 480px;
  height: 360px;
  transform: translateX(-50%);
}

.auth-shell .auth-character {
  position: absolute;
  bottom: 0;
  transform-origin: bottom center;
  transition:
    height 700ms ease-in-out,
    transform 700ms ease-in-out;
  will-change: transform, height;
}

.auth-shell .auth-char-purple {
  left: 60px;
  width: 170px;
  height: 370px;
  z-index: 1;
  border-radius: 10px 10px 0 0;
  background: #6c3ff5;
}

.auth-shell .auth-char-black {
  left: 220px;
  width: 115px;
  height: 290px;
  z-index: 2;
  border-radius: 8px 8px 0 0;
  background: #2d2d2d;
}

.auth-shell .auth-char-orange {
  left: 0;
  width: 230px;
  height: 190px;
  z-index: 3;
  border-radius: 115px 115px 0 0;
  background: #ff9b6b;
}

.auth-shell .auth-char-yellow {
  left: 290px;
  width: 135px;
  height: 215px;
  z-index: 4;
  border-radius: 68px 68px 0 0;
  background: #e8d754;
}

.auth-shell .auth-eyes {
  position: absolute;
  display: flex;
  transition:
    left 700ms ease-in-out,
    top 700ms ease-in-out;
}

.auth-shell .auth-eyeball {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: 50%;
  background: #ffffff;
  transition: height 150ms ease;
}

.auth-shell .auth-pupil,
.auth-shell .auth-bare-pupil {
  display: inline-block;
  border-radius: 50%;
  background: #2d2d2d;
}

.auth-shell .auth-pupil {
  transition: transform 100ms ease-out;
}

.auth-shell .auth-bare-pupil {
  width: 12px;
  height: 12px;
  transition: transform 700ms ease-in-out;
}

.auth-shell .auth-yellow-mouth {
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

.auth-shell .auth-orange-mouth {
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

.auth-shell .auth-purple-body {
  box-shadow: var(--auth-purple-shadow);
}

.auth-shell .auth-black-body {
  box-shadow: var(--auth-black-shadow);
}

.auth-shell .auth-orange-body {
  box-shadow: var(--auth-orange-shadow);
}

.auth-shell .auth-yellow-body {
  box-shadow: var(--auth-yellow-shadow);
}

.auth-shell .semi-form-field-label {
  color: var(--auth-label) !important;
}

.auth-shell .semi-input-wrapper {
  background: var(--auth-input-bg) !important;
  border-color: var(--auth-input-border) !important;
}

.auth-shell .semi-input-wrapper:hover,
.auth-shell .semi-input-wrapper-focus {
  background: var(--auth-input-bg-hover) !important;
  border-color: var(--auth-input-border-focus) !important;
}

.auth-shell .semi-input {
  color: var(--auth-text) !important;
}

.auth-shell .semi-input::placeholder {
  color: var(--auth-input-placeholder) !important;
}

.auth-shell .auth-oauth-button.semi-button {
  background: var(--auth-oauth-bg) !important;
  border-color: var(--auth-oauth-border) !important;
  box-shadow: var(--auth-oauth-shadow) !important;
  color: var(--auth-oauth-text) !important;
}

.auth-shell .auth-oauth-button.semi-button:hover {
  background: var(--auth-oauth-bg-hover) !important;
  border-color: var(--auth-oauth-border-hover) !important;
}

.auth-shell .auth-telegram-wrap {
  background: var(--auth-oauth-bg);
  border-color: var(--auth-oauth-border);
  box-shadow: var(--auth-oauth-shadow);
}

.auth-shell .auth-primary-button.semi-button {
  position: relative !important;
  overflow: hidden !important;
  background: var(--auth-primary-bg) !important;
  border-color: transparent !important;
  color: var(--auth-primary-text) !important;
}

.auth-shell .auth-primary-button .semi-button-content {
  position: relative;
  z-index: 1;
  display: inline-flex;
  width: 100%;
  align-items: center;
  justify-content: center;
}

.auth-shell .auth-primary-text {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  transition:
    opacity 300ms ease,
    transform 300ms ease;
}

.auth-shell .auth-primary-hover-content {
  position: absolute;
  inset: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  border-radius: inherit;
  background: var(--auth-primary-hover-layer);
  color: var(--auth-primary-hover-layer-text);
  opacity: 0;
  transform: translateX(-10px);
  transition:
    opacity 300ms ease,
    transform 300ms ease;
}

.auth-shell .auth-primary-arrow {
  display: inline-flex;
  transform: translateX(0);
  transition: transform 300ms ease;
}

.auth-shell .auth-primary-button.semi-button:hover {
  background: var(--auth-primary-hover) !important;
}

.auth-shell
  .auth-primary-button.semi-button:not(:disabled):not([disabled]):not(
    .semi-button-disabled
  ):hover
  .auth-primary-text {
  opacity: 0;
  transform: translateX(40px);
}

.auth-shell
  .auth-primary-button.semi-button:not(:disabled):not([disabled]):not(
    .semi-button-disabled
  ):hover
  .auth-primary-hover-content {
  opacity: 1;
  transform: translateX(0);
}

.auth-shell
  .auth-primary-button.semi-button:not(:disabled):not([disabled]):not(
    .semi-button-disabled
  ):hover
  .auth-primary-arrow {
  transform: translateX(3px);
}

.auth-shell .auth-primary-button.semi-button:disabled,
.auth-shell .auth-primary-button.semi-button[disabled],
.auth-shell .auth-primary-button.semi-button.semi-button-disabled {
  background: var(--auth-primary-disabled-bg) !important;
  border-color: transparent !important;
  color: var(--auth-primary-disabled-text) !important;
  cursor: not-allowed;
  opacity: 1 !important;
}

.auth-shell .auth-primary-button.semi-button:disabled *,
.auth-shell .auth-primary-button.semi-button[disabled] *,
.auth-shell .auth-primary-button.semi-button.semi-button-disabled * {
  color: inherit !important;
}

.auth-shell a,
.auth-shell .auth-link {
  color: var(--auth-link) !important;
}

.auth-shell a:hover,
.auth-shell .auth-link:hover {
  color: var(--auth-link-hover) !important;
}

.auth-shell .auth-muted,
.auth-shell .semi-typography.auth-muted {
  color: var(--auth-muted) !important;
}
`;

const AuthShell = ({ logo, systemName, title, subtitle, children }) => {
  return (
    <div className='auth-shell min-h-screen w-full overflow-hidden'>
      <style>{authShellStyles}</style>
      <div className='grid min-h-screen grid-cols-1 lg:grid-cols-[56%_44%]'>
        <section className='relative hidden min-h-screen items-center justify-center overflow-hidden lg:flex'>
          <div className='auth-left-aura absolute inset-0' />
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
              <div className='auth-brand-name text-2xl font-bold tracking-tight'>
                {systemName}
              </div>
            </div>

            <div className='auth-panel rounded-[28px] border px-7 py-8 backdrop-blur-xl sm:px-9'>
              <div className='mb-7 text-center'>
                <h1 className='auth-title m-0 text-3xl font-black tracking-tight'>
                  {title}
                </h1>
                {subtitle && (
                  <p className='auth-subtitle mt-2 text-sm'>{subtitle}</p>
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
