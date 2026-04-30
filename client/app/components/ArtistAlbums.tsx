import { useQuery } from "@tanstack/react-query";
import {
  apiFetch,
  imageUrl,
  type PaginatedResponse,
  type Ranked,
  type Album,
} from "api/api";
import { Link } from "react-router";

const getArtistAlbums = (artistId: number) =>
  apiFetch<PaginatedResponse<Ranked<Album>>>("/apis/web/v1/top-albums", {
    period: "all_time",
    limit: 99,
    artist_id: artistId,
  });

interface Props {
  artistId: number;
  name: string;
  period: string;
}

export default function ArtistAlbums({ artistId, name }: Props) {
  const { isPending, isError, data, error } = useQuery({
    queryKey: ["artist-albums", artistId],
    queryFn: () => getArtistAlbums(artistId),
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
        {data.items.length < 1 && "Nothing to show"}
        {data.items.map((item) => (
          <Link
            to={`/album/${item.item.id}`}
            className="flex gap-2 items-start"
          >
            <img
              src={imageUrl(item.item.image, "medium")}
              alt={item.item.title}
              style={{ width: 130 }}
            />
            <div className="w-[180px] flex flex-col items-start gap-1">
              <p>{item.item.title}</p>
              <p className="text-sm color-fg-secondary">
                {item.item.listen_count} play
                {item.item.listen_count > 1 ? "s" : ""}
              </p>
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}
