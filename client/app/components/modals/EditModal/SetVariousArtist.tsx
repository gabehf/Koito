import { useProfile } from "~/providers/ProfileProvider";
import { useQuery } from "@tanstack/react-query";

import { useEffect, useState } from "react";
import SubHeader from "~/components/primitives/SubHeader";

interface Props {
  id: number;
}

export default function SetVariousArtists({ id }: Props) {
  const { profile, api } = useProfile();

  const [err, setErr] = useState("");
  const [va, setVA] = useState(false);
  const [success, setSuccess] = useState("");

  const { isPending, isError, data, error } = useQuery({
    queryKey: ["profile", profile.id, "get-album", id],
    queryFn: () => api.getAlbum(id),
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
    api
      .fetch(`/apis/web/v1/album/${id}`, {
        method: "PATCH",
        body: JSON.stringify({ is_various_artists: val }),
        headers: {
          "Content-Type": "application/json",
        },
      })
      .then((r) => {
        if (r.ok) {
          setSuccess("Successfully updated album");
        } else {
          r.json().then((r) => setErr(r.error));
        }
      });
  };

  return (
    <div className="w-full">
      <SubHeader>Mark as Various Artists</SubHeader>
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
