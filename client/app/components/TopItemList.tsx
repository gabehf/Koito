import { Link, useNavigate } from "react-router";
import ArtistLinks from "./ArtistLinks";
import {
  imageUrl,
  type Album,
  type Artist,
  type Track,
  type PaginatedResponse,
} from "api/api";

type Item = Album | Track | Artist;

interface Props<T extends Item> {
  data: PaginatedResponse<T>;
  separators?: ConstrainBoolean;
  ranked?: boolean;
  type: "album" | "track" | "artist";
  className?: string;
}

export default function TopItemList<T extends Item>({
  data,
  separators,
  type,
  className,
  ranked,
}: Props<T>) {
  const currentParams = new URLSearchParams(location.search);
  const page = Math.max(parseInt(currentParams.get("page") || "1"), 1);

  let lastRank = 0;

  const calculateRank = (data: Item[], page: number, index: number): number => {
    if (
      index === 0 ||
      data[index] == undefined ||
      !(data[index].listen_count === data[index - 1].listen_count)
    ) {
      lastRank = index + 1 + (page - 1) * 100;
    }
    return lastRank;
  };

  return (
    <div className={`flex flex-col gap-1 ${className} min-w-[200px]`}>
      {data.items.map((item, index) => {
        const key = `${type}-${item.id}`;
        return (
          <div
            key={key}
            style={{ fontSize: 12 }}
            className={`${
              separators && index !== data.items.length - 1
                ? "border-b border-(--color-fg-tertiary) mb-1 pb-2"
                : ""
            }`}
          >
            <ItemCard
              ranked={ranked}
              rank={calculateRank(data.items, page, index)}
              item={item}
              type={type}
              key={type + item.id}
            />
          </div>
        );
      })}
    </div>
  );
}

function ItemCard({
  item,
  type,
  rank,
  ranked,
}: {
  item: Item;
  type: "album" | "track" | "artist";
  rank: number;
  ranked?: boolean;
}) {
  const itemClasses = `flex items-center gap-2`;

  switch (type) {
    case "album": {
      const album = item as Album;

      return (
        <div style={{ fontSize: 12 }} className={itemClasses}>
          {ranked && <div className="w-7 text-end">{rank}</div>}
          <Link to={`/album/${album.id}`}>
            <img
              loading="lazy"
              src={imageUrl(album.image, "small")}
              alt={album.title}
              className="min-w-[48px]"
            />
          </Link>
          <div>
            <Link
              to={`/album/${album.id}`}
              className="hover:text-(--color-fg-secondary)"
            >
              <span style={{ fontSize: 14 }}>{album.title}</span>
            </Link>
            <br />
            {album.is_various_artists ? (
              <span className="color-fg-secondary">Various Artists</span>
            ) : (
              <div>
                <ArtistLinks
                  artists={
                    album.artists
                      ? [album.artists[0]]
                      : [{ id: 0, name: "Unknown Artist" }]
                  }
                />
              </div>
            )}
            <div className="color-fg-secondary">{album.listen_count} plays</div>
          </div>
        </div>
      );
    }
    case "track": {
      const track = item as Track;

      return (
        <div style={{ fontSize: 12 }} className={itemClasses}>
          {ranked && <div className="w-7 text-end">{rank}</div>}
          <Link to={`/track/${track.id}`}>
            <img
              loading="lazy"
              src={imageUrl(track.image, "small")}
              alt={track.title}
              className="min-w-[48px]"
            />
          </Link>
          <div>
            <Link
              to={`/track/${track.id}`}
              className="hover:text-(--color-fg-secondary)"
            >
              <span style={{ fontSize: 14 }}>{track.title}</span>
            </Link>
            <br />
            <div>
              <ArtistLinks
                artists={track.artists || [{ id: 0, Name: "Unknown Artist" }]}
              />
            </div>
            <div className="color-fg-secondary">{track.listen_count} plays</div>
          </div>
        </div>
      );
    }
    case "artist": {
      const artist = item as Artist;
      return (
        <div style={{ fontSize: 12 }} className={itemClasses}>
          {ranked && <div className="w-7 text-end">{rank}</div>}
          <Link
            className={
              itemClasses + " mt-1 mb-[6px] hover:text-(--color-fg-secondary)"
            }
            to={`/artist/${artist.id}`}
          >
            <img
              loading="lazy"
              src={imageUrl(artist.image, "small")}
              alt={artist.name}
              className="min-w-[48px]"
            />
            <div>
              <span style={{ fontSize: 14 }}>{artist.name}</span>
              <div className="color-fg-secondary">
                {artist.listen_count} plays
              </div>
            </div>
          </Link>
        </div>
      );
    }
  }
}
