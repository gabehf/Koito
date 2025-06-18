import { useState } from "react";
import { AsyncButton } from "../AsyncButton";
import { getExport } from "api/api";

export default function ExportModal() {
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState('')

    const handleExport = () => {
        setLoading(true)
        fetch(`/apis/web/v1/export`, {
            method: "GET"
        })
        .then(res => {
            if (res.ok) {
                res.blob()
                .then(blob => {
                    const url = window.URL.createObjectURL(blob)
                    const a = document.createElement("a")
                    a.href = url
                    a.download = "koito_export.json"
                    document.body.appendChild(a)
                    a.click()
                    a.remove()
                    window.URL.revokeObjectURL(url)
                    setLoading(false)
                })
            } else {
                res.json().then(r => setError(r.error))
                setLoading(false)
            }
        }).catch(err => {
            setError(err)
            setLoading(false)
        })
    }

    return (
        <div>
            <h2>Export</h2>
            <AsyncButton loading={loading} onClick={handleExport}>Export Data</AsyncButton>
            {error && <p className="error">{error}</p>}
        </div>
    )
}