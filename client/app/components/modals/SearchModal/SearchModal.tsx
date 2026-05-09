import { useDeferredValue, useEffect, useRef, useState } from "react";
import ReactDOM from "react-dom";
import { search, type SearchResponse } from "api/api";
import SearchResults from "./SearchResults";
import SubHeader from "../../primitives/SubHeader";

interface Props {
  open: boolean;
  setOpen: Function;
}

export default function SearchModal({ open, setOpen }: Props) {
  const [query, setQuery] = useState("");
  const [data, setData] = useState<SearchResponse>();
  const deferredQuery = useDeferredValue(query);

  const modalRef = useRef<HTMLDivElement>(null);
  const [shouldRender, setShouldRender] = useState(open);
  const [isClosing, setIsClosing] = useState(false);

  const closeSearchModal = () => {
    setOpen(false);
    setQuery("");
    setData(undefined);
  };

  useEffect(() => {
    if (deferredQuery) {
      search(deferredQuery).then((r) => setData(r));
    }
  }, [deferredQuery]);

  useEffect(() => {
    if (open) {
      setShouldRender(true);
      setIsClosing(false);
    } else if (shouldRender) {
      setIsClosing(true);
      const timeout = setTimeout(() => setShouldRender(false), 100);
      return () => clearTimeout(timeout);
    }
  }, [open, shouldRender]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        closeSearchModal();
      } else if (e.key === "Tab") {
        if (modalRef.current) {
          const focusableEls = modalRef.current.querySelectorAll<HTMLElement>(
            'button:not(:disabled), [href], input:not(:disabled), select:not(:disabled), textarea:not(:disabled), [tabindex]:not([tabindex="-1"])'
          );
          const firstEl = focusableEls[0];
          const lastEl = focusableEls[focusableEls.length - 1];
          const activeEl = document.activeElement;
          if (e.shiftKey && activeEl === firstEl) {
            e.preventDefault();
            lastEl.focus();
          } else if (!e.shiftKey && activeEl === lastEl) {
            e.preventDefault();
            firstEl.focus();
          } else if (
            !Array.from(focusableEls).find((node) => node.isEqualNode(activeEl))
          ) {
            e.preventDefault();
            firstEl.focus();
          }
        }
      }
    };
    if (open) document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [open]);

  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (modalRef.current && !modalRef.current.contains(e.target as Node)) {
        closeSearchModal();
      }
    };
    if (open) document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, [open]);

  if (!shouldRender) return null;

  return ReactDOM.createPortal(
    <div
      className={`fixed inset-0 z-50 flex flex-col items-center justify-start bg-black/80 backdrop-blur-sm transition-opacity duration-100 ${
        isClosing ? "animate-fade-out" : "animate-fade-in"
      }`}
    >
      <div
        ref={modalRef}
        className="flex flex-col w-full overflow-y-auto max-h-screen py-16 px-7"
        style={{ maxWidth: 600 }}
      >
        <SubHeader isOffset>Search</SubHeader>
        <div
          className={`border bg-secondary rounded-(--border-radius) p-6 w-full relative transition-all duration-100 ${
            isClosing ? "animate-fade-out-scale" : "animate-fade-in-scale"
          }`}
        >
          <input
            type="text"
            autoFocus
            placeholder="Search for an artist, album, or track"
            className="w-full mx-auto fg bg rounded p-2"
            onChange={(e) => setQuery(e.target.value)}
          />
          <button
            onClick={closeSearchModal}
            className="absolute text-xl top-1 right-2 sm:top-3 sm:right-4 text-(--color-fg-tertiary) hover:text-(--color-fg) hover:cursor-pointer"
          >
            ×
          </button>
        </div>
        <SearchResults data={data} onSelect={closeSearchModal} />
      </div>
    </div>,
    document.body
  );
}
