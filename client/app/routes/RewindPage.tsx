import Rewind from "~/components/rewind/Rewind";
import type { Route } from "./+types/Home";
import { getRewindStats, type RewindStats } from "api/api";
import { useEffect, useState } from "react";

export function meta({}: Route.MetaArgs) {
  return [{ title: "Koito" }, { name: "description", content: "Koito" }];
}

export default function RewindPage() {
  const [stats, setStats] = useState<RewindStats | undefined>(undefined);
  useEffect(() => {
    getRewindStats({ year: 2025 }).then((r) => setStats(r));
  }, []);
  return (
    <div className="flex flex-col items-start">
      <h2 className="mt-12">Your 2025 Rewind</h2>
      {stats !== undefined && <Rewind stats={stats} />}
    </div>
  );
}
