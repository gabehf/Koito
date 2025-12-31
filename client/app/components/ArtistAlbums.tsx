import { useQuery } from "@tanstack/react-query";
import { getTopAlbums, imageUrl, type getItemsArgs } from "api/api";
import { Link } from "react-router";

interface Props {
  artistId: number;
  name: string;
  period: string;
}

export default function ArtistAlbums({ artistId, name, period }: Props) {
  const { isPending, isError, data, error } = useQuery({
    queryKey: [
      "top-albums",
      { limit: 99, period: "all_time", artist_id: artistId, page: 0 },
    ],
    queryFn: ({ queryKey }) => getTopAlbums(queryKey[1] as getItemsArgs),
  });

  if (isPending) {
    return (
      <div>
        <h3>Albums From This Artist</h3>
        <p>Loading...</p>
      </div>
    );
  }
  if (isError) {
    return (
      <div>
        <h3>Albums From This Artist</h3>
        <p className="error">Error:{error.message}</p>
      </div>
    );
  }

  return (
    <div>
      <h3>Albums featuring {name}</h3>
      <div className="flex flex-wrap gap-8">
        {data.items.map((item) => (
          <Link to={`/album/${item.id}`} className="flex gap-2 items-start">
            <img
              src={imageUrl(item.image, "medium")}
              alt={item.title}
              style={{ width: 130 }}
            />
            <div className="w-[180px] flex flex-col items-start gap-1">
              <p>{item.title}</p>
              <p className="text-sm color-fg-secondary">
                {item.listen_count} play{item.listen_count > 1 ? "s" : ""}
              </p>
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}
