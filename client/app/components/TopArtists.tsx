import { useQuery } from "@tanstack/react-query";
import ArtistLinks from "./ArtistLinks";
import { getTopArtists, imageUrl, type getItemsArgs } from "api/api";
import { Link } from "react-router";
import TopListSkeleton from "./skeletons/TopListSkeleton";
import TopItemList from "./TopItemList";

interface Props {
  limit: number;
  period: string;
  artistId?: Number;
  albumId?: Number;
}

export default function TopArtists(props: Props) {
  const { isPending, isError, data, error } = useQuery({
    queryKey: [
      "top-artists",
      { limit: props.limit, period: props.period, page: 0 },
    ],
    queryFn: ({ queryKey }) => getTopArtists(queryKey[1] as getItemsArgs),
  });

  if (isPending) {
    return (
      <div className="w-[300px]">
        <h2>Top Artists</h2>
        <p>Loading...</p>
      </div>
    );
  } else if (isError) {
    return (
      <div className="w-[300px]">
        <h2>Top Artists</h2>
        <p className="error">Error: {error.message}</p>
      </div>
    );
  }

  return (
    <div>
      <h2 className="hover:underline">
        <Link to={`/chart/top-artists?period=${props.period}`}>
          Top Artists
        </Link>
      </h2>
      <div className="max-w-[300px]">
        <TopItemList type="artist" data={data} />
        {data.items.length < 1 ? "Nothing to show" : ""}
      </div>
    </div>
  );
}
