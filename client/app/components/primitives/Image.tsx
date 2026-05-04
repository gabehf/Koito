interface Props {
  size: number;
  imageUrl: string;
}

export default function ActivityGrid(props: Props) {

  return <>
    <img
      src={props.imageUrl}
      className="image rounded-[10px] border"
      style={{ height: props.size, width: props.size }} />
  </>
}
