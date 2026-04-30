import { useQuery } from "@tanstack/react-query";
import {
  apiFetch,
  type PaginatedResponse,
  type Ranked,
  type Track,
} from "api/api";
import { Link } from "react-router";
import TopItemList from "./TopItemList";

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
      <div className="w-[300px]">
        <h3>{header}</h3>
        <p>Loading...</p>
      </div>
    );
  } else if (isError) {
    return (
      <div className="w-[300px]">
        <h3>{header}</h3>
        <p className="error">Error: {error.message}</p>
      </div>
    );
  }
  if (!data.items) return;

  let params = "";
  params += props.artistId ? `&artist_id=${props.artistId}` : "";
  params += props.albumId ? `&album_id=${props.albumId}` : "";

  return (
    <div>
      <h3 className="hover:underline">
        <Link to={`/chart/top-tracks?period=${props.period}${params}`}>
          {header}
        </Link>
      </h3>
      <div className="max-w-[300px]">
        <TopItemList type="track" data={data} />
        {data.items.length < 1 ? "Nothing to show" : ""}
      </div>
    </div>
  );
};

export default TopTracks;
