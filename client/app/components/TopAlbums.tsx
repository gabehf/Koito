import { useQuery } from "@tanstack/react-query";
import ArtistLinks from "./ArtistLinks";
import {
  getTopAlbums,
  getTopTracks,
  imageUrl,
  type getItemsArgs,
} from "api/api";
import { Link } from "react-router";
import TopListSkeleton from "./skeletons/TopListSkeleton";
import TopItemList from "./TopItemList";

interface Props {
  limit: number;
  period: string;
  artistId?: Number;
}

export default function TopAlbums(props: Props) {
  const { isPending, isError, data, error } = useQuery({
    queryKey: [
      "top-albums",
      {
        limit: props.limit,
        period: props.period,
        artistId: props.artistId,
        page: 0,
      },
    ],
    queryFn: ({ queryKey }) => getTopAlbums(queryKey[1] as getItemsArgs),
  });

  if (isPending) {
    return (
      <div className="w-[300px]">
        <h2>Top Albums</h2>
        <p>Loading...</p>
      </div>
    );
  } else if (isError) {
    return (
      <div className="w-[300px]">
        <h2>Top Albums</h2>
        <p className="error">Error: {error.message}</p>
      </div>
    );
  }

  return (
    <div>
      <h2 className="hover:underline">
        <Link
          to={`/chart/top-albums?period=${props.period}${
            props.artistId ? `&artist_id=${props.artistId}` : ""
          }`}
        >
          Top Albums
        </Link>
      </h2>
      <div className="max-w-[300px]">
        <TopItemList type="album" data={data} />
        {data.items.length < 1 ? "Nothing to show" : ""}
      </div>
    </div>
  );
}
