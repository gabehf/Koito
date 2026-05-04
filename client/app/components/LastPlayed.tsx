import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { timeSince } from "~/utils/utils";
import ArtistLinks from "./ArtistLinks";
import Image from "./primitives/Image";
import {
  apiFetch,
  deleteListen,
  imageUrl,
  type Listen,
  type NowPlaying,
  type PaginatedResponse,
} from "api/api";
import { Link } from "react-router";
import { useAppContext } from "~/providers/AppProvider";

interface Props {
  limit: number;
  artistId?: Number;
  albumId?: Number;
  trackId?: number;
  hideArtists?: boolean;
  showNowPlaying?: boolean;
}

const getLastListens = (args: {
  limit: number;
  period: string;
  artist_id?: number;
  album_id?: number;
  track_id?: number;
  page: number;
}) => apiFetch<PaginatedResponse<Listen>>("/apis/web/v1/listens", args);

const getNowPlaying = () => apiFetch<NowPlaying>("/apis/web/v1/now-playing");

export default function LastPlays(props: Props) {
  const { user } = useAppContext();
  const args = {
    limit: props.limit,
    period: "all_time",
    artist_id: props.artistId as number | undefined,
    album_id: props.albumId as number | undefined,
    track_id: props.trackId,
    page: 0,
  };
  const { isPending, isError, data, error } = useQuery({
    queryKey: ["last-listens", args],
    queryFn: () => getLastListens(args),
  });
  const { data: npData } = useQuery({
    queryKey: ["now-playing"],
    queryFn: () => getNowPlaying(),
  });

  const header = "Last played";

  const [items, setItems] = useState<Listen[] | null>(null);

  const handleDelete = async (listen: Listen) => {
    if (!data) return;
    try {
      const res = await deleteListen(listen);
      if (res.ok || (res.status >= 200 && res.status < 300)) {
        setItems((prev) =>
          (prev ?? data.items).filter((i) => i.time !== listen.time)
        );
      } else {
        console.error("Failed to delete listen:", res.status);
      }
    } catch (err) {
      console.error("Error deleting listen:", err);
    }
  };

  if (isPending) {
    return (
      <div className="w-[300px] sm:w-[500px]">
        <h3>{header}</h3>
        <p>Loading...</p>
      </div>
    );
  } else if (isError) {
    return (
      <div className="w-[300px] sm:w-[500px]">
        <h3>{header}</h3>
        <p className="error">Error: {error.message}</p>
      </div>
    );
  }

  const listens = items ?? data.items;

  let params = "";
  params += props.artistId ? `&artist_id=${props.artistId}` : "";
  params += props.albumId ? `&album_id=${props.albumId}` : "";
  params += props.trackId ? `&track_id=${props.trackId}` : "";

  return (
    <div className="text-sm sm:text-[15px]">
      <h3 className="hover:underline">
        <Link to={`/listens?period=all_time${params}`}>{header}</Link>
      </h3>
      {listens.length < 1 && "Nothing to show"}
      <table className="table-fixed border-collapse mt-10">
        <tbody>
          {props.showNowPlaying && npData && npData.currently_playing && (
            <tr className="group hover:bg-[--color-bg-secondary]">
              <td>
                <Link
                  className="hover:text-[--color-fg-secondary]"
                  to={`/track/${npData.track.id}`}
                >
                  {npData.track.title}
                </Link>
              </td>
              <td className="text-ellipsis overflow-hidden text-center sm:max-w-[600px]">
                {props.hideArtists ? null : (
                  <>
                    <ArtistLinks artists={npData.track.artists} /> –{" "}
                  </>
                )}
              </td>
              <td className="color-fg-tertiary pr-2 sm:pr-4 text-sm whitespace-nowrap w-0">
                Now Playing
              </td>
            </tr>
          )}
          {listens.map((item) => (
            <tr
              key={`last_listen_${item.time}`}
              className="group border-b-1 border-(--color-bg-tertiary) relative last:border-b-0"
            >
              <td className="py-2 pr-3">
                <Link to={`/track/${item.track.id}`}>
                  <Image
                    imageUrl={imageUrl(item.track.image, "small")}
                    size={32}
                  />
                </Link>
              </td>
              <td className="min-w-[150px]">
                <Link
                  className="hover:text-[--color-fg-secondary]"
                  to={`/track/${item.track.id}`}
                >
                  {item.track.title}
                </Link>
              </td>
              <td className="text-ellipsis overflow-hidden text-center min-w-[150px]">
                {props.hideArtists ? null : (
                  <>
                    <ArtistLinks artists={item.track.artists} />
                  </>
                )}
              </td>
              <td
                className="color-fg-tertiary pr-2 sm:pr-4 text-sm text-end whitespace-nowrap w-0 min-w-[150px]"
                title={new Date(item.time).toString()}
              >
                <p className="-mr-[18px]">{timeSince(new Date(item.time))}</p>
              </td>
              <td className="pr-2 align-middle">
                <button
                  onClick={() => handleDelete(item)}
                  className="absolute top-3.5 -right-5 opacity-0 group-hover:opacity-100 transition-opacity text-(--color-fg-tertiary) hover:text-(--color-error)"
                  aria-label="Delete"
                  hidden={user === null || user === undefined}
                >
                  ×
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
