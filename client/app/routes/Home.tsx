import type { Route } from "./+types/Home";
import TopTracks from "~/components/TopTracks";
import LastPlayed from "~/components/LastPlayed";
import ActivityGrid from "~/components/ActivityGrid";
import TopAlbums from "~/components/TopAlbums";
import TopArtistsCard from "~/components/TopArtistsCard";
import AllTimeStats from "~/components/AllTimeStats";
import { useState } from "react";
import PeriodSelector from "~/components/PeriodSelector";
import { useAppContext } from "~/providers/AppProvider";
import TopAlbumsCard from "~/components/TopAlbumsCard";

export function meta({}: Route.MetaArgs) {
  return [{ title: "Koito" }, { name: "description", content: "Koito" }];
}

export default function Home() {
  const [period, setPeriod] = useState("week");

  const { homeItems } = useAppContext();

  return (
    <main className="flex flex-grow justify-center pb-4 w-full bg-linear-to-b to-(--color-bg) from-(--color-bg-secondary) to-60%">
      <div className="flex-1 flex flex-col items-center gap-16 min-h-0 sm:mt-20 mt-10">
        <div className="flex flex-col md:flex-row gap-10 md:gap-20">
          <AllTimeStats />
          <ActivityGrid configurable />
        </div>
        <PeriodSelector setter={setPeriod} current={period} />
        <div className="flex flex-wrap gap-10 2xl:gap-20 xl:gap-10 justify-between mx-5 md:gap-5">
          <TopArtistsCard period={period} />
          <TopAlbumsCard period={period} />
          <TopTracks period={period} limit={homeItems} />
          <LastPlayed showNowPlaying={true} limit={14} />
        </div>
      </div>
    </main>
  );
}
