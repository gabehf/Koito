import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { timeSince } from "~/utils/utils";
import ArtistLinks from "./ArtistLinks";
import Image from "./primitives/Image";
import {
  apiFetch,
  deleteListen,
  type Listen,
  type NowPlaying,
  type PaginatedResponse,
} from "api/api";
import { Link } from "react-router";
import { useAppContext } from "~/providers/AppProvider";
import CardHeader from "./primitives/CardHeader";
import ListensTable from "./ListensTable";

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

  const showNP = (): boolean => {
    if (!props.showNowPlaying || !npData?.currently_playing) return false;

    const { albumId, artistId, trackId } = props;
    const track = npData.track;

    if (!albumId && !artistId && !trackId) return true;

    if (albumId) return track.album_id === albumId;
    if (artistId) return track.artists.some((a) => a.id === artistId);
    if (trackId) return track.id === trackId;

    return false;
  };

  const listens = items ?? data.items;

  let params = "";
  params += props.artistId ? `&artist_id=${props.artistId}` : "";
  params += props.albumId ? `&album_id=${props.albumId}` : "";
  params += props.trackId ? `&track_id=${props.trackId}` : "";

  return (
    <div className="text-[13px] sm:text-[15px] w-[350px] md:w-full max-w-[725px] xl:max-w-[1100px]">
      <CardHeader to={`/listens?period=all_time${params}`}>{header}</CardHeader>
      {listens.length < 1 && "Nothing to show"}
      {listens.length < 1 ? (
        "Nothing to show"
      ) : (
        <ListensTable
          listens={listens}
          npData={npData}
          showNP={showNP()}
          hideArtists={props.hideArtists}
          onDelete={handleDelete}
        />
      )}
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
