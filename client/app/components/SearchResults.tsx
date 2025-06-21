import { imageUrl, type SearchResponse } from "api/api"
import { useState } from "react"
import SearchResultItem from "./SearchResultItem"
import SearchResultSelectorItem from "./SearchResultSelectorItem"

interface Props {
    data?: SearchResponse
    onSelect: Function
    selectorMode?: boolean
}
export default function SearchResults({ data, onSelect, selectorMode }: Props) {
    const [selected, setSelected] = useState(0)
    const classes = "flex flex-col items-start bg rounded w-full"
    const hClasses = "pt-4 pb-2"

    const selectItem = (title: string, id: number) => {
        if (selected === id) {
            setSelected(0)
            onSelect({id: 0, title: ''})
        } else {
            setSelected(id)
            onSelect({id: id, title: title})
        }
    }

    if (!data) {
        return <></>
    }
    return (
        <div className="w-full">
            { data.artists && data.artists.length > 0 &&
            <>
            <h3 className={hClasses}>Artists</h3>
            <div className={classes}>
            {data.artists.map((artist) => (
                selectorMode ? 
                <SearchResultSelectorItem 
                    id={artist.id}
                    onClick={() => selectItem(artist.name, artist.id)}
                    text={artist.name}
                    img={imageUrl(artist.image, "small")}
                    active={selected === artist.id}
                /> : 
                <SearchResultItem 
                    to={`/artist/${artist.id}`} 
                    onClick={() => onSelect(artist.id)}
                    text={artist.name}
                    img={imageUrl(artist.image, "small")}
                />
                
            ))}
            </div>
            </>
            }
            { data.albums && data.albums.length > 0 &&
            <>
            <h3 className={hClasses}>Albums</h3>
            <div className={classes}>
            {data.albums.map((album) => (
                selectorMode ? 
                <SearchResultSelectorItem 
                    id={album.id}
                    onClick={() => selectItem(album.title, album.id)}
                    text={album.title}
                    subtext={album.is_various_artists ? "Various Artists" : album.artists[0].name}
                    img={imageUrl(album.image, "small")}
                    active={selected === album.id}
                /> : 
                <SearchResultItem 
                    to={`/album/${album.id}`} 
                    onClick={() => onSelect(album.id)}
                    text={album.title}
                    subtext={album.is_various_artists ? "Various Artists" : album.artists[0].name}
                    img={imageUrl(album.image, "small")}
                />
            ))}
            </div>
            </>
            }
            { data.tracks && data.tracks.length > 0 &&
            <>
            <h3 className={hClasses}>Tracks</h3>
            <div className={classes}>
            {data.tracks.map((track) => (
                selectorMode ? 
                <SearchResultSelectorItem 
                    id={track.id}
                    onClick={() => selectItem(track.title, track.id)}
                    text={track.title}
                    subtext={track.artists.map((a) => a.name).join(', ')}
                    img={imageUrl(track.image, "small")}
                    active={selected === track.id}
                /> : 
                <SearchResultItem 
                    to={`/track/${track.id}`} 
                    onClick={() => onSelect(track.id)}
                    text={track.title}
                    subtext={track.artists.map((a) => a.name).join(', ')}
                    img={imageUrl(track.image, "small")}
                />
            ))}
            </div>
            </>
            }
        </div>
    )
}