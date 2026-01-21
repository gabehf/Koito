import { useEffect, useRef, useState } from 'react';
import ReactDOM from 'react-dom';

export function Modal({
  isOpen,
  onClose,
  children,
  maxW,
  h
}: {
  isOpen: boolean;
  onClose: () => void;
  children: React.ReactNode;
  maxW?: number;
  h?: number;
}) {
  const modalRef = useRef<HTMLDivElement>(null);
  const [shouldRender, setShouldRender] = useState(isOpen);
  const [isClosing, setIsClosing] = useState(false);

  // Show/hide logic
  useEffect(() => {
    if (isOpen) {
      setShouldRender(true);
      setIsClosing(false);
    } else if (shouldRender) {
      setIsClosing(true);
      const timeout = setTimeout(() => {
        setShouldRender(false);
      }, 100); // Match fade-out duration
      return () => clearTimeout(timeout);
    }
  }, [isOpen, shouldRender]);

  // Handle keyboard events
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Close on Escape key
      if (e.key === 'Escape') {
        onClose()
      // Trap tab navigation to the modal
      } else if (e.key === 'Tab') {
        if (modalRef.current) {
          const focusableEls = modalRef.current.querySelectorAll<HTMLElement>(
            'button:not(:disabled), [href], input:not(:disabled), select:not(:disabled), textarea:not(:disabled), [tabindex]:not([tabindex="-1"])'
          );
          const firstEl = focusableEls[0];
          const lastEl = focusableEls[focusableEls.length - 1];
          const activeEl = document.activeElement

          if (e.shiftKey && activeEl === firstEl) {
            e.preventDefault();
            lastEl.focus();
          } else if (!e.shiftKey && activeEl === lastEl) {
            e.preventDefault();
            firstEl.focus();
          } else if (!Array.from(focusableEls).find(node => node.isEqualNode(activeEl))) {
            e.preventDefault();
            firstEl.focus();
          }
        }
      };
    };
    if (isOpen) document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, onClose]);

  // Close on outside click
  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (
        modalRef.current &&
        !modalRef.current.contains(e.target as Node)
      ) {
        onClose();
      }
    };
    if (isOpen) document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, [isOpen, onClose]);

  if (!shouldRender) return null;

  return ReactDOM.createPortal(
    <div
      className={`fixed inset-0 z-50 flex items-center justify-center bg-black/50 transition-opacity duration-100 ${
        isClosing ? 'animate-fade-out' : 'animate-fade-in'
      }`}
    >
      <div
        ref={modalRef}
        className={`bg-secondary rounded-lg shadow-md p-6 w-full relative max-h-3/4 overflow-y-auto transition-all duration-100 ${
          isClosing ? 'animate-fade-out-scale' : 'animate-fade-in-scale'
        }`}
        style={{ maxWidth: maxW ?? 600, height: h ?? '' }}
      >
        {children}
        <button
          onClick={onClose}
          className="absolute top-2 right-2 color-fg-tertiary hover:cursor-pointer"
        >
          🞪
        </button>
      </div>
    </div>,
    document.body
  );
}
