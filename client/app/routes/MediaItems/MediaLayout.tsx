import React, { useEffect, useState } from "react";
import { average } from "color.js";
import { imageUrl, type SearchResponse } from "api/api";
import ImageDropHandler from "~/components/ImageDropHandler";
import { Edit, ImageIcon, Merge, Plus, Trash } from "lucide-react";
import { useAppContext } from "~/providers/AppProvider";
import MergeModal from "~/components/modals/MergeModal";
import ImageReplaceModal from "~/components/modals/ImageReplaceModal";
import DeleteModal from "~/components/modals/DeleteModal";
import RenameModal from "~/components/modals/EditModal/EditModal";
import EditModal from "~/components/modals/EditModal/EditModal";
import AddListenModal from "~/components/modals/AddListenModal";

export type MergeFunc = (from: number, to: number, replaceImage: boolean) => Promise<Response>
export type MergeSearchCleanerFunc = (r: SearchResponse, id: number) => SearchResponse

interface Props {
    type: "Track" | "Album" | "Artist" 
    title: string 
    img: string
    id: number
    musicbrainzId: string
    imgItemId: number
    mergeFunc: MergeFunc
    mergeCleanerFunc: MergeSearchCleanerFunc
    children: React.ReactNode
    subContent: React.ReactNode
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
        average(imageUrl(props.img, 'small'), { amount: 1 }).then((color) => {
        setBgColor(`rgba(${color[0]},${color[1]},${color[2]},0.4)`);
        });
    }, [props.img]);

    const replaceImageCallback = () => {
        window.location.reload()
    }

    const title = `${props.title} - Koito`

    const mobileIconSize = 22
    const normalIconSize = 30

    let vw = Math.max(document.documentElement.clientWidth || 0, window.innerWidth || 0)

    let iconSize = vw > 768 ? normalIconSize : mobileIconSize

    return (
        <main
        className="w-full flex flex-col flex-grow"
        style={{
            background: `linear-gradient(to bottom, ${bgColor}, var(--color-bg) 700px)`,
            transition: '1000',
        }}
        >
        <ImageDropHandler itemType={props.type.toLowerCase() === 'artist' ? 'artist' : 'album'} onComplete={replaceImageCallback} />
        <title>{title}</title>
        <meta property="og:title" content={title} />
        <meta
        name="description"
        content={title}
        />
            <div className="w-19/20 mx-auto pt-12">
                <div className="flex gap-8 flex-wrap md:flex-nowrap relative">
                    <div className="flex flex-col justify-around">
                        <img style={{zIndex: 5}} src={imageUrl(props.img, "large")} alt={props.title} className="md:min-w-[385px] w-[220px] h-auto shadow-(--color-shadow) shadow-lg" />
                    </div>
                    <div className="flex flex-col items-start">
                        <h3>{props.type}</h3>
                        <h1>{props.title}</h1>
                        {props.subContent}
                    </div>
                    { user &&
                    <div className="absolute left-1 sm:right-1 sm:left-auto -top-9 sm:top-1 flex gap-3 items-center">
                        { props.type === "Track" &&
                        <>
                            <button title="Add Listen" className="hover:cursor-pointer" onClick={() => setAddListenModalOpen(true)}><Plus size={iconSize} /></button>
                            <AddListenModal open={addListenModalOpen} setOpen={setAddListenModalOpen} trackid={props.id} />
                        </>
                        }
                        <button title="Edit Item" className="hover:cursor-pointer" onClick={() => setRenameModalOpen(true)}><Edit size={iconSize} /></button>
                        <button title="Replace Image" className="hover:cursor-pointer" onClick={() => setImageModalOpen(true)}><ImageIcon size={iconSize} /></button>
                        <button title="Merge Items" className="hover:cursor-pointer" onClick={() => setMergeModalOpen(true)}><Merge size={iconSize} /></button>
                        <button title="Delete Item" className="hover:cursor-pointer" onClick={() => setDeleteModalOpen(true)}><Trash size={iconSize} /></button>
                        <EditModal open={renameModalOpen} setOpen={setRenameModalOpen} type={props.type.toLowerCase()} id={props.id}/>
                        <ImageReplaceModal open={imageModalOpen} setOpen={setImageModalOpen} id={props.imgItemId} musicbrainzId={props.musicbrainzId} type={props.type === "Track" ? "Album" : props.type} />
                        <MergeModal currentTitle={props.title} mergeFunc={props.mergeFunc} mergeCleanerFunc={props.mergeCleanerFunc} type={props.type} currentId={props.id} open={mergeModalOpen} setOpen={setMergeModalOpen} />
                        <DeleteModal open={deleteModalOpen} setOpen={setDeleteModalOpen} title={props.title} id={props.id} type={props.type} />
                    </div>
                    }
                </div>
                {props.children}
            </div>
        </main>
    );
}
