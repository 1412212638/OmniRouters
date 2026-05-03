import { useEffect, useMemo, useRef, useState } from 'react'

const clamp = (value: number, min: number, max: number) =>
  Math.min(Math.max(value, min), max)

const isPasswordInput = (input: HTMLInputElement | null) => {
  if (!input) return false
  const marker = `${input.type || ''} ${input.name || ''} ${input.id || ''} ${
    input.autocomplete || ''
  }`.toLowerCase()
  return marker.includes('password') || marker.includes('pwd')
}

export function AuthMascot() {
  const prefersReducedMotion = useMemo(() => {
    if (typeof window === 'undefined') return false
    return window.matchMedia?.('(prefers-reduced-motion: reduce)')?.matches
  }, [])

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
  })

  const activePasswordInputRef = useRef<HTMLInputElement | null>(null)
  const purpleRef = useRef<HTMLDivElement | null>(null)
  const blackRef = useRef<HTMLDivElement | null>(null)
  const orangeRef = useRef<HTMLDivElement | null>(null)
  const yellowRef = useRef<HTMLDivElement | null>(null)
  const purpleEyeRef = useRef<HTMLSpanElement | null>(null)
  const blackEyeRef = useRef<HTMLSpanElement | null>(null)
  const orangePupilRef = useRef<HTMLSpanElement | null>(null)
  const yellowPupilRef = useRef<HTMLSpanElement | null>(null)
  const typingTimerRef = useRef<number | undefined>(undefined)
  const peekTimerRef = useRef<number | undefined>(undefined)
  const purpleBlinkTimerRef = useRef<number | undefined>(undefined)
  const blackBlinkTimerRef = useRef<number | undefined>(undefined)

  const setTyping = (typing: boolean) => {
    setSceneState((current) => ({
      ...current,
      isTyping: typing,
      isLookingAtEachOther: typing,
    }))
    if (typing) {
      window.clearTimeout(typingTimerRef.current)
      typingTimerRef.current = window.setTimeout(() => {
        setSceneState((current) => ({
          ...current,
          isLookingAtEachOther: false,
        }))
      }, 800)
    }
  }

  const syncPasswordState = (input: HTMLInputElement | null) => {
    activePasswordInputRef.current = input
    setSceneState((current) => ({
      ...current,
      isPasswordFocused: document.activeElement === input,
      activePasswordVisible: input?.type === 'text',
      activePasswordLength: input?.value?.length || 0,
    }))
  }

  const schedulePeek = () => {
    window.clearTimeout(peekTimerRef.current)
    const input = activePasswordInputRef.current
    if (!input || input.type !== 'text' || !input.value.length) {
      setSceneState((current) => ({ ...current, isPurplePeeking: false }))
      return
    }

    peekTimerRef.current = window.setTimeout(
      () => {
        const latestInput = activePasswordInputRef.current
        if (!latestInput || latestInput.type !== 'text' || !latestInput.value) {
          setSceneState((current) => ({ ...current, isPurplePeeking: false }))
          return
        }
        setSceneState((current) => ({ ...current, isPurplePeeking: true }))
        window.setTimeout(() => {
          setSceneState((current) => ({ ...current, isPurplePeeking: false }))
          schedulePeek()
        }, 800)
      },
      Math.random() * 3000 + 2000
    )
  }

  useEffect(() => {
    if (typeof window === 'undefined') return undefined

    setSceneState((current) => ({
      ...current,
      mouseX: window.innerWidth / 2,
      mouseY: window.innerHeight / 2,
    }))

    if (prefersReducedMotion) return undefined

    const handleMouseMove = (event: MouseEvent) => {
      setSceneState((current) => ({
        ...current,
        mouseX: event.clientX,
        mouseY: event.clientY,
      }))
    }

    const handleFocusIn = (event: FocusEvent) => {
      const target = event.target
      if (
        !(target instanceof HTMLInputElement) ||
        !target.closest('.auth-classic-shell')
      ) {
        return
      }

      if (isPasswordInput(target)) {
        syncPasswordState(target)
        window.setTimeout(() => syncPasswordState(target), 80)
        schedulePeek()
      } else {
        setTyping(true)
      }
    }

    const handleFocusOut = (event: FocusEvent) => {
      const target = event.target
      if (
        !(target instanceof HTMLInputElement) ||
        !target.closest('.auth-classic-shell')
      ) {
        return
      }

      if (isPasswordInput(target)) {
        window.setTimeout(() => {
          if (document.activeElement !== target) {
            setSceneState((current) => ({
              ...current,
              isPasswordFocused: false,
            }))
          }
        }, 60)
      } else {
        setSceneState((current) => ({
          ...current,
          isTyping: false,
          isLookingAtEachOther: false,
        }))
      }
    }

    const handleInput = (event: Event) => {
      const target = event.target
      if (
        !(target instanceof HTMLInputElement) ||
        !target.closest('.auth-classic-shell')
      ) {
        return
      }

      if (isPasswordInput(target)) {
        syncPasswordState(target)
        schedulePeek()
      } else {
        setTyping(true)
      }
    }

    const handleDocumentClick = () => {
      const input = activePasswordInputRef.current
      if (input) window.setTimeout(() => syncPasswordState(input), 80)
    }

    const scheduleBlink = (
      key: 'isPurpleBlinking' | 'isBlackBlinking',
      timerRef: React.MutableRefObject<number | undefined>
    ) => {
      timerRef.current = window.setTimeout(
        () => {
          setSceneState((current) => ({ ...current, [key]: true }))
          window.setTimeout(() => {
            setSceneState((current) => ({ ...current, [key]: false }))
            scheduleBlink(key, timerRef)
          }, 150)
        },
        Math.random() * 4000 + 3000
      )
    }

    document.addEventListener('mousemove', handleMouseMove)
    document.addEventListener('focusin', handleFocusIn)
    document.addEventListener('focusout', handleFocusOut)
    document.addEventListener('input', handleInput)
    document.addEventListener('click', handleDocumentClick)
    scheduleBlink('isPurpleBlinking', purpleBlinkTimerRef)
    scheduleBlink('isBlackBlinking', blackBlinkTimerRef)

    return () => {
      document.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('focusin', handleFocusIn)
      document.removeEventListener('focusout', handleFocusOut)
      document.removeEventListener('input', handleInput)
      document.removeEventListener('click', handleDocumentClick)
      window.clearTimeout(typingTimerRef.current)
      window.clearTimeout(peekTimerRef.current)
      window.clearTimeout(purpleBlinkTimerRef.current)
      window.clearTimeout(blackBlinkTimerRef.current)
    }
  }, [prefersReducedMotion])

  const calculatePosition = (element: HTMLElement | null) => {
    if (!element) return { faceX: 0, faceY: 0, bodySkew: 0 }
    const rect = element.getBoundingClientRect()
    const centerX = rect.left + rect.width / 2
    const centerY = rect.top + rect.height / 3
    const dx = sceneState.mouseX - centerX
    const dy = sceneState.mouseY - centerY

    return {
      faceX: clamp(dx / 20, -15, 15),
      faceY: clamp(dy / 30, -10, 10),
      bodySkew: clamp(-dx / 120, -6, 6),
    }
  }

  const calculatePupilOffset = (
    element: HTMLElement | null,
    maxDistance: number
  ) => {
    if (!element) return { x: 0, y: 0 }
    const rect = element.getBoundingClientRect()
    const centerX = rect.left + rect.width / 2
    const centerY = rect.top + rect.height / 2
    const dx = sceneState.mouseX - centerX
    const dy = sceneState.mouseY - centerY
    const distance = Math.min(Math.sqrt(dx * dx + dy * dy), maxDistance)
    const angle = Math.atan2(dy, dx)

    return {
      x: Math.cos(angle) * distance,
      y: Math.sin(angle) * distance,
    }
  }

  const purplePos = calculatePosition(purpleRef.current)
  const blackPos = calculatePosition(blackRef.current)
  const orangePos = calculatePosition(orangeRef.current)
  const yellowPos = calculatePosition(yellowRef.current)
  const isShowingPassword =
    sceneState.activePasswordLength > 0 && sceneState.activePasswordVisible
  const isLookingAway =
    sceneState.isPasswordFocused && !sceneState.activePasswordVisible

  const purplePupil = calculatePupilOffset(purpleEyeRef.current, 5)
  const blackPupil = calculatePupilOffset(blackEyeRef.current, 4)
  const orangePupil = calculatePupilOffset(orangePupilRef.current, 5)
  const yellowPupil = calculatePupilOffset(yellowPupilRef.current, 5)

  let purpleTransform = `skewX(${purplePos.bodySkew}deg)`
  let purpleHeight = 370
  let purpleEyes = {
    left: 45 + purplePos.faceX,
    top: 40 + purplePos.faceY,
    pupilX: purplePupil.x,
    pupilY: purplePupil.y,
  }

  if (isShowingPassword) {
    purpleTransform = 'skewX(0deg)'
    purpleEyes = {
      left: 20,
      top: 35,
      pupilX: sceneState.isPurplePeeking ? 4 : -4,
      pupilY: sceneState.isPurplePeeking ? 5 : -4,
    }
  } else if (isLookingAway) {
    purpleTransform = 'skewX(-14deg) translateX(-20px)'
    purpleHeight = 410
    purpleEyes = { left: 20, top: 25, pupilX: -5, pupilY: -5 }
  } else if (sceneState.isTyping) {
    purpleTransform = `skewX(${purplePos.bodySkew - 12}deg) translateX(40px)`
    purpleHeight = 410
  }

  if (!isShowingPassword && !isLookingAway && sceneState.isLookingAtEachOther) {
    purpleEyes = { left: 55, top: 65, pupilX: 3, pupilY: 4 }
  }

  let blackTransform = `skewX(${blackPos.bodySkew}deg)`
  let blackEyes = {
    left: 26 + blackPos.faceX,
    top: 32 + blackPos.faceY,
    pupilX: blackPupil.x,
    pupilY: blackPupil.y,
  }

  if (isShowingPassword) {
    blackTransform = 'skewX(0deg)'
    blackEyes = { left: 10, top: 28, pupilX: -4, pupilY: -4 }
  } else if (isLookingAway) {
    blackTransform = 'skewX(12deg) translateX(-10px)'
    blackEyes = { left: 10, top: 20, pupilX: -4, pupilY: -5 }
  } else if (sceneState.isLookingAtEachOther) {
    blackTransform = `skewX(${blackPos.bodySkew * 1.5 + 10}deg) translateX(20px)`
    blackEyes = { left: 32, top: 12, pupilX: 0, pupilY: -4 }
  } else if (sceneState.isTyping) {
    blackTransform = `skewX(${blackPos.bodySkew * 1.5}deg)`
  }

  let orangeEyes = {
    left: 82 + orangePos.faceX,
    top: 90 + orangePos.faceY,
    pupilX: orangePupil.x,
    pupilY: orangePupil.y,
  }
  if (isLookingAway) {
    orangeEyes = { left: 50, top: 75, pupilX: -5, pupilY: -5 }
  } else if (isShowingPassword) {
    orangeEyes = { left: 50, top: 85, pupilX: -5, pupilY: -4 }
  }

  let yellowEyes = {
    left: 52 + yellowPos.faceX,
    top: 40 + yellowPos.faceY,
    pupilX: yellowPupil.x,
    pupilY: yellowPupil.y,
  }
  let yellowMouth = {
    left: 40 + yellowPos.faceX,
    top: 88 + yellowPos.faceY,
  }
  if (isLookingAway) {
    yellowEyes = { left: 20, top: 30, pupilX: -5, pupilY: -5 }
    yellowMouth = { left: 15, top: 78 }
  } else if (isShowingPassword) {
    yellowEyes = { left: 20, top: 35, pupilX: -5, pupilY: -4 }
    yellowMouth = { left: 10, top: 88 }
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
          style={{ height: purpleHeight, transform: purpleTransform }}
        >
          <div
            className='auth-eyes'
            style={{ left: purpleEyes.left, top: purpleEyes.top, gap: 28 }}
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
          style={{ transform: blackTransform }}
        >
          <div
            className='auth-eyes'
            style={{ left: blackEyes.left, top: blackEyes.top, gap: 20 }}
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
            style={{ left: orangeEyes.left, top: orangeEyes.top, gap: 28 }}
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
            style={{ left: yellowEyes.left, top: yellowEyes.top, gap: 20 }}
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
            style={{ left: yellowMouth.left, top: yellowMouth.top }}
          />
        </div>
      </div>
    </div>
  )
}
