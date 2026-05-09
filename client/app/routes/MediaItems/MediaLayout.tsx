import React, { useEffect, useState } from "react";
import { average } from "color.js";
import { type ImageList, type SearchResponse } from "api/api";
import ImageDropHandler from "~/components/ImageDropHandler";
import { Edit, ImageIcon, Merge, Plus, Trash } from "lucide-react";
import { useAppContext } from "~/providers/AppProvider";
import MergeModal from "~/components/modals/MergeModal";
import ImageReplaceModal from "~/components/modals/ImageReplaceModal";
import DeleteModal from "~/components/modals/DeleteModal";
import EditModal from "~/components/modals/EditModal/EditModal";
import AddListenModal from "~/components/modals/AddListenModal";
import MbzIcon from "~/components/icons/MbzIcon";
import { Link } from "react-router";

export type MergeFunc = (
  from: number,
  to: number,
  replaceImage: boolean
) => Promise<Response>;
export type MergeSearchCleanerFunc = (
  r: SearchResponse,
  id: number
) => SearchResponse;

interface Props {
  type: "Track" | "Album" | "Artist";
  title: string;
  img: ImageList;
  id: number;
  rank: number;
  musicbrainzId: string;
  imgItemId: number;
  mergeFunc: MergeFunc;
  mergeCleanerFunc: MergeSearchCleanerFunc;
  children: React.ReactNode;
  subContent: React.ReactNode;
}

export default function MediaLayout(props: Props) {
  const [bgColor, setBgColor] = useState<string>("(--color-bg)");
  const [mergeModalOpen, setMergeModalOpen] = useState(false);
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const [imageModalOpen, setImageModalOpen] = useState(false);
  const [renameModalOpen, setRenameModalOpen] = useState(false);
  const [addListenModalOpen, setAddListenModalOpen] = useState(false);
  const { user } = useAppContext();

  useEffect(() => {
    average(props.img.xs, { amount: 1 }).then((color) => {
      setBgColor(`rgba(${color[0]},${color[1]},${color[2]},0.2)`);
    });
  }, [props.img]);

  const replaceImageCallback = () => {
    window.location.reload();
  };

  const title = `${props.title} - Koito`;

  const mobileIconSize = 22;
  const normalIconSize = 30;

  let vw = Math.max(
    document.documentElement.clientWidth || 0,
    window.innerWidth || 0
  );

  let iconSize = vw > 768 ? normalIconSize : mobileIconSize;

  console.log("MBZ:", props.musicbrainzId);

  const headersizeclass =
    props.title.length > 28 ? "text-2xl lg:text-5xl" : "text-3xl lg:text-7xl";

  return (
    <main
      className="w-full flex flex-col flex-grow"
      style={{
        background: `linear-gradient(to bottom, ${bgColor}, var(--color-bg) 700px)`,
        transition: "1000",
      }}
    >
      <ImageDropHandler
        itemType={props.type.toLowerCase() === "artist" ? "artist" : "album"}
        onComplete={replaceImageCallback}
      />
      <title>{title}</title>
      <meta property="og:title" content={title} />
      <meta name="description" content={title} />
      <div className="mx-auto w-19/20 sm:w-auto sm:ml-18 pt-12">
        <div className="flex gap-8 flex-wrap md:flex-nowrap relative sm:mb-6">
          <div className="flex flex-col justify-around">
            <img
              style={{ zIndex: 5 }}
              src={props.img.large}
              alt={props.title}
              className="md:min-w-[385px] rounded-(--border-radius) border w-[280px] h-auto"
            />
          </div>
          <div className="flex flex-col items-start">
            <h3>{props.type}</h3>
            <div className="flex">
              <h5
                className={`mt-2 sm:mt-4 mb-1 sm:mb-3 font-semibold ${headersizeclass}`}
              >
                {props.title}
                <span className="text-xl font-medium text-(--color-fg-secondary) pl-2">
                  {props.rank !== 0 && "#" + props.rank}
                </span>
              </h5>
            </div>
            {props.subContent}
          </div>
          <div className="absolute left-1 sm:right-10 sm:left-auto -top-9 sm:top-0 flex gap-3 items-center">
            {props.musicbrainzId && (
              <Link
                title="View on MusicBrainz"
                target="_blank"
                to={`https://musicbrainz.org/${props.type.toLowerCase()}/${
                  props.musicbrainzId
                }`}
              >
                <MbzIcon size={iconSize} hover />
              </Link>
            )}
            {user && (
              <>
                {props.type === "Track" && (
                  <>
                    <button
                      title="Add Listen"
                      className="hover:cursor-pointer"
                      onClick={() => setAddListenModalOpen(true)}
                    >
                      <Plus size={iconSize} />
                    </button>
                    <AddListenModal
                      open={addListenModalOpen}
                      setOpen={setAddListenModalOpen}
                      trackid={props.id}
                    />
                  </>
                )}
                <button
                  title="Edit Item"
                  className="hover:cursor-pointer"
                  onClick={() => setRenameModalOpen(true)}
                >
                  <Edit size={iconSize} />
                </button>

                {props.type !== "Track" && (
                  <button
                    title="Replace Image"
                    className="hover:cursor-pointer"
                    onClick={() => setImageModalOpen(true)}
                  >
                    <ImageIcon size={iconSize} />
                  </button>
                )}
                <button
                  title="Merge Items"
                  className="hover:cursor-pointer"
                  onClick={() => setMergeModalOpen(true)}
                >
                  <Merge size={iconSize} />
                </button>
                <button
                  title="Delete Item"
                  className="hover:cursor-pointer"
                  onClick={() => setDeleteModalOpen(true)}
                >
                  <Trash size={iconSize} />
                </button>
                <EditModal
                  open={renameModalOpen}
                  setOpen={setRenameModalOpen}
                  type={props.type.toLowerCase()}
                  id={props.id}
                />
                <ImageReplaceModal
                  open={imageModalOpen}
                  setOpen={setImageModalOpen}
                  id={props.imgItemId}
                  musicbrainzId={props.musicbrainzId}
                  type={props.type === "Track" ? "Album" : props.type}
                />
                <MergeModal
                  currentTitle={props.title}
                  mergeFunc={props.mergeFunc}
                  mergeCleanerFunc={props.mergeCleanerFunc}
                  type={props.type}
                  currentId={props.id}
                  open={mergeModalOpen}
                  setOpen={setMergeModalOpen}
                />
                <DeleteModal
                  open={deleteModalOpen}
                  setOpen={setDeleteModalOpen}
                  title={props.title}
                  id={props.id}
                  type={props.type}
                />
              </>
            )}
          </div>
        </div>
        {props.children}
      </div>
    </main>
  );
}
