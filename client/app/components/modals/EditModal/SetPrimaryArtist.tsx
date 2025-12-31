import { useQuery } from "@tanstack/react-query";
import { getAlbum, type Artist } from "api/api";
import { useEffect, useState } from "react";

interface Props {
  id: number;
  type: string;
}

export default function SetPrimaryArtist({ id, type }: Props) {
  const [err, setErr] = useState("");
  const [primary, setPrimary] = useState<Artist>();
  const [success, setSuccess] = useState("");

  const { isPending, isError, data, error } = useQuery({
    queryKey: [
      "get-artists-" + type.toLowerCase(),
      {
        id: id,
      },
    ],
    queryFn: () => {
      return fetch(
        "/apis/web/v1/artists?" + type.toLowerCase() + "_id=" + id
      ).then((r) => r.json()) as Promise<Artist[]>;
    },
  });

  useEffect(() => {
    if (data) {
      for (let a of data) {
        if (a.is_primary) {
          setPrimary(a);
          break;
        }
      }
    }
  }, [data]);

  if (isError) {
    return <p className="error">Error: {error.message}</p>;
  }
  if (isPending) {
    return <p>Loading...</p>;
  }

  const updatePrimary = (artist: number, val: boolean) => {
    setErr("");
    setSuccess("");
    fetch(
      `/apis/web/v1/artists/primary?artist_id=${artist}&${type.toLowerCase()}_id=${id}&is_primary=${val}`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/x-www-form-urlencoded",
        },
      }
    ).then((r) => {
      if (r.ok) {
        setSuccess("successfully updated primary artists");
      } else {
        r.json().then((r) => setErr(r.error));
      }
    });
  };

  return (
    <div className="w-full">
      <h3>Set Primary Artist</h3>
      <div className="flex flex-col gap-4">
        <select
          name="mark-various-artists"
          id="mark-various-artists"
          className="w-60 px-3 py-2 rounded-md"
          value={primary?.name || ""}
          onChange={(e) => {
            for (let a of data) {
              if (a.name === e.target.value) {
                setPrimary(a);
                updatePrimary(a.id, true);
              }
            }
          }}
        >
          <option value="" disabled>
            Select an artist
          </option>
          {data.map((a) => (
            <option key={a.id} value={a.name}>
              {a.name}
            </option>
          ))}
        </select>
        {err && <p className="error">{err}</p>}
        {success && <p className="success">{success}</p>}
      </div>
    </div>
  );
}
