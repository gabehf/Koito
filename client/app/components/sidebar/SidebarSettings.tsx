import { Settings2 } from "lucide-react";
import SettingsModal from "../modals/SettingsModal";
import SidebarItem from "./SidebarItem";
import { useEffect, useState } from "react";

interface Props {
  size: number;
}

export default function SidebarSettings({ size }: Props) {
  const [open, setOpen] = useState(false);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const active = document.activeElement;
      const isTyping =
        active &&
        (active.tagName === "INPUT" ||
          active.tagName === "TEXTAREA" ||
          (active as HTMLElement).isContentEditable);

      if (!isTyping && e.key === "\\") {
        e.preventDefault();
        setOpen(!open);
      }
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [open]);

  return (
    <SidebarItem
      space={30}
      keyHint="\"
      name="Settings"
      onClick={() => setOpen(true)}
      modal={<SettingsModal open={open} setOpen={setOpen} />}
    >
      <Settings2 size={size} />
    </SidebarItem>
  );
}
