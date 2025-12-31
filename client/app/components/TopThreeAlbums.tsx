import { useQuery } from "@tanstack/react-query";
import { getTopAlbums, type getItemsArgs } from "api/api";
import AlbumDisplay from "./AlbumDisplay";

interface Props {
  period: string;
  artistId?: Number;
  vert?: boolean;
  hideTitle?: boolean;
}

export default function TopThreeAlbums(props: Props) {
  const { isPending, isError, data, error } = useQuery({
    queryKey: [
      "top-albums",
      { limit: 3, period: props.period, artist_id: props.artistId, page: 0 },
    ],
    queryFn: ({ queryKey }) => getTopAlbums(queryKey[1] as getItemsArgs),
  });

  if (isPending) {
    return <p>Loading...</p>;
  }
  if (isError) {
    return <p className="error">Error:{error.message}</p>;
  }

  console.log(data);

  return (
    <div>
      {!props.hideTitle && <h3>Top Three Albums</h3>}
      <div
        className={`flex ${props.vert ? "flex-col" : ""}`}
        style={{ gap: 15 }}
      >
        {data.items.map((item, index) => (
          <AlbumDisplay album={item} size={index === 0 ? 190 : 130} />
        ))}
      </div>
    </div>
  );
}
