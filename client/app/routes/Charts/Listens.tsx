import { createProfileApi } from "api/api";
import { useProfile } from "~/providers/ProfileProvider";
import ChartLayout from "./ChartLayout";
import { useLoaderData, type LoaderFunctionArgs } from "react-router";
import { ProfileLink as Link } from "~/components/ProfileLink";
import { type Listen, type PaginatedResponse } from "api/api";
import { timeSince } from "~/utils/utils";
import ArtistLinks from "~/components/ArtistLinks";
import { useState } from "react";
import { useAppContext } from "~/providers/AppProvider";
import ListensTable from "~/components/ListensTable";

export async function clientLoader({ request, params }: LoaderFunctionArgs) {
  const api = createProfileApi(params.slug!, () => request.signal);
  const url = new URL(request.url);
  const page = url.searchParams.get("page") || "0";
  url.searchParams.set("page", page);

  const res = await api.fetch(
    `/apis/web/v1/listens?${url.searchParams.toString()}`,
  );
  if (!res.ok) {
    throw new Response("Failed to load top tracks", { status: res.status });
  }

  const listens: PaginatedResponse<Listen> = await res.json();
  return { listens };
}

export default function Listens() {
  const { api } = useProfile();

  const { listens: initialData } = useLoaderData<{
    listens: PaginatedResponse<Listen>;
  }>();

  const [items, setItems] = useState<Listen[] | null>(null);
  const { user } = useAppContext();

  const handleDelete = async (listen: Listen) => {
    if (!initialData) return;
    try {
      const res = await api.deleteListen(listen);
      if (res.ok || (res.status >= 200 && res.status < 300)) {
        setItems((prev) =>
          (prev ?? initialData.items).filter((i) => i.time !== listen.time),
        );
      } else {
        console.error("Failed to delete listen:", res.status);
      }
    } catch (err) {
      console.error("Error deleting listen:", err);
    }
  };

  const listens = items ?? initialData.items;

  return (
    <ChartLayout
      title="Last Played"
      initialData={initialData}
      endpoint="listens"
      render={({ data, page, onNext, onPrev }) => (
        <div className="flex flex-col gap-5 text-sm md:text-[16px] w-11/12 max-w-[1000px]">
          <div className="flex gap-15 mx-auto">
            <button className="default" onClick={onPrev} disabled={page <= 1}>
              Prev
            </button>
            <button
              className="default"
              onClick={onNext}
              disabled={!data.has_next_page}
            >
              Next
            </button>
          </div>
          <div>
            <ListensTable listens={data.items} onDelete={handleDelete} />
          </div>
          <div className="flex gap-15 mx-auto">
            <button className="default" onClick={onPrev} disabled={page === 0}>
              Prev
            </button>
            <button
              className="default"
              onClick={onNext}
              disabled={!data.has_next_page}
            >
              Next
            </button>
          </div>
        </div>
      )}
    />
  );
}
