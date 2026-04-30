import { useEffect, useState } from "react";
import SidebarItem from "./SidebarItem";
import { Search } from "lucide-react";
import SearchModal from "../modals/SearchModal";

interface Props {
  size: number;
}

export default function SidebarSearch({ size }: Props) {
  const [open, setModalOpen] = useState(false);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const active = document.activeElement;
      const isTyping =
        active &&
        (active.tagName === "INPUT" ||
          active.tagName === "TEXTAREA" ||
          (active as HTMLElement).isContentEditable);

      if (!isTyping && e.key === "/") {
        e.preventDefault();
        setModalOpen(!open);
      }
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [open]);

  return (
    <SidebarItem
      space={26}
      onClick={() => setModalOpen(true)}
      name="Search"
      keyHint="/"
      children={<Search size={size} />}
      modal={<SearchModal open={open} setOpen={setModalOpen} />}
    />
  );
}
