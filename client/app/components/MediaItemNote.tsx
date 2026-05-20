import { useState } from "react";
import { Modal } from "./modals/Modal";

interface Props {
  children: React.ReactNode;
}

export default function MediaItemNote({ children }: Props) {
  const [open, setOpen] = useState(false);
  return (
    <>
      <Modal isOpen={open} onClose={() => setOpen(false)}>
        <h3 className="mb-3">Note</h3>
        {children}
      </Modal>
      <div
        className="p-4 m-2 border-l border-(--color-fg-tertiary) max-w-[400px] hover:cursor-pointer select-none"
        onClick={() => setOpen(true)}
      >
        <div className="line-clamp-4">"{children}"</div>
      </div>
    </>
  );
}
