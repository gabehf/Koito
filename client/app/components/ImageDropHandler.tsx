import { replaceImage } from 'api/api';
import { useEffect } from 'react';

interface Props {
    itemType: string,
    onComplete: Function
}

export default function ImageDropHandler({ itemType, onComplete }: Props) {
  useEffect(() => {
    const handleDragOver = (e: DragEvent) => {
        console.log('dragover!!')
        e.preventDefault(); 
    };

    const handleDrop = async (e: DragEvent) => {
        e.preventDefault();
        if (!e.dataTransfer?.files.length) return;

        const imageFile = Array.from(e.dataTransfer.files).find(file =>
            file.type.startsWith('image/')
        );
        if (!imageFile) return;

        const formData = new FormData();
        formData.append('image', imageFile);
        const pathname = window.location.pathname;
        const segments = pathname.split('/');
        const filteredSegments = segments.filter(segment => segment !== '');
        const lastSegment = filteredSegments[filteredSegments.length - 1];
        formData.append(itemType.toLowerCase()+'_id', lastSegment)
        replaceImage(formData).then((r) => {
            if (r.status >= 200 && r.status < 300) {
                onComplete()
                console.log("Replacement image uploaded successfully")
            } else {
                r.json().then((body) => {
                    console.log(`Upload failed: ${r.statusText} - ${body}`)
                })
            }
        }).catch((err) => {
            console.log(`Upload failed: ${err}`)
        })
    };

    window.addEventListener('dragover', handleDragOver);
    window.addEventListener('drop', handleDrop);

    return () => {
        window.removeEventListener('dragover', handleDragOver);
        window.removeEventListener('drop', handleDrop);
    };
  }, []);

  return null;
}
