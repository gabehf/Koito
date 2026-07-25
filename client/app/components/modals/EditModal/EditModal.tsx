import { Modal } from "../Modal";
import AliasManager from "./AliasManager";
import SetVariousArtists from "./SetVariousArtist";
import SetExcludeFromTopLists from "./SetExcludeFromTopLists";
import UpdateMbzID from "./UpdateMbzID";
import ArtistManager from "./ArtistManager";

interface Props {
  type: string;
  id: number;
  open: boolean;
  setOpen: Function;
}

export default function EditModal({ open, setOpen, type, id }: Props) {
  type = type.toLowerCase();

  const handleClose = () => {
    setOpen(false);
  };

  return (
    <Modal maxW={1000} isOpen={open} onClose={handleClose}>
      <div className="flex flex-col items-start gap-6 w-full">
        <AliasManager id={id} type={type} />
        <div className="w-full flex flex-row">
          {type === "album" && (
            <>
              <SetVariousArtists id={id} />
            </>
          )}
          {(type === "album" || type === "artist") && (
            <SetExcludeFromTopLists id={id} type={type} />
          )}
        </div>
        {type !== "artist" && <ArtistManager id={id} type={type} />}
        <UpdateMbzID type={type} id={id} />
      </div>
    </Modal>
  );
}
