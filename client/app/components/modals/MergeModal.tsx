import { useDeferredValue, useEffect, useState } from "react";
import { Modal } from "./Modal";
import { search, type SearchResponse } from "api/api";
import SearchResults from "../SearchResults";
import type {
  MergeFunc,
  MergeSearchCleanerFunc,
} from "~/routes/MediaItems/MediaLayout";
import { useNavigate } from "react-router";

interface Props {
  open: boolean;
  setOpen: Function;
  type: string;
  currentId: number;
  currentTitle: string;
  mergeFunc: MergeFunc;
  mergeCleanerFunc: MergeSearchCleanerFunc;
}

export default function MergeModal(props: Props) {
  const [query, setQuery] = useState(props.currentTitle);
  const [data, setData] = useState<SearchResponse>();
  const deferredQuery = useDeferredValue(query);
  const [mergeTarget, setMergeTarget] = useState<{ title: string; id: number }>(
    { title: "", id: 0 }
  );
  const [mergeOrderReversed, setMergeOrderReversed] = useState(false);
  const [replaceImage, setReplaceImage] = useState(false);
  const navigate = useNavigate();

  const closeMergeModal = () => {
    props.setOpen(false);
    setQuery("");
    setData(undefined);
    setMergeOrderReversed(false);
    setMergeTarget({ title: "", id: 0 });
  };

  const toggleSelect = ({ title, id }: { title: string; id: number }) => {
    setMergeTarget({ title: title, id: id });
  };

  useEffect(() => {
    console.log("mergeTarget", mergeTarget);
  }, [mergeTarget]);

  const doMerge = () => {
    let from, to;
    if (!mergeOrderReversed) {
      from = mergeTarget;
      to = { id: props.currentId, title: props.currentTitle };
    } else {
      from = { id: props.currentId, title: props.currentTitle };
      to = mergeTarget;
    }
    props
      .mergeFunc(from.id, to.id, replaceImage)
      .then((r) => {
        if (r.ok) {
          if (mergeOrderReversed) {
            navigate(`/${props.type.toLowerCase()}/${mergeTarget.id}`);
            closeMergeModal();
          } else {
            window.location.reload();
          }
        } else {
          // TODO: handle error
          console.log(r);
        }
      })
      .catch((err) => console.log(err));
  };

  useEffect(() => {
    if (deferredQuery) {
      search(deferredQuery).then((r) => {
        r = props.mergeCleanerFunc(r, props.currentId);
        setData(r);
      });
    }
  }, [deferredQuery]);

  return (
    <Modal isOpen={props.open} onClose={closeMergeModal}>
      <h3>Merge {props.type}s</h3>
      <div className="flex flex-col items-center">
        <input
          type="text"
          autoFocus
          defaultValue={props.currentTitle}
          // i find my stupid a(n) logic to be a little silly so im leaving it in even if its not optimal
          placeholder={`Search for a${props.type.toLowerCase()[0] === "a" ? "n" : ""
            } ${props.type.toLowerCase()} to be merged into the current ${props.type.toLowerCase()}`}
          className="w-full mx-auto fg bg rounded p-2"
          onFocus={(e) => { setQuery(e.target.value); e.target.select()}}
          onChange={(e) => setQuery(e.target.value)}
        />
        <SearchResults selectorMode data={data} onSelect={toggleSelect} />
        {mergeTarget.id !== 0 ? (
          <>
            {mergeOrderReversed ? (
              <p className="mt-5">
                <strong>{props.currentTitle}</strong> will be merged into{" "}
                <strong>{mergeTarget.title}</strong>
              </p>
            ) : (
              <p className="mt-5">
                <strong>{mergeTarget.title}</strong> will be merged into{" "}
                <strong>{props.currentTitle}</strong>
              </p>
            )}
            <button
              className="hover:cursor-pointer px-5 py-2 rounded-md mt-5 bg-(--color-bg) hover:bg-(--color-bg-tertiary)"
              onClick={doMerge}
            >
              Merge Items
            </button>
            <div className="flex items-center gap-2 mt-3">
              <input
                type="checkbox"
                name="reverse-merge-order"
                checked={mergeOrderReversed}
                onChange={() => setMergeOrderReversed(!mergeOrderReversed)}
              />
              <label htmlFor="reverse-merge-order">Reverse merge order</label>
            </div>
            {(props.type.toLowerCase() === "album" ||
              props.type.toLowerCase() === "artist") && (
              <div className="flex items-center gap-2 mt-3">
                <input
                  type="checkbox"
                  name="replace-image"
                  checked={replaceImage}
                  onChange={() => setReplaceImage(!replaceImage)}
                />
                <label htmlFor="replace-image">Replace image</label>
              </div>
            )}
          </>
        ) : (
          ""
        )}
      </div>
    </Modal>
  );
}
