import { useState } from "react";
import { Modal } from "./Modal";
import { AsyncButton } from "../AsyncButton";
import { submitListen } from "api/api";
import { useNavigate } from "react-router";

interface Props {
    open: boolean 
    setOpen: Function
    trackid: number
}

export default function AddListenModal({ open, setOpen, trackid }: Props) {
    const [ts, setTS] = useState<Date>(new Date);
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState('')
    const navigate = useNavigate()

    const close = () => {
        setOpen(false)
    }

    const submit = () => {
        setLoading(true)
        submitListen(trackid.toString(), ts)
        .then(r => {
            if(r.ok) {
                setLoading(false)
                navigate(0)
            } else {
                r.json().then(r => setError(r.error))
                setLoading(false)
            }
        })
    }

    const formatForDatetimeLocal = (d: Date) => {
        const pad = (n: number) => n.toString().padStart(2, "0");
        return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
    };

    return (
        <Modal isOpen={open} onClose={close}>
            <h2>Add Listen</h2>
            <div className="flex flex-col items-center gap-4">
                <input
                    type="datetime-local"
                    className="w-full mx-auto fg bg rounded p-2"
                    value={formatForDatetimeLocal(ts)}
                    onChange={(e) => setTS(new Date(e.target.value))}
                />
                <AsyncButton loading={loading} onClick={submit}>Submit</AsyncButton>
                <p className="error">{error}</p>
            </div>
        </Modal>
    )
}
