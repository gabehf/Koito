import type { Route } from "./+types/Home";
import TopTracks from "~/components/TopTracks";
import LastPlays from "~/components/LastPlays";
import ActivityGrid from "~/components/ActivityGrid";
import TopAlbums from "~/components/TopAlbums";
import TopArtists from "~/components/TopArtists";
import AllTimeStats from "~/components/AllTimeStats";
import { useState } from "react";
import PeriodSelector from "~/components/PeriodSelector";
import { useAppContext } from "~/providers/AppProvider";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Koito" },
    { name: "description", content: "Koito" },
  ];
}

export default function Home() {
  const [period, setPeriod] = useState('week')

  const { homeItems } = useAppContext();

  return (
    <main className="flex flex-grow justify-center pb-4">
      <div className="flex-1 flex flex-col items-center gap-16 min-h-0 mt-20">
        <div className="flex flex-col md:flex-row gap-10 md:gap-20">
          <AllTimeStats />
          <ActivityGrid configurable />
        </div>
        <PeriodSelector setter={setPeriod} current={period} />
        <div className="flex flex-wrap gap-10 2xl:gap-20 xl:gap-10 justify-between mx-5 md:gap-5">
          <TopArtists period={period} limit={homeItems} />
          <TopAlbums period={period} limit={homeItems} />
          <TopTracks period={period} limit={homeItems} />
          <LastPlays showNowPlaying={true} limit={Math.floor(homeItems * 2.7)} />
        </div>
      </div>
    </main>
  );
}
