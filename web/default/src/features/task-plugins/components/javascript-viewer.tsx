import { javascript } from '@codemirror/lang-javascript'
import { EditorState } from '@codemirror/state'
import { EditorView, lineNumbers } from '@codemirror/view'
import { useEffect, useRef } from 'react'

type JavaScriptViewerProps = {
  value: string
  className?: string
}

export function JavaScriptViewer(props: JavaScriptViewerProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const viewRef = useRef<EditorView | null>(null)

  useEffect(() => {
    if (!containerRef.current) return
    const view = new EditorView({
      parent: containerRef.current,
      state: EditorState.create({
        doc: props.value,
        extensions: [
          lineNumbers(),
          javascript(),
          EditorState.readOnly.of(true),
          EditorView.editable.of(false),
          EditorView.lineWrapping,
          EditorView.theme({
            '&': { height: '100%', backgroundColor: 'transparent' },
            '.cm-scroller': { overflow: 'auto', fontFamily: 'monospace' },
          }),
        ],
      }),
    })
    viewRef.current = view
    return () => {
      view.destroy()
      viewRef.current = null
    }
  }, [props.value])

  return <div ref={containerRef} className={props.className} />
}
