import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { timeSince } from "~/utils/utils";
import ArtistLinks from "./ArtistLinks";
import {
  deleteListen,
  getLastListens,
  getNowPlaying,
  type getItemsArgs,
  type Listen,
  type Track,
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

export default function LastPlays(props: Props) {
  const { user } = useAppContext();
  const { isPending, isError, data, error } = useQuery({
    queryKey: [
      "last-listens",
      {
        limit: props.limit,
        period: "all_time",
        artist_id: props.artistId,
        album_id: props.albumId,
        track_id: props.trackId,
      },
    ],
    queryFn: ({ queryKey }) => getLastListens(queryKey[1] as getItemsArgs),
  });
  const { data: npData } = useQuery({
    queryKey: ["now-playing"],
    queryFn: () => getNowPlaying(),
  });

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
        <h2>Last Played</h2>
        <p>Loading...</p>
      </div>
    );
  } else if (isError) {
    return (
      <div className="w-[300px] sm:w-[500px]">
        <h2>Last Played</h2>
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
    <div className="text-sm sm:text-[16px]">
      <h2 className="hover:underline">
        <Link to={`/listens?period=all_time${params}`}>Last Played</Link>
      </h2>
      <table className="-ml-4">
        <tbody>
          {props.showNowPlaying && npData && npData.currently_playing && (
            <tr className="group hover:bg-[--color-bg-secondary]">
              <td className="w-[18px] pr-2 align-middle"></td>
              <td className="color-fg-tertiary pr-2 sm:pr-4 text-sm whitespace-nowrap w-0">
                Now Playing
              </td>
              <td className="text-ellipsis overflow-hidden max-w-[400px] sm:max-w-[600px]">
                {props.hideArtists ? null : (
                  <>
                    <ArtistLinks artists={npData.track.artists} /> –{" "}
                  </>
                )}
                <Link
                  className="hover:text-[--color-fg-secondary]"
                  to={`/track/${npData.track.id}`}
                >
                  {npData.track.title}
                </Link>
              </td>
            </tr>
          )}
          {listens.map((item) => (
            <tr
              key={`last_listen_${item.time}`}
              className="group hover:bg-[--color-bg-secondary]"
            >
              <td className="w-[18px] pr-2 align-middle">
                <button
                  onClick={() => handleDelete(item)}
                  className="opacity-0 group-hover:opacity-100 transition-opacity text-(--color-fg-tertiary) hover:text-(--color-error)"
                  aria-label="Delete"
                  hidden={user === null || user === undefined}
                >
                  ×
                </button>
              </td>
              <td
                className="color-fg-tertiary pr-2 sm:pr-4 text-sm whitespace-nowrap w-0"
                title={new Date(item.time).toString()}
              >
                {timeSince(new Date(item.time))}
              </td>
              <td className="text-ellipsis overflow-hidden max-w-[400px] sm:max-w-[600px]">
                {props.hideArtists ? null : (
                  <>
                    <ArtistLinks artists={item.track.artists} /> –{" "}
                  </>
                )}
                <Link
                  className="hover:text-[--color-fg-secondary]"
                  to={`/track/${item.track.id}`}
                >
                  {item.track.title}
                </Link>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
