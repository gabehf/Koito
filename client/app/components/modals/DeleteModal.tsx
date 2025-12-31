import { deleteItem } from "api/api";
import { AsyncButton } from "../AsyncButton";
import { Modal } from "./Modal";
import { useNavigate } from "react-router";
import { useState } from "react";

interface Props {
  open: boolean;
  setOpen: Function;
  title: string;
  id: number;
  type: string;
}

export default function DeleteModal({ open, setOpen, title, id, type }: Props) {
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const doDelete = () => {
    setLoading(true);
    deleteItem(type.toLowerCase(), id).then((r) => {
      if (r.ok) {
        navigate("/");
      } else {
        console.log(r);
      }
    });
  };

  return (
    <Modal isOpen={open} onClose={() => setOpen(false)}>
      <h3>Delete "{title}"?</h3>
      <p>This action is irreversible!</p>
      <div className="flex flex-col mt-3 items-center">
        <AsyncButton loading={loading} onClick={doDelete}>
          Yes, Delete It
        </AsyncButton>
      </div>
    </Modal>
  );
}
