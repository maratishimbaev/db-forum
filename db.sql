--
-- PostgreSQL database dump
--

-- Dumped from database version 12.0 (Ubuntu 12.0-2.pgdg16.04+1)
-- Dumped by pg_dump version 12.0 (Ubuntu 12.0-2.pgdg16.04+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: citext; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA public;


--
-- Name: EXTENSION citext; Type: COMMENT; Schema: -; Owner:
--

COMMENT ON EXTENSION citext IS 'data type for case-insensitive character strings';


--
-- Name: ltree; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS ltree WITH SCHEMA public;


--
-- Name: EXTENSION ltree; Type: COMMENT; Schema: -; Owner:
--

COMMENT ON EXTENSION ltree IS 'data type for hierarchical tree-like structures';


--
-- Name: post_user_add_func(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.post_user_add_func() RETURNS trigger
    LANGUAGE plpgsql
AS $$
begin
    insert into forum_user (forum, "user")
    values (
               (select id from forum where slug = new.forum),
               (select id from "user" where nickname = new.author)
           )
    on conflict do nothing;

    return new;
end
$$;


ALTER FUNCTION public.post_user_add_func() OWNER TO postgres;

--
-- Name: thread_user_add_func(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.thread_user_add_func() RETURNS trigger
    LANGUAGE plpgsql
AS $$
begin
    insert into forum_user (forum, "user")
    values (
               (select id from forum where slug = new.forum),
               (select id from "user" where nickname = new.author)
           )
    on conflict do nothing;

    return new;
end
$$;


ALTER FUNCTION public.thread_user_add_func() OWNER TO postgres;

--
-- Name: threads_add_func(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.threads_add_func() RETURNS trigger
    LANGUAGE plpgsql
AS $$
begin
    if (tg_op = 'INSERT') then
        update forum set threads = threads + 1 where slug = new.forum;
        return new;
    elsif (tg_op = 'DELETE') then
        update forum set threads = threads - 1 where slug = old.forum;
        return old;
    end if;
    return null;
end
$$;


ALTER FUNCTION public.threads_add_func() OWNER TO postgres;

--
-- Name: votes_add_func(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.votes_add_func() RETURNS trigger
    LANGUAGE plpgsql
AS $$
begin
    raise notice 'qq';
    if (tg_op = 'INSERT') then
        update thread set votes = votes + new.voice where id = new.thread;
        raise notice 'qqq';
        return new;
    elsif (tg_op = 'UPDATE') then
        update thread set votes = votes - old.voice + new.voice where id = new.thread;
        raise notice 'qqqq';
        return new;
    end if;
    return null;
end
$$;


ALTER FUNCTION public.votes_add_func() OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: forum; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.forum (
                              id integer NOT NULL,
                              title character varying(256) NOT NULL,
                              posts integer DEFAULT 0,
                              threads integer DEFAULT 0,
                              "user" public.citext NOT NULL,
                              slug public.citext NOT NULL
);


ALTER TABLE public.forum OWNER TO postgres;

--
-- Name: forum_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.forum_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.forum_id_seq OWNER TO postgres;

--
-- Name: forum_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.forum_id_seq OWNED BY public.forum.id;


--
-- Name: forum_user; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.forum_user (
                                   "user" integer NOT NULL,
                                   forum integer NOT NULL
);


ALTER TABLE public.forum_user OWNER TO postgres;

--
-- Name: post; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.post (
                             id integer NOT NULL,
                             created timestamp with time zone,
                             is_edited boolean NOT NULL,
                             message text NOT NULL,
                             thread integer,
                             parent integer,
                             author public.citext NOT NULL,
                             path integer[],
                             forum public.citext NOT NULL
);


ALTER TABLE public.post OWNER TO postgres;

--
-- Name: post_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.post_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.post_id_seq OWNER TO postgres;

--
-- Name: post_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.post_id_seq OWNED BY public.post.id;


--
-- Name: thread; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.thread (
                               id integer NOT NULL,
                               created timestamp with time zone,
                               message text NOT NULL,
                               title character varying(128) NOT NULL,
                               votes integer DEFAULT 0,
                               author public.citext NOT NULL,
                               forum public.citext NOT NULL,
                               slug public.citext
);


ALTER TABLE public.thread OWNER TO postgres;

--
-- Name: thread_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.thread_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.thread_id_seq OWNER TO postgres;

--
-- Name: thread_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.thread_id_seq OWNED BY public.thread.id;


--
-- Name: tree; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.tree
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.tree OWNER TO postgres;

--
-- Name: user; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public."user" (
                               id integer NOT NULL,
                               about text,
                               fullname character varying(128) NOT NULL,
                               email public.citext NOT NULL,
                               nickname public.citext NOT NULL COLLATE pg_catalog."C"
);


ALTER TABLE public."user" OWNER TO postgres;

--
-- Name: user_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.user_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_id_seq OWNER TO postgres;

--
-- Name: user_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.user_id_seq OWNED BY public."user".id;


--
-- Name: vote; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.vote (
                             "user" integer NOT NULL,
                             voice integer NOT NULL,
                             thread integer NOT NULL
);


ALTER TABLE public.vote OWNER TO postgres;

--
-- Name: forum id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.forum ALTER COLUMN id SET DEFAULT nextval('public.forum_id_seq'::regclass);


--
-- Name: post id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.post ALTER COLUMN id SET DEFAULT nextval('public.post_id_seq'::regclass);


--
-- Name: thread id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.thread ALTER COLUMN id SET DEFAULT nextval('public.thread_id_seq'::regclass);


--
-- Name: user id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."user" ALTER COLUMN id SET DEFAULT nextval('public.user_id_seq'::regclass);


--
-- Name: forum forum_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.forum
    ADD CONSTRAINT forum_pkey PRIMARY KEY (id);


--
-- Name: post post_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.post
    ADD CONSTRAINT post_pkey PRIMARY KEY (id);


--
-- Name: thread thread_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.thread
    ADD CONSTRAINT thread_pkey PRIMARY KEY (id);


--
-- Name: vote unique_user_and_thread; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.vote
    ADD CONSTRAINT unique_user_and_thread UNIQUE ("user", thread);


--
-- Name: user user_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."user"
    ADD CONSTRAINT user_pkey PRIMARY KEY (id);


--
-- Name: forum_slug_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX forum_slug_idx ON public.forum USING btree (slug);


--
-- Name: forum_user_forum_and_user_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX forum_user_forum_and_user_idx ON public.forum_user USING btree (forum, "user");


--
-- Name: forum_user_forum_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX forum_user_forum_idx ON public.forum_user USING btree (forum);


--
-- Name: forum_user_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX forum_user_idx ON public.forum USING btree ("user");


--
-- Name: forum_user_user_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX forum_user_user_idx ON public.forum_user USING btree ("user");


--
-- Name: index_vote_thread_user; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX index_vote_thread_user ON public.vote USING btree (thread, "user");


--
-- Name: post_created_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX post_created_idx ON public.post USING btree (created);


--
-- Name: post_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX post_idx ON public.post USING btree (thread, id, author, created, forum, is_edited, message, parent);


--
-- Name: post_parent_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX post_parent_idx ON public.post USING btree (parent);


--
-- Name: post_path_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX post_path_idx ON public.post USING gin (path);


--
-- Name: post_thread_and_id_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX post_thread_and_id_idx ON public.post USING btree (thread, id);

ALTER TABLE public.post CLUSTER ON post_thread_and_id_idx;


--
-- Name: post_thread_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX post_thread_idx ON public.post USING btree (thread);


--
-- Name: thread_created_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX thread_created_idx ON public.thread USING btree (created);


--
-- Name: thread_forum_and_created_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX thread_forum_and_created_idx ON public.thread USING btree (forum, created);


--
-- Name: thread_forum_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX thread_forum_idx ON public.thread USING btree (forum);


--
-- Name: thread_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX thread_idx ON public.thread USING btree (forum, created, id, author, message, slug, title, forum, votes);


--
-- Name: thread_slug_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX thread_slug_idx ON public.thread USING btree (slug);


--
-- Name: user_email_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX user_email_idx ON public."user" USING btree (email);


--
-- Name: user_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX user_idx ON public."user" USING btree (nickname, email, about, fullname);


--
-- Name: user_nickname_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX user_nickname_idx ON public."user" USING btree (nickname);


--
-- Name: vote_user_and_thread_key; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX vote_user_and_thread_key ON public.vote USING btree ("user", thread);


--
-- Name: post post_user_add; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER post_user_add AFTER INSERT ON public.post FOR EACH ROW EXECUTE PROCEDURE public.post_user_add_func();


--
-- Name: thread thread_user_add; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER thread_user_add AFTER INSERT ON public.thread FOR EACH ROW EXECUTE PROCEDURE public.thread_user_add_func();


--
-- Name: thread threads_add; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER threads_add AFTER INSERT OR UPDATE ON public.thread FOR EACH ROW EXECUTE PROCEDURE public.threads_add_func();


--
-- Name: vote votes_add; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER votes_add AFTER INSERT OR UPDATE ON public.vote FOR EACH ROW EXECUTE PROCEDURE public.votes_add_func();


--
-- Name: thread forum_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.thread
    ADD CONSTRAINT forum_fkey FOREIGN KEY (forum) REFERENCES public.forum(slug);


--
-- Name: vote vote_thread_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.vote
    ADD CONSTRAINT vote_thread_fkey FOREIGN KEY (thread) REFERENCES public.thread(id);


--
-- Name: vote vote_user_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.vote
    ADD CONSTRAINT vote_user_fkey FOREIGN KEY ("user") REFERENCES public."user"(id);


--
-- PostgreSQL database dump complete
--
