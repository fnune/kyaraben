import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { INPUT_BASE_CLASSES } from './inputStyles'

export interface SelectOption {
  value: string
  label: string
}

export interface SelectProps {
  readonly value: string
  readonly options: readonly SelectOption[]
  readonly onChange: (value: string) => void
  readonly disabled?: boolean
  readonly className?: string
  readonly size?: 'sm' | 'md'
  readonly error?: boolean
}

export function Select({
  value,
  options,
  onChange,
  disabled = false,
  className = '',
  size = 'md',
  error = false,
}: SelectProps) {
  const [open, setOpen] = useState(false)
  const [focusedIndex, setFocusedIndex] = useState(-1)
  const [dropdownPosition, setDropdownPosition] = useState({ top: 0, left: 0, width: 0 })
  const containerRef = useRef<HTMLDivElement>(null)
  const buttonRef = useRef<HTMLButtonElement>(null)
  const listboxRef = useRef<HTMLDivElement>(null)

  const selectedOption = options.find((opt) => opt.value === value)
  const selectedIndex = options.findIndex((opt) => opt.value === value)

  useEffect(() => {
    if (open) {
      setFocusedIndex(selectedIndex >= 0 ? selectedIndex : 0)
    }
  }, [open, selectedIndex])

  useLayoutEffect(() => {
    if (open && buttonRef.current) {
      const rect = buttonRef.current.getBoundingClientRect()
      setDropdownPosition({
        top: rect.bottom + 4,
        left: rect.left,
        width: rect.width,
      })
    }
  }, [open])

  useEffect(() => {
    if (!open) return

    function handleClickOutside(e: MouseEvent) {
      if (
        containerRef.current &&
        !containerRef.current.contains(e.target as Node) &&
        listboxRef.current &&
        !listboxRef.current.contains(e.target as Node)
      ) {
        setOpen(false)
      }
    }

    function handleScroll(e: Event) {
      if (listboxRef.current?.contains(e.target as Node)) return
      setOpen(false)
    }

    document.addEventListener('mousedown', handleClickOutside)
    document.addEventListener('scroll', handleScroll, true)
    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
      document.removeEventListener('scroll', handleScroll, true)
    }
  }, [open])

  useEffect(() => {
    if (open && focusedIndex >= 0 && listboxRef.current) {
      const focusedElement = listboxRef.current.children[focusedIndex] as HTMLElement
      focusedElement?.scrollIntoView({ block: 'nearest' })
    }
  }, [open, focusedIndex])

  function handleKeyDown(e: React.KeyboardEvent) {
    if (disabled) return

    switch (e.key) {
      case 'Enter':
      case ' ': {
        e.preventDefault()
        const focusedOption = options[focusedIndex]
        if (open && focusedOption) {
          onChange(focusedOption.value)
          setOpen(false)
        } else {
          setOpen(true)
        }
        break
      }
      case 'ArrowDown':
        e.preventDefault()
        if (!open) {
          setOpen(true)
        } else {
          setFocusedIndex((prev) => Math.min(prev + 1, options.length - 1))
        }
        break
      case 'ArrowUp':
        e.preventDefault()
        if (!open) {
          setOpen(true)
        } else {
          setFocusedIndex((prev) => Math.max(prev - 1, 0))
        }
        break
      case 'Escape':
        e.preventDefault()
        setOpen(false)
        break
      case 'Home':
        if (open) {
          e.preventDefault()
          setFocusedIndex(0)
        }
        break
      case 'End':
        if (open) {
          e.preventDefault()
          setFocusedIndex(options.length - 1)
        }
        break
    }
  }

  function handleOptionClick(optionValue: string) {
    onChange(optionValue)
    setOpen(false)
  }

  const buttonId = useRef(`select-button-${Math.random().toString(36).slice(2)}`).current
  const listboxId = useRef(`select-listbox-${Math.random().toString(36).slice(2)}`).current

  const dropdown = open
    ? createPortal(
        <div
          id={listboxId}
          ref={listboxRef}
          role="listbox"
          tabIndex={0}
          aria-activedescendant={
            focusedIndex >= 0 ? `${listboxId}-option-${focusedIndex}` : undefined
          }
          style={{
            position: 'fixed',
            top: dropdownPosition.top,
            left: dropdownPosition.left,
            minWidth: dropdownPosition.width,
          }}
          className={`z-50 max-h-48 overflow-auto py-0.5 ${INPUT_BASE_CLASSES}`}
        >
          {options.map((option, index) => (
            <div
              key={option.value}
              id={`${listboxId}-option-${index}`}
              role="option"
              tabIndex={-1}
              aria-selected={option.value === value}
              onClick={() => handleOptionClick(option.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  e.preventDefault()
                  handleOptionClick(option.value)
                }
              }}
              onMouseEnter={() => setFocusedIndex(index)}
              className={`
                font-mono tabular-nums cursor-pointer
                ${size === 'sm' ? 'px-2 py-1 text-xs' : 'px-3 py-2 text-sm'}
                ${option.value === value ? 'text-accent' : 'text-on-surface-secondary'}
                ${focusedIndex === index ? 'bg-accent/10' : ''}
              `}
            >
              {option.label}
            </div>
          ))}
        </div>,
        document.body,
      )
    : null

  return (
    <div ref={containerRef} className={`relative inline-block ${className}`}>
      <button
        ref={buttonRef}
        id={buttonId}
        type="button"
        role="combobox"
        aria-haspopup="listbox"
        aria-expanded={open}
        aria-controls={listboxId}
        disabled={disabled}
        onClick={() => !disabled && setOpen(!open)}
        onKeyDown={handleKeyDown}
        className={`
          flex items-center gap-1.5 tabular-nums font-mono text-on-surface-secondary
          ${size === 'sm' ? 'px-2 py-1 text-xs' : 'px-3 py-2 text-sm'}
          ${INPUT_BASE_CLASSES}
          ${error ? 'border-status-error' : ''}
          ${disabled ? 'opacity-40 cursor-not-allowed' : 'cursor-pointer'}
        `}
      >
        <span className="truncate">{selectedOption?.label ?? ''}</span>
        <svg
          aria-hidden="true"
          className={`w-3 h-3 shrink-0 transition-transform ${open ? 'rotate-180' : ''}`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {dropdown}
    </div>
  )
}
