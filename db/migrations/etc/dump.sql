-- +goose Up
-- +goose StatementBegin

--
-- PostgreSQL database dump
--

-- Dumped from database version 16.9 (Debian 16.9-1.pgdg120+1)
-- Dumped by pg_dump version 16.4 (Debian 16.4-1.pgdg120+1)

-- Started on 2025-06-11 14:30:57 UTC

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- TOC entry 2 (class 3079 OID 16511)
-- Name: pg_trgm; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;


--
-- TOC entry 3536 (class 0 OID 0)
-- Dependencies: 2
-- Name: EXTENSION pg_trgm; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pg_trgm IS 'text similarity measurement and index searching based on trigrams';


--
-- TOC entry 921 (class 1247 OID 16885)
-- Name: role; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.role AS ENUM (
    'admin',
    'user'
);


ALTER TYPE public.role OWNER TO postgres;

--
-- TOC entry 269 (class 1255 OID 16963)
-- Name: delete_orphan_releases(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.delete_orphan_releases() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM artist_releases
        WHERE release_id = OLD.release_id
    ) THEN
        DELETE FROM releases WHERE id = OLD.release_id;
    END IF;
    RETURN NULL;
END;
$$;


ALTER FUNCTION public.delete_orphan_releases() OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 231 (class 1259 OID 16901)
-- Name: api_keys; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.api_keys (
    id integer NOT NULL,
    key text NOT NULL,
    user_id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    label text
);


ALTER TABLE public.api_keys OWNER TO postgres;

--
-- TOC entry 230 (class 1259 OID 16900)
-- Name: api_keys_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.api_keys ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.api_keys_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 220 (class 1259 OID 16402)
-- Name: artist_aliases; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.artist_aliases (
    artist_id integer NOT NULL,
    alias text NOT NULL,
    source text NOT NULL,
    is_primary boolean NOT NULL
);


ALTER TABLE public.artist_aliases OWNER TO postgres;

--
-- TOC entry 227 (class 1259 OID 16839)
-- Name: artist_releases; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.artist_releases (
    artist_id integer NOT NULL,
    release_id integer NOT NULL
);


ALTER TABLE public.artist_releases OWNER TO postgres;

--
-- TOC entry 223 (class 1259 OID 16469)
-- Name: artist_tracks; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.artist_tracks (
    artist_id integer NOT NULL,
    track_id integer NOT NULL
);


ALTER TABLE public.artist_tracks OWNER TO postgres;

--
-- TOC entry 219 (class 1259 OID 16393)
-- Name: artists; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.artists (
    id integer NOT NULL,
    musicbrainz_id uuid,
    image text,
    image_source text
);


ALTER TABLE public.artists OWNER TO postgres;

--
-- TOC entry 218 (class 1259 OID 16392)
-- Name: artists_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.artists ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.artists_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 235 (class 1259 OID 16980)
-- Name: artists_with_name; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.artists_with_name AS
 SELECT a.id,
    a.musicbrainz_id,
    a.image,
    a.image_source,
    aa.alias AS name
   FROM (public.artists a
     JOIN public.artist_aliases aa ON ((aa.artist_id = a.id)))
  WHERE (aa.is_primary = true);


ALTER VIEW public.artists_with_name OWNER TO postgres;

--
-- TOC entry 217 (class 1259 OID 16386)
-- Name: goose_db_version; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.goose_db_version (
    id integer NOT NULL,
    version_id bigint NOT NULL,
    is_applied boolean NOT NULL,
    tstamp timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.goose_db_version OWNER TO postgres;

--
-- TOC entry 216 (class 1259 OID 16385)
-- Name: goose_db_version_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.goose_db_version ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.goose_db_version_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 224 (class 1259 OID 16485)
-- Name: listens; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.listens (
    track_id integer NOT NULL,
    listened_at timestamp with time zone NOT NULL,
    client text,
    user_id integer NOT NULL
);


ALTER TABLE public.listens OWNER TO postgres;

--
-- TOC entry 232 (class 1259 OID 16916)
-- Name: release_aliases; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.release_aliases (
    release_id integer NOT NULL,
    alias text NOT NULL,
    source text NOT NULL,
    is_primary boolean NOT NULL
);


ALTER TABLE public.release_aliases OWNER TO postgres;

--
-- TOC entry 226 (class 1259 OID 16825)
-- Name: releases; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.releases (
    id integer NOT NULL,
    musicbrainz_id uuid,
    image uuid,
    various_artists boolean DEFAULT false NOT NULL,
    image_source text
);


ALTER TABLE public.releases OWNER TO postgres;

--
-- TOC entry 225 (class 1259 OID 16824)
-- Name: releases_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.releases ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.releases_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 236 (class 1259 OID 16984)
-- Name: releases_with_title; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.releases_with_title AS
 SELECT r.id,
    r.musicbrainz_id,
    r.image,
    r.various_artists,
    r.image_source,
    ra.alias AS title
   FROM (public.releases r
     JOIN public.release_aliases ra ON ((ra.release_id = r.id)))
  WHERE (ra.is_primary = true);


ALTER VIEW public.releases_with_title OWNER TO postgres;

--
-- TOC entry 233 (class 1259 OID 16940)
-- Name: sessions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.sessions (
    id uuid NOT NULL,
    user_id integer NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    expires_at timestamp without time zone NOT NULL,
    persistent boolean DEFAULT false NOT NULL
);


ALTER TABLE public.sessions OWNER TO postgres;

--
-- TOC entry 234 (class 1259 OID 16967)
-- Name: track_aliases; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.track_aliases (
    track_id integer NOT NULL,
    alias text NOT NULL,
    is_primary boolean NOT NULL,
    source text NOT NULL
);


ALTER TABLE public.track_aliases OWNER TO postgres;

--
-- TOC entry 222 (class 1259 OID 16455)
-- Name: tracks; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tracks (
    id integer NOT NULL,
    musicbrainz_id uuid,
    duration integer DEFAULT 0 NOT NULL,
    release_id integer NOT NULL
);


ALTER TABLE public.tracks OWNER TO postgres;

--
-- TOC entry 221 (class 1259 OID 16454)
-- Name: tracks_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.tracks ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.tracks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 237 (class 1259 OID 16988)
-- Name: tracks_with_title; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.tracks_with_title AS
 SELECT t.id,
    t.musicbrainz_id,
    t.duration,
    t.release_id,
    ta.alias AS title
   FROM (public.tracks t
     JOIN public.track_aliases ta ON ((ta.track_id = t.id)))
  WHERE (ta.is_primary = true);


ALTER VIEW public.tracks_with_title OWNER TO postgres;

--
-- TOC entry 229 (class 1259 OID 16890)
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id integer NOT NULL,
    username text NOT NULL,
    role public.role DEFAULT 'user'::public.role NOT NULL,
    password bytea NOT NULL
);


ALTER TABLE public.users OWNER TO postgres;

--
-- TOC entry 228 (class 1259 OID 16889)
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.users ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- TOC entry 3361 (class 2606 OID 16910)
-- Name: api_keys api_keys_key_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_key_key UNIQUE (key);


--
-- TOC entry 3363 (class 2606 OID 16908)
-- Name: api_keys api_keys_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_pkey PRIMARY KEY (id);


--
-- TOC entry 3335 (class 2606 OID 16408)
-- Name: artist_aliases artist_aliases_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.artist_aliases
    ADD CONSTRAINT artist_aliases_pkey PRIMARY KEY (artist_id, alias);


--
-- TOC entry 3354 (class 2606 OID 16843)
-- Name: artist_releases artist_releases_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.artist_releases
    ADD CONSTRAINT artist_releases_pkey PRIMARY KEY (artist_id, release_id);


--
-- TOC entry 3344 (class 2606 OID 16473)
-- Name: artist_tracks artist_tracks_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.artist_tracks
    ADD CONSTRAINT artist_tracks_pkey PRIMARY KEY (artist_id, track_id);


--
-- TOC entry 3331 (class 2606 OID 16401)
-- Name: artists artists_musicbrainz_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.artists
    ADD CONSTRAINT artists_musicbrainz_id_key UNIQUE (musicbrainz_id);


--
-- TOC entry 3333 (class 2606 OID 16399)
-- Name: artists artists_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.artists
    ADD CONSTRAINT artists_pkey PRIMARY KEY (id);


--
-- TOC entry 3329 (class 2606 OID 16391)
-- Name: goose_db_version goose_db_version_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.goose_db_version
    ADD CONSTRAINT goose_db_version_pkey PRIMARY KEY (id);


--
-- TOC entry 3347 (class 2606 OID 16622)
-- Name: listens listens_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.listens
    ADD CONSTRAINT listens_pkey PRIMARY KEY (track_id, listened_at);


--
-- TOC entry 3366 (class 2606 OID 16922)
-- Name: release_aliases release_aliases_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.release_aliases
    ADD CONSTRAINT release_aliases_pkey PRIMARY KEY (release_id, alias);


--
-- TOC entry 3350 (class 2606 OID 16833)
-- Name: releases releases_musicbrainz_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.releases
    ADD CONSTRAINT releases_musicbrainz_id_key UNIQUE (musicbrainz_id);


--
-- TOC entry 3352 (class 2606 OID 16831)
-- Name: releases releases_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.releases
    ADD CONSTRAINT releases_pkey PRIMARY KEY (id);


--
-- TOC entry 3369 (class 2606 OID 16946)
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (id);


--
-- TOC entry 3371 (class 2606 OID 16973)
-- Name: track_aliases track_aliases_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.track_aliases
    ADD CONSTRAINT track_aliases_pkey PRIMARY KEY (track_id, alias);


--
-- TOC entry 3340 (class 2606 OID 16463)
-- Name: tracks tracks_musicbrainz_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tracks
    ADD CONSTRAINT tracks_musicbrainz_id_key UNIQUE (musicbrainz_id);


--
-- TOC entry 3342 (class 2606 OID 16461)
-- Name: tracks tracks_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tracks
    ADD CONSTRAINT tracks_pkey PRIMARY KEY (id);


--
-- TOC entry 3357 (class 2606 OID 16897)
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- TOC entry 3359 (class 2606 OID 16899)
-- Name: users users_username_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- TOC entry 3336 (class 1259 OID 16936)
-- Name: idx_artist_aliases_alias_trgm; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_artist_aliases_alias_trgm ON public.artist_aliases USING gin (alias public.gin_trgm_ops);


--
-- TOC entry 3337 (class 1259 OID 16495)
-- Name: idx_artist_aliases_artist_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_artist_aliases_artist_id ON public.artist_aliases USING btree (artist_id);


--
-- TOC entry 3355 (class 1259 OID 16855)
-- Name: idx_artist_releases; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_artist_releases ON public.artist_releases USING btree (artist_id, release_id);


--
-- TOC entry 3364 (class 1259 OID 16937)
-- Name: idx_release_aliases_alias_trgm; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_release_aliases_alias_trgm ON public.release_aliases USING gin (alias public.gin_trgm_ops);


--
-- TOC entry 3338 (class 1259 OID 16854)
-- Name: idx_tracks_release_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_tracks_release_id ON public.tracks USING btree (release_id);


--
-- TOC entry 3345 (class 1259 OID 16498)
-- Name: listens_listened_at_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX listens_listened_at_idx ON public.listens USING btree (listened_at);


--
-- TOC entry 3348 (class 1259 OID 16499)
-- Name: listens_track_id_listened_at_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX listens_track_id_listened_at_idx ON public.listens USING btree (track_id, listened_at);


--
-- TOC entry 3367 (class 1259 OID 16992)
-- Name: release_aliases_release_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX release_aliases_release_id_idx ON public.release_aliases USING btree (release_id) WHERE (is_primary = true);


--
-- TOC entry 3372 (class 1259 OID 16993)
-- Name: track_aliases_track_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX track_aliases_track_id_idx ON public.track_aliases USING btree (track_id) WHERE (is_primary = true);


--
-- TOC entry 3384 (class 2620 OID 16964)
-- Name: artist_releases trg_delete_orphan_releases; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_delete_orphan_releases AFTER DELETE ON public.artist_releases FOR EACH ROW EXECUTE FUNCTION public.delete_orphan_releases();


--
-- TOC entry 3380 (class 2606 OID 16957)
-- Name: api_keys api_keys_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_user_id_fkey FOREIGN KEY (id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- TOC entry 3373 (class 2606 OID 16409)
-- Name: artist_aliases artist_aliases_artist_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.artist_aliases
    ADD CONSTRAINT artist_aliases_artist_id_fkey FOREIGN KEY (artist_id) REFERENCES public.artists(id) ON DELETE CASCADE;


--
-- TOC entry 3378 (class 2606 OID 16844)
-- Name: artist_releases artist_releases_artist_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.artist_releases
    ADD CONSTRAINT artist_releases_artist_id_fkey FOREIGN KEY (artist_id) REFERENCES public.artists(id) ON DELETE CASCADE;


--
-- TOC entry 3379 (class 2606 OID 16849)
-- Name: artist_releases artist_releases_release_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.artist_releases
    ADD CONSTRAINT artist_releases_release_id_fkey FOREIGN KEY (release_id) REFERENCES public.releases(id) ON DELETE CASCADE;


--
-- TOC entry 3374 (class 2606 OID 16474)
-- Name: artist_tracks artist_tracks_artist_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.artist_tracks
    ADD CONSTRAINT artist_tracks_artist_id_fkey FOREIGN KEY (artist_id) REFERENCES public.artists(id) ON DELETE CASCADE;


--
-- TOC entry 3375 (class 2606 OID 16479)
-- Name: artist_tracks artist_tracks_track_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.artist_tracks
    ADD CONSTRAINT artist_tracks_track_id_fkey FOREIGN KEY (track_id) REFERENCES public.tracks(id) ON DELETE CASCADE;


--
-- TOC entry 3376 (class 2606 OID 16490)
-- Name: listens listens_track_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.listens
    ADD CONSTRAINT listens_track_id_fkey FOREIGN KEY (track_id) REFERENCES public.tracks(id) ON DELETE CASCADE;


--
-- TOC entry 3377 (class 2606 OID 16952)
-- Name: listens listens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.listens
    ADD CONSTRAINT listens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- TOC entry 3381 (class 2606 OID 16923)
-- Name: release_aliases release_aliases_release_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.release_aliases
    ADD CONSTRAINT release_aliases_release_id_fkey FOREIGN KEY (release_id) REFERENCES public.releases(id) ON DELETE CASCADE;


--
-- TOC entry 3382 (class 2606 OID 16947)
-- Name: sessions sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- TOC entry 3383 (class 2606 OID 16974)
-- Name: track_aliases track_aliases_track_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.track_aliases
    ADD CONSTRAINT track_aliases_track_id_fkey FOREIGN KEY (track_id) REFERENCES public.tracks(id) ON DELETE CASCADE;


-- Completed on 2025-06-11 14:30:58 UTC

--
-- PostgreSQL database dump complete
--

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin


-- +goose StatementEnd