import { Check } from "lucide-react";

interface Props {
  id: number;
  onClick: React.MouseEventHandler<HTMLButtonElement>;
  img: string;
  text: string;
  subtext?: string;
  active: boolean;
}

export default function SearchResultSelectorItem(props: Props) {
  return (
    <button
      className="px-3 py-2 flex gap-3 items-center hover:text-(--color-fg-secondary) hover:cursor-pointer w-full"
      style={{
        border: props.active ? "1px solid var(--color-fg-tertiary" : "",
      }}
      onClick={props.onClick}
    >
      <img src={props.img} alt={props.text} />
      <div className="flex justify-between items-center w-full">
        <div className="flex flex-col items-start text-start">
          {props.text}
          {props.subtext ? (
            <>
              <br />
              <span className="color-fg-secondary">{props.subtext}</span>
            </>
          ) : (
            ""
          )}
        </div>
        {props.active ? (
          <div className="px-2">
            <Check size={24} />
          </div>
        ) : (
          ""
        )}
      </div>
    </button>
  );
}
