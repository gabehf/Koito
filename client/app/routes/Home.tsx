import type { Route } from "./+types/Home";
import TopTracks from "~/components/TopTracks";
import LastPlayed from "~/components/LastPlayed";
import ActivityGrid from "~/components/ActivityGrid";
import TopArtistsCard from "~/components/TopArtistsCard";
import AllTimeStats from "~/components/AllTimeStats";
import { useSearchParams } from "react-router";
import PeriodSelector from "~/components/PeriodSelector";
import { useAppContext } from "~/providers/AppProvider";
import TopAlbumsCard from "~/components/TopAlbumsCard";
import PinnedItemGrid from "~/components/PinnedItemGrid";

export function meta({}: Route.MetaArgs) {
  return [{ title: "Koito" }, { name: "description", content: "Koito" }];
}

export default function Home() {
  const [searchParams, setSearchParams] = useSearchParams();
  const selectedPeriod = searchParams.get("period");
  const period = ["day", "week", "month", "year", "all_time"].includes(
    selectedPeriod ?? "",
  )
    ? selectedPeriod!
    : "week";
  const setPeriod = (value: string) =>
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      next.set("period", value);
      return next;
    });

  const { homeItems } = useAppContext();

  const gradientClasses =
    "bg-linear-to-b to-(--color-bg) from-(--color-bg-secondary) to-60%";

  return (
    <main className="flex grow justify-center pb-4 w-full">
      <div className="flex-1 flex flex-col items-center gap-10 md:gap-12 min-h-0 mt-8 sm:mt-10 mx-4 sm:mx-10">
        <div className="flex flex-col lg:flex-row gap-10 md:gap-20">
          <AllTimeStats />
          <ActivityGrid configurable />
        </div>
        <PeriodSelector
          setter={setPeriod}
          current={period}
          disableCache={searchParams.has("period")}
        />
        <div className="container justify-center flex flex-wrap gap-10">
          {/*<PinnedItemGrid />*/}
          <TopArtistsCard period={period} />
          <TopAlbumsCard period={period} />
          <TopTracks
            period={period}
            limit={10}
            className="min-w-[350px] w-full max-w-[750px] 2xl:max-w-[450px]"
          />
          <LastPlayed showNowPlaying={true} limit={28} showSeeMore />
        </div>
      </div>
    </main>
  );
}
