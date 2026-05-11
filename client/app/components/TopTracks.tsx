import { useQuery } from "@tanstack/react-query";
import {
  apiFetch,
  type PaginatedResponse,
  type Ranked,
  type Track,
} from "api/api";
import { Link } from "react-router";
import TopItemList, { TopItemListSkeleton } from "./TopItemList";
import CardHeader from "./primitives/CardHeader";

interface Props {
  limit: number;
  period: string;
  artistId?: Number;
  albumId?: Number;
}

const getTopTracks = (args: {
  limit: number;
  period: string;
  artist_id?: number;
  album_id?: number;
  page: number;
}) =>
  apiFetch<PaginatedResponse<Ranked<Track>>>("/apis/web/v1/top-tracks", args);

const TopTracks = (props: Props) => {
  const args = {
    limit: props.limit,
    period: props.period,
    artist_id: props.artistId as number | undefined,
    album_id: props.albumId as number | undefined,
    page: 0,
  };
  const { isPending, isError, data, error } = useQuery({
    queryKey: ["top-tracks", args],
    queryFn: () => getTopTracks(args),
  });

  const header = "Top tracks";

  if (isPending) {
    return (
      <div className="min-w-[350px] sm:min-w-[450px] md:w-full max-w-[725px] xl:max-w-[450px]">
        <CardHeader isOffset>{header}</CardHeader>
        <div className="mt-7">
          <TopItemListSkeleton count={props.limit} ranked type="track" />
        </div>
      </div>
    );
  } else if (isError) {
    return (
      <div className="min-w-[350px] sm:min-w-[450px] ">
        <h3>{header}</h3>
        <p className="error">Error: {error.message}</p>
      </div>
    );
  }
  if (!data.items) return;

  let params = "";
  params += props.artistId ? `&artist_id=${props.artistId}` : "";
  params += props.albumId ? `&album_id=${props.albumId}` : "";

  const slug = `/chart/top-tracks?period=${props.period}&limit=100${params}`;

  if (!data.items[0]) {
    return (
      <div className="min-w-[350px] sm:min-w-[450px] md:w-full max-w-[725px] xl:max-w-[450px]">
        <CardHeader
          isOffset
          to={`/chart/top-tracks?period=${props.period}${params}`}
        >
          {header}
        </CardHeader>
        <p className="ml-6 mt-6">Nothing to show</p>
      </div>
    );
  }

  return (
    <div className="min-w-[350px] sm:min-w-[450px] md:w-full max-w-[725px] xl:max-w-[450px]">
      <CardHeader
        isOffset
        to={`/chart/top-tracks?period=${props.period}${params}`}
      >
        {header}
      </CardHeader>
      <div className="mt-7">
        <TopItemList
          ranked
          type="track"
          data={data}
          separators
          slug={slug}
          showSeeMore
        />
      </div>
    </div>
  );
};

export default TopTracks;
