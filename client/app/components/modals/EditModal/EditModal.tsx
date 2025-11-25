import { useQuery } from "@tanstack/react-query";
import {
  createAlias,
  deleteAlias,
  getAliases,
  setPrimaryAlias,
  type Alias,
} from "api/api";
import { Modal } from "../Modal";
import { AsyncButton } from "../../AsyncButton";
import { useEffect, useState } from "react";
import { Trash } from "lucide-react";
import SetVariousArtists from "./SetVariousArtist";
import SetPrimaryArtist from "./SetPrimaryArtist";

interface Props {
  type: string;
  id: number;
  open: boolean;
  setOpen: Function;
}

export default function EditModal({ open, setOpen, type, id }: Props) {
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const [err, setError] = useState<string>();
  const [displayData, setDisplayData] = useState<Alias[]>([]);

  const { isPending, isError, data, error } = useQuery({
    queryKey: [
      "aliases",
      {
        type: type,
        id: id,
      },
    ],
    queryFn: ({ queryKey }) => {
      const params = queryKey[1] as { type: string; id: number };
      return getAliases(params.type, params.id);
    },
  });

  useEffect(() => {
    if (data) {
      setDisplayData(data);
    }
  }, [data]);

  if (isError) {
    return <p className="error">Error: {error.message}</p>;
  }
  if (isPending) {
    return <p>Loading...</p>;
  }

  const handleSetPrimary = (alias: string) => {
    setError(undefined);
    setLoading(true);
    setPrimaryAlias(type, id, alias).then((r) => {
      if (r.ok) {
        window.location.reload();
      } else {
        r.json().then((r) => setError(r.error));
      }
    });
    setLoading(false);
  };

  const handleNewAlias = () => {
    setError(undefined);
    if (input === "") {
      setError("alias must be provided");
      return;
    }
    setLoading(true);
    createAlias(type, id, input).then((r) => {
      if (r.ok) {
        setDisplayData([
          ...displayData,
          { alias: input, source: "Manual", is_primary: false, id: id },
        ]);
      } else {
        r.json().then((r) => setError(r.error));
      }
    });
    setLoading(false);
  };

  const handleDeleteAlias = (alias: string) => {
    setError(undefined);
    setLoading(true);
    deleteAlias(type, id, alias).then((r) => {
      if (r.ok) {
        setDisplayData(displayData.filter((v) => v.alias != alias));
      } else {
        r.json().then((r) => setError(r.error));
      }
    });
    setLoading(false);
  };

  const handleClose = () => {
    setOpen(false);
    setInput("");
  };

  return (
    <Modal maxW={1000} isOpen={open} onClose={handleClose}>
      <div className="flex flex-col items-start gap-6 w-full">
        <div className="w-full">
          <h2>Alias Manager</h2>
          <div className="flex flex-col gap-4">
            {displayData.map((v) => (
              <div className="flex gap-2">
                <div className="bg p-3 rounded-md flex-grow" key={v.alias}>
                  {v.alias} (source: {v.source})
                </div>
                <AsyncButton
                  loading={loading}
                  onClick={() => handleSetPrimary(v.alias)}
                  disabled={v.is_primary}
                >
                  Set Primary
                </AsyncButton>
                <AsyncButton
                  loading={loading}
                  onClick={() => handleDeleteAlias(v.alias)}
                  confirm
                  disabled={v.is_primary}
                >
                  <Trash size={16} />
                </AsyncButton>
              </div>
            ))}
            <div className="flex gap-2 w-3/5">
              <input
                type="text"
                placeholder="Add a new alias"
                className="mx-auto fg bg rounded-md p-3 flex-grow"
                value={input}
                onChange={(e) => setInput(e.target.value)}
              />
              <AsyncButton loading={loading} onClick={handleNewAlias}>
                Submit
              </AsyncButton>
            </div>
            {err && <p className="error">{err}</p>}
          </div>
        </div>
        {type.toLowerCase() === "album" && (
          <>
            <SetVariousArtists id={id} />
            <SetPrimaryArtist id={id} type="album" />
          </>
        )}
        {type.toLowerCase() === "track" && (
          <SetPrimaryArtist id={id} type="track" />
        )}
      </div>
    </Modal>
  );
}
