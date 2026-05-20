import { useQuery } from "@tanstack/react-query";
import {
  apiFetch,
  type PaginatedResponse,
  type Ranked,
  type Album,
} from "api/api";
import MediaItem from "./primitives/MediaItem";
import CardHeader from "./primitives/CardHeader";

const getArtistAlbums = (artistId: number) =>
  apiFetch<PaginatedResponse<Ranked<Album>>>("/apis/web/v1/top/albums", {
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
        <CardHeader>Albums From This Artist</CardHeader>
        <p>Loading...</p>
      </div>
    );
  }
  if (isError) {
    return (
      <div>
        <CardHeader>Albums From This Artist</CardHeader>
        <p className="error">Error:{error.message}</p>
      </div>
    );
  }

  return (
    <div>
      <CardHeader>Albums featuring {name}</CardHeader>
      <div className="flex flex-wrap gap-8 mt-8">
        {data.items.length < 1 && "Nothing to show"}
        {data.items.map((item) => (
          <div className="w-[330px]" key={item.item.id}>
            <MediaItem
              image={item.item.image}
              size="lg"
              link={`/album/${item.item.id}`}
              alt={item.item.title}
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
