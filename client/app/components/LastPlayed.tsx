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
import CardHeader from "./primitives/CardHeader";

interface Props {
  limit: number;
  artistId?: Number;
  albumId?: Number;
  trackId?: number;
  hideArtists?: boolean;
  showNowPlaying?: boolean;
  showSeeMore?: boolean;
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
    <div className="w-[350px] md:w-full max-w-[725px] xl:max-w-[1100px]">
      <CardHeader to={`/listens?period=all_time${params}`}>{header}</CardHeader>
      {listens.length < 1 && "Nothing to show"}
      <table className="table-fixed border-collapse mt-6 w-[350px] sm:w-full">
        <tbody>
          {props.showNowPlaying && npData && npData.currently_playing && (
            <tr className="group border-b-1 border-(--color-bg-tertiary) relative last:border-b-0">
              <td className="py-3 pr-3 w-11">
                <Link to={`/track/${npData.track.id}`}>
                  <Image
                    src={imageUrl(npData.track.image, "small")}
                    size={32}
                  />
                </Link>
              </td>
              <td className="w-[150px] sm:w-full">
                {props.hideArtists ? null : (
                  <>
                    <ArtistLinks artists={npData.track.artists} />
                    {" — "}
                  </>
                )}
                <Link
                  className="hover:text-[--color-fg-secondary]"
                  to={`/track/${npData.track.id}`}
                >
                  {npData.track.title}
                </Link>
              </td>
              <td className="color-fg-tertiary pr-2 sm:pr-4 text-sm text-end whitespace-nowrap w-[100px]">
                <div className="sm:-mr-[18px] relative">
                  <div className="h-1.5 w-1.5 rounded-full bg-(--color-primary) absolute top-1.5 left-3"></div>{" "}
                  Now Playing
                </div>
              </td>
            </tr>
          )}
          {listens.map((item) => (
            <tr
              key={`last_listen_${item.time}`}
              className="group border-b-1 border-(--color-bg-tertiary) relative last:border-b-0"
            >
              <td className="py-3 pr-3 w-11">
                <Link to={`/track/${item.track.id}`}>
                  <Image src={imageUrl(item.track.image, "small")} size={32} />
                </Link>
              </td>
              <td className="w-[150px] sm:w-full">
                {props.hideArtists ? null : (
                  <>
                    <ArtistLinks artists={item.track.artists} />
                    {" — "}
                  </>
                )}
                <Link
                  className="hover:text-(--color-fg-secondary)"
                  to={`/track/${item.track.id}`}
                >
                  {item.track.title}
                </Link>
              </td>
              <td
                className="color-fg-tertiary pr-2 sm:pr-4 text-sm text-end whitespace-nowrap w-[100px]"
                title={new Date(item.time).toString()}
              >
                <p className="sm:-mr-[18px]">
                  {timeSince(new Date(item.time))}
                </p>
              </td>
              <td className="pr-2 align-middle hidden sm:table">
                <button
                  onClick={() => handleDelete(item)}
                  className="absolute top-1/2 -translate-y-1/2 -right-5 opacity-0 group-hover:opacity-100 transition-opacity text-(--color-fg-tertiary) hover:text-(--color-error)"
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
      {props.showSeeMore && (
        <div className="flex items-center w-[350px] sm:w-full">
          <Link
            to={`/listens?period=all_time${params}`}
            className="inline-block w-fit mx-auto text-(--color-fg-secondary) hover:text-(--color-fg) hover:cursor-pointer"
          >
            SEE MORE →
          </Link>
        </div>
      )}
    </div>
  );
}
