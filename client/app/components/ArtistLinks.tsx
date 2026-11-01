import React from "react";
import { ProfileLink as Link } from "~/components/ProfileLink";

type Artist = {
  id: number;
  name: string;
};

type ArtistLinksProps = {
  artists: Artist[];
};

const ArtistLinks: React.FC<ArtistLinksProps> = ({ artists }) => {
  return (
    <>
      {artists.map((artist, index) => (
        <span key={artist.id} className="color-fg-secondary">
          <Link
            className="hover:text-(--color-fg-tertiary)"
            to={`/artist/${artist.id}`}
          >
            {artist.name}
          </Link>
          {index < artists.length - 1 ? ", " : ""}
        </span>
      ))}
    </>
  );
};

export default ArtistLinks;
