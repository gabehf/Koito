import { updateMbzId } from "api/api";
import { useState } from "react";
import { AsyncButton } from "~/components/AsyncButton";

interface Props {
  type: string;
  id: number;
}

export default function UpdateMbzID({ type, id }: Props) {
  const [err, setError] = useState<string | undefined>();
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const [mbzid, setMbzid] = useState<"">();
  const [success, setSuccess] = useState("");

  const handleUpdateMbzID = () => {
    setError(undefined);
    if (input === "") {
      setError("no input");
      return;
    }
    setLoading(true);
    updateMbzId(type, id, input).then((r) => {
      if (r.ok) {
        setSuccess("successfully updated MusicBrainz ID");
      } else {
        r.json().then((r) => setError(r.error));
      }
    });
    setLoading(false);
  };

  return (
    <div className="w-full">
      <h3>Update MusicBrainz ID</h3>
      <div className="flex gap-2 w-3/5">
        <input
          type="text"
          placeholder="Update MusicBrainz ID"
          className="mx-auto fg bg rounded-md p-3 flex-grow"
          value={input}
          onChange={(e) => setInput(e.target.value)}
        />
        <AsyncButton loading={loading} onClick={handleUpdateMbzID}>
          Submit
        </AsyncButton>
      </div>
      {err && <p className="error">{err}</p>}
      {success && <p className="success">{success}</p>}
    </div>
  );
}
