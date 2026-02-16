import { useDeferredValue, useEffect, useState } from "react";
import { Modal } from "./Modal";
import { search, type SearchResponse } from "api/api";
import SearchResults from "../SearchResults";

interface Props {
  open: boolean;
  setOpen: Function;
}

export default function SearchModal({ open, setOpen }: Props) {
  const [query, setQuery] = useState("");
  const [data, setData] = useState<SearchResponse>();
  const deferredQuery = useDeferredValue(query);

  const closeSearchModal = () => {
    setOpen(false);
    setQuery("");
    setData(undefined);
  };

  useEffect(() => {
    if (deferredQuery) {
      search(deferredQuery).then((r) => {
        setData(r);
      });
    }
  }, [deferredQuery]);

  return (
    <Modal isOpen={open} onClose={closeSearchModal}>
      <h3>Search</h3>
      <div className="flex flex-col items-center">
        <input
          type="text"
          autoFocus
          placeholder="Search for an artist, album, or track"
          className="w-full mx-auto fg bg rounded p-2"
          onChange={(e) => setQuery(e.target.value)}
        />
        <div className="h-3/4 w-full">
          <SearchResults data={data} onSelect={closeSearchModal} />
        </div>
      </div>
    </Modal>
  );
}
