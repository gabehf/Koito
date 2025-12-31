interface Props {
  figure: string;
  text: string;
}

export default function RewindStatText(props: Props) {
  return (
    <div className="flex items-baseline gap-1.5">
      <div className="w-23 text-end shrink-0">
        <span
          className="
            relative inline-block
            text-2xl font-semibold
          "
        >
          <span
            className="
              absolute inset-0
              -translate-x-2 translate-y-8
              bg-(--color-primary)
              z-0
              h-1
            "
            aria-hidden
          />
          <span className="relative z-1">{props.figure}</span>
        </span>
      </div>
      <span className="text-sm">{props.text}</span>
    </div>
  );
}
