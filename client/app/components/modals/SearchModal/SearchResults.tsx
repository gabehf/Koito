import { type SearchResponse } from "api/api";
import { useState } from "react";
import {
  SearchResultItem,
  SearchResultArtistItem,
} from "../../SearchResultItem";
import SearchResultSelectorItem from "../../SearchResultSelectorItem";
import SubHeader from "../../primitives/SubHeader";
import CardHeader from "~/components/primitives/CardHeader";

interface Props {
  data?: SearchResponse;
  onSelect: Function;
  selectorMode?: boolean;
}

export default function SearchResults({ data, onSelect, selectorMode }: Props) {
  const [selected, setSelected] = useState(0);

  const selectItem = (title: string, id: number) => {
    if (selected === id) {
      setSelected(0);
      onSelect({ id: 0, title: "" });
    } else {
      setSelected(id);
      onSelect({ id: id, title: title });
    }
  };

  if (!data) return <></>;

  const itemClasses = "flex flex-col items-start bg rounded w-full";
  const cardClasses = `border bg-secondary rounded-(--border-radius) p-4 w-full`;
  const hClasses = "mt-4 mb-2";

  const renderSection = (label: string, items: React.ReactNode) => (
    <div className="mt-2 sm:mt-4">
      {selectorMode ? (
        <h3 className={hClasses}>{label}</h3>
      ) : (
        <CardHeader isOffset>{label}</CardHeader>
      )}
      <div className={selectorMode ? itemClasses : cardClasses}>{items}</div>
    </div>
  );

  return (
    <div className="w-full">
      {data.artists &&
        data.artists.length > 0 &&
        renderSection(
          "Artists",
          <div className="p-2 flex flex-wrap gap-2 sm:gap-4">
            {data.artists.map((artist) =>
              selectorMode ? (
                <SearchResultSelectorItem
                  key={artist.id}
                  id={artist.id}
                  onClick={() => selectItem(artist.name, artist.id)}
                  text={artist.name}
                  img={artist.image.small}
                  active={selected === artist.id}
                />
              ) : (
                <SearchResultArtistItem
                  key={artist.id}
                  to={`/artist/${artist.id}`}
                  onClick={() => onSelect(artist.id)}
                  text={artist.name}
                  img={artist.image.small}
                />
              )
            )}
          </div>
        )}
      {data.albums &&
        data.albums.length > 0 &&
        renderSection(
          "Albums",
          data.albums.map((album) =>
            selectorMode ? (
              <SearchResultSelectorItem
                key={album.id}
                id={album.id}
                onClick={() => selectItem(album.title, album.id)}
                text={album.title}
                subtext={
                  album.is_various_artists
                    ? "Various Artists"
                    : album.artists[0].name
                }
                img={album.image.small}
                active={selected === album.id}
              />
            ) : (
              <SearchResultItem
                key={album.id}
                to={`/album/${album.id}`}
                onClick={() => onSelect(album.id)}
                text={album.title}
                subtext={
                  album.is_various_artists
                    ? "Various Artists"
                    : album.artists[0].name
                }
                img={album.image.small}
              />
            )
          )
        )}
      {data.tracks &&
        data.tracks.length > 0 &&
        renderSection(
          "Tracks",
          data.tracks.map((track) =>
            selectorMode ? (
              <SearchResultSelectorItem
                key={track.id}
                id={track.id}
                onClick={() => selectItem(track.title, track.id)}
                text={track.title}
                subtext={track.artists.map((a) => a.name).join(", ")}
                img={track.image.small}
                active={selected === track.id}
              />
            ) : (
              <SearchResultItem
                key={track.id}
                to={`/track/${track.id}`}
                onClick={() => onSelect(track.id)}
                text={track.title}
                subtext={track.artists.map((a) => a.name).join(", ")}
                img={track.image.small}
                imgSize={75}
              />
            )
          )
        )}
    </div>
  );
}
