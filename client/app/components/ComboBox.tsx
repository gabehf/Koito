import { useDeferredValue, useEffect, useState } from "react";
import { search, type SearchResponse } from "api/api";
import { useCombobox } from "downshift";

interface Props {
  onSelection: (selection: any) => void;
  filterFunction: (r: SearchResponse) => SearchResponse;
}

export default function ComboBox({ onSelection, filterFunction }: Props) {
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const [data, setData] = useState<any[]>([]);

  useEffect(() => {
    if (deferredQuery) {
      search(deferredQuery).then((r) => {
        const filtered = filterFunction(r);
        setData([...filtered.artists, ...filtered.albums, ...filtered.tracks]);
      });
    }
  }, [deferredQuery]);

  const { isOpen, getMenuProps, getInputProps, getItemProps, selectedItem } =
    useCombobox({
      items: data,
      itemToString(item) {
        return item ? item.title || item.name : "";
      },
      onInputValueChange: ({ inputValue }) => {
        setQuery(inputValue);
      },
      onSelectedItemChange: ({ selectedItem: newSelectedItem }) => {
        if (newSelectedItem) {
          setQuery(newSelectedItem.name);
          onSelection(selectedItem);
        }
      },
    });

  return (
    <div className="flex-grow">
      <input
        {...getInputProps()}
        value={query}
        placeholder="Add an artist"
        className="mx-auto fg bg rounded-md p-3 w-full"
      />
      <ul
        {...getMenuProps()}
        className={`bg rounded-b-md p-3 absolute ${
          !(isOpen && data.length) && "hidden"
        }`}
      >
        {isOpen &&
          data &&
          data.map((item, index) => (
            <li
              className="fg py-2 px-3 rounded-md shadow-sm cursor-pointer hover:bg-(--color-bg-tertiary)"
              key={item.id}
              {...getItemProps({ item, index })}
            >
              {item.title || item.name}
            </li>
          ))}
      </ul>
    </div>
  );
}
