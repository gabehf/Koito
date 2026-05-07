import type { Route } from "./+types/Home";
import TopTracks from "~/components/TopTracks";
import LastPlayed from "~/components/LastPlayed";
import ActivityGrid from "~/components/ActivityGrid";
import TopAlbums from "~/components/TopAlbums";
import TopArtists from "~/components/TopArtists";
import TopArtistsCard from "~/components/TopArtistsCard";
import AllTimeStats from "~/components/AllTimeStats";
import { useState } from "react";
import PeriodSelector from "~/components/PeriodSelector";
import { useAppContext } from "~/providers/AppProvider";
import TopAlbumsCard from "~/components/TopAlbumsCard";
import PinnedItemGrid from "~/components/PinnedItemGrid";

export function meta({}: Route.MetaArgs) {
  return [{ title: "Koito" }, { name: "description", content: "Koito" }];
}

export default function Home() {
  const [period, setPeriod] = useState("week");

  const { homeItems } = useAppContext();

  const gradientClasses =
    "bg-linear-to-b to-(--color-bg) from-(--color-bg-secondary) to-60%";

  return (
    <main className="flex flex-grow justify-center pb-4 w-full">
      <div className="flex-1 flex flex-col items-center gap-16 min-h-0 sm:mt-20 mt-10 mx-10">
        <div className="flex flex-col lg:flex-row gap-10 md:gap-20">
          <AllTimeStats />
          <ActivityGrid configurable />
        </div>
        <PeriodSelector setter={setPeriod} current={period} />
        <div className="container justify-center flex flex-wrap gap-10">
          {/*<PinnedItemGrid />*/}
          <TopArtistsCard period={period} />
          <TopAlbumsCard period={period} />
          <TopTracks period={period} limit={10} />
          <LastPlayed showNowPlaying={true} limit={28} />
          {/*<TopArtists period={period} limit={10} />
          <TopAlbums period={period} limit={10} />*/}
        </div>
      </div>
    </main>
  );
}
