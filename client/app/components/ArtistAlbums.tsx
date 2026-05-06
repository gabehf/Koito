import { useQuery } from "@tanstack/react-query";
import {
  apiFetch,
  imageUrl,
  type PaginatedResponse,
  type Ranked,
  type Album,
} from "api/api";
import MediaItem from "./primitives/MediaItem";

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
          <div className="w-[330px]">
            <MediaItem
              image={imageUrl(item.item.image, "medium")}
              imageSize={125}
              link={`/album/${item.item.id}`}
              alignTop
              title={item.item.title}
              meta={`${item.item.listen_count} plays`}
            />
          </div>
        ))}
      </div>
    </div>
  );
}
