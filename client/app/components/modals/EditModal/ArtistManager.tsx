import { useQuery, useQueryClient } from "@tanstack/react-query";
import { type Artist, type SearchResponse } from "api/api";
import { Trash } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { AsyncButton } from "../../AsyncButton";
import ComboBox from "~/components/ComboBox";
import SubHeader from "~/components/primitives/SubHeader";

interface Props {
  type: string;
  id: number;
}

export default function ArtistManager({ type, id }: Props) {
  const [loading, setLoading] = useState(false);
  const [err, setError] = useState<string>();
  const [displayData, setDisplayData] = useState<Artist[]>([]);
  const [addArtistTarget, setAddArtistTarget] = useState<Artist>();
  const queryClient = useQueryClient();

  const { isPending, isError, data, error } = useQuery({
    queryKey: ["get-artists-" + type.toLowerCase(), { id: id }],
    queryFn: () => {
      return fetch(`/apis/web/v1/${type.toLowerCase()}/${id}/artists`).then(
        (r) => r.json(),
      ) as Promise<Artist[]>;
    },
  });

  const handleSelectArtist = useCallback(
    (artist: Artist) => {
      setAddArtistTarget(artist);
    },
    [type, id],
  );

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

  const handleSetPrimary = (artist: Artist, val: boolean) => {
    setError(undefined);
    setLoading(true);
    fetch(`/apis/web/v1/${type.toLowerCase()}/${id}/artists/${artist.id}`, {
      method: "PATCH",
      body: JSON.stringify({ is_primary: val }),
      headers: {
        "Content-Type": "application/json",
      },
    }).then(async (r) => {
      if (r.ok) {
        await queryClient.invalidateQueries({
          queryKey: ["get-artists-" + type.toLowerCase(), { id: id }],
        });
      } else {
        r.json().then((r) => setError(r.error));
      }
    });
    setLoading(false);
  };

  const handleAddArtist = () => {
    setError(undefined);
    if (!addArtistTarget) {
      setError("no artist selected");
      return;
    }
    setLoading(true);
    fetch(`/apis/web/v1/${type}/${id}/artists`, {
      method: "POST",
      body: JSON.stringify({ artist_ids: [addArtistTarget.id] }),
      headers: {
        "Content-Type": "application/json",
      },
    }).then(async (r) => {
      if (r.ok) {
        await queryClient.invalidateQueries({
          queryKey: ["get-artists-" + type.toLowerCase(), { id: id }],
        });
      } else {
        r.json().then((r) => setError(r.error));
      }
    });
    setLoading(false);
  };

  const handleDeleteArtist = (artist: number) => {
    setError(undefined);
    setLoading(true);
    fetch(`/apis/web/v1/${type}/${id}/artists/${artist}`, {
      method: "DELETE",
    }).then(async (r) => {
      if (r.ok) {
        await queryClient.invalidateQueries({
          queryKey: ["get-artists-" + type.toLowerCase(), { id: id }],
        });
      } else {
        r.json().then((r) => setError(r.error));
      }
    });
    setLoading(false);
  };

  return (
    <div className="w-full">
      <SubHeader>Artist Manager</SubHeader>
      <div className="flex flex-col gap-4">
        {displayData.map((v) => (
          <div className="flex gap-2">
            <div className="bg p-3 rounded-md flex-grow" key={v.name}>
              {v.name}
            </div>
            <AsyncButton
              loading={loading}
              onClick={() => handleSetPrimary(v, true)}
              disabled={v.is_primary}
            >
              Set Primary
            </AsyncButton>
            {type == "track" && (
              <AsyncButton
                loading={loading}
                onClick={() => handleDeleteArtist(v.id)}
                confirm
                disabled={v.is_primary}
                danger
              >
                <Trash size={16} />
              </AsyncButton>
            )}
          </div>
        ))}
        {type == "track" && (
          <div className="flex gap-2 w-3/5">
            <ComboBox
              onSelection={handleSelectArtist}
              filterFunction={(r: SearchResponse) => {
                r.albums = [];
                r.tracks = [];
                const ids = displayData.map((d) => d.id);
                r.artists = r.artists.filter((a) => !ids.includes(a.id));
                return r;
              }}
            />
            <AsyncButton loading={loading} onClick={handleAddArtist}>
              Submit
            </AsyncButton>
          </div>
        )}
        {err && <p className="error">{err}</p>}
      </div>
    </div>
  );
}
