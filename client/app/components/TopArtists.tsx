import { useQuery } from "@tanstack/react-query";
import {
  apiFetch,
  type PaginatedResponse,
  type Ranked,
  type Artist,
} from "api/api";
import { Link } from "react-router";
import TopItemList from "./TopItemList";

interface Props {
  limit: number;
  period: string;
  artistId?: Number;
  albumId?: Number;
}

const getTopArtists = (args: { limit: number; period: string; page: number }) =>
  apiFetch<PaginatedResponse<Ranked<Artist>>>("/apis/web/v1/top-artists", args);

export default function TopArtists(props: Props) {
  const args = { limit: props.limit, period: props.period, page: 0 };
  const { isPending, isError, data, error } = useQuery({
    queryKey: ["top-artists", args],
    queryFn: () => getTopArtists(args),
  });

  const header = "Top artists";

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

  return (
    <div>
      <h3 className="hover:underline">
        <Link to={`/chart/top-artists?period=${props.period}`}>{header}</Link>
      </h3>
      <div className="max-w-[300px]">
        <TopItemList type="artist" data={data} />
        {data.items.length < 1 ? "Nothing to show" : ""}
      </div>
    </div>
  );
}
