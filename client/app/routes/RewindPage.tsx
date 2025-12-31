import Rewind from "~/components/rewind/Rewind";
import type { Route } from "./+types/Home";
import { type RewindStats } from "api/api";
import { useState } from "react";
import type { LoaderFunctionArgs } from "react-router";
import { useLoaderData } from "react-router";
import { getRewindYear } from "~/utils/utils";

export async function clientLoader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url);
  const year = url.searchParams.get("year") || getRewindYear();

  const res = await fetch(`/apis/web/v1/summary?year=${year}`);
  if (!res.ok) {
    throw new Response("Failed to load summary", { status: 500 });
  }

  const stats: RewindStats = await res.json();
  stats.title = `Your ${year} Rewind`;
  return { stats };
}

export function meta({}: Route.MetaArgs) {
  return [
    { title: `Rewind - Koito` },
    { name: "description", content: "Rewind - Koito" },
  ];
}

export default function RewindPage() {
  const [showTime, setShowTime] = useState(false);
  const { stats: stats } = useLoaderData<{ stats: RewindStats }>();
  return (
    <main className="w-18/20">
      <title>{stats.title} - Koito</title>
      <meta property="og:title" content={`${stats.title} - Koito`} />
      <meta name="description" content={`${stats.title} - Koito`} />
      <div className="flex flex-col items-start mt-20 gap-10">
        <div className="flex items-center gap-3">
          <label htmlFor="show-time-checkbox">Show time listened?</label>
          <input
            type="checkbox"
            name="show-time-checkbox"
            checked={showTime}
            onChange={(e) => setShowTime(!showTime)}
          ></input>
        </div>
        {stats !== undefined && <Rewind stats={stats} includeTime={showTime} />}
      </div>
    </main>
  );
}
