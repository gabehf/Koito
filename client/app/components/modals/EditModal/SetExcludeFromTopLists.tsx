import { useQuery } from "@tanstack/react-query";
import {
  getAlbum,
  getArtist,
  updateExcludeFromTopLists,
  type Album,
  type Artist,
} from "api/api";
import { useEffect, useState } from "react";
import SubHeader from "~/components/primitives/SubHeader";

interface Props {
  type: string;
  id: number;
}

export default function SetExcludeFromTopLists({ type, id }: Props) {
  const [err, setErr] = useState("");
  const [excluded, setExcluded] = useState(false);
  const [success, setSuccess] = useState("");

  const { isPending, isError, data, error } = useQuery<Album | Artist>({
    queryKey: [type === "artist" ? "get-artist" : "get-album", id],
    queryFn: () => (type === "artist" ? getArtist(id) : getAlbum(id)),
  });

  useEffect(() => {
    if (data) {
      setExcluded(data.exclude_from_top_lists);
    }
  }, [data]);

  if (isError) {
    return <p className="error">Error: {error.message}</p>;
  }
  if (isPending) {
    return <p>Loading...</p>;
  }

  const updateExcluded = (val: boolean) => {
    setErr("");
    setSuccess("");
    updateExcludeFromTopLists(type, id, val).then((r) => {
      if (r.ok) {
        setSuccess(`Successfully updated ${type}`);
      } else {
        r.json().then((r) => setErr(r.error));
      }
    });
  };

  return (
    <div className="flex-grow">
      <SubHeader>Exclude from Top Lists</SubHeader>
      <div className="flex flex-col gap-4">
        <select
          name="exclude-from-top-lists"
          id="exclude-from-top-lists"
          className="w-30 px-3 py-2 rounded-md"
          value={excluded.toString()}
          onChange={(e) => {
            const val = e.target.value === "true";
            setExcluded(val);
            updateExcluded(val);
          }}
        >
          <option value="true">True</option>
          <option value="false">False</option>
        </select>
        {err && <p className="error">{err}</p>}
        {success && <p className="success">{success}</p>}
      </div>
    </div>
  );
}
