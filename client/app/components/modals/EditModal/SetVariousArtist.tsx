import { useQuery } from "@tanstack/react-query";
import { getAlbum } from "api/api";
import { useEffect, useState } from "react";

interface Props {
  id: number;
}

export default function SetVariousArtists({ id }: Props) {
  const [err, setErr] = useState("");
  const [va, setVA] = useState(false);
  const [success, setSuccess] = useState("");

  const { isPending, isError, data, error } = useQuery({
    queryKey: [
      "get-album",
      {
        id: id,
      },
    ],
    queryFn: ({ queryKey }) => {
      const params = queryKey[1] as { id: number };
      return getAlbum(params.id);
    },
  });

  useEffect(() => {
    if (data) {
      setVA(data.is_various_artists);
    }
  }, [data]);

  if (isError) {
    return <p className="error">Error: {error.message}</p>;
  }
  if (isPending) {
    return <p>Loading...</p>;
  }

  const updateVA = (val: boolean) => {
    setErr("");
    setSuccess("");
    fetch(`/apis/web/v1/album?id=${id}&is_various_artists=${val}`, {
      method: "PATCH",
    }).then((r) => {
      if (r.ok) {
        setSuccess("Successfully updated album");
      } else {
        r.json().then((r) => setErr(r.error));
      }
    });
  };

  return (
    <div className="w-full">
      <h3>Mark as Various Artists</h3>
      <div className="flex flex-col gap-4">
        <select
          name="mark-various-artists"
          id="mark-various-artists"
          className="w-30 px-3 py-2 rounded-md"
          value={va.toString()}
          onChange={(e) => {
            const val = e.target.value === "true";
            setVA(val);
            updateVA(val);
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
