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
-- Name: ltree; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS ltree WITH SCHEMA public;


--
-- Name: EXTENSION ltree; Type: COMMENT; Schema: -; Owner:
--

COMMENT ON EXTENSION ltree IS 'data type for hierarchical tree-like structures';


--
-- Name: post_before_insert_func(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.post_before_insert_func() RETURNS trigger
    LANGUAGE plpgsql
AS $$
declare
    _left_key integer;
    _right_key integer;
    _tree integer;
begin
    if new.parent = 0 then
        new.left_key := 1;
        new.right_key := 2;
        new.tree := nextval('tree');
    else
        select left_key, right_key, tree
        into _left_key, _right_key, _tree
        from post
        where id = new.parent;

        update post
        set left_key = left_key + 2
        where left_key > _right_key and tree = _tree;

        update post
        set right_key = right_key + 2
        where right_key >= _right_key and tree = _tree;

        new.left_key := _right_key;
        new.right_key := _right_key + 1;
        new.tree := _tree;
    end if;

    return new;
end
$$;


ALTER FUNCTION public.post_before_insert_func() OWNER TO postgres;

--
-- Name: post_user_add_func(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.post_user_add_func() RETURNS trigger
    LANGUAGE plpgsql
AS $$
begin
    insert into forum_user (forum, "user")
    values (new.forum, new.author)
    on conflict do nothing;

    return new;
end
$$;


ALTER FUNCTION public.post_user_add_func() OWNER TO postgres;

--
-- Name: posts_add_func(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.posts_add_func() RETURNS trigger
    LANGUAGE plpgsql
AS $$
begin
    if (tg_op = 'INSERT') then
        update forum set posts = posts + 1 where slug = new.forum;
        return new;
    elsif (tg_op = 'DELETE') then
        update forum set posts = posts - 1 where slug = old.forum;
        return old;
    end if;
    return null;
end
$$;


ALTER FUNCTION public.posts_add_func() OWNER TO postgres;

--
-- Name: thread_user_add_func(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.thread_user_add_func() RETURNS trigger
    LANGUAGE plpgsql
AS $$
begin
    insert into forum_user (forum, "user")
    values (new.forum, new.author)
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
                              slug character varying(256) NOT NULL,
                              title character varying(256) NOT NULL,
                              posts integer DEFAULT 0,
                              threads integer DEFAULT 0,
                              "user" character varying(64) NOT NULL
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
                                   forum character varying(256) NOT NULL,
                                   "user" character varying(64) NOT NULL
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
                             left_key integer NOT NULL,
                             right_key integer NOT NULL,
                             tree integer NOT NULL,
                             forum character varying(256) NOT NULL,
                             author character varying(64) NOT NULL
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
                               slug character varying(256),
                               title character varying(128) NOT NULL,
                               votes integer DEFAULT 0,
                               author character varying(64) NOT NULL,
                               forum character varying(256) NOT NULL
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
                               email character varying(256) NOT NULL,
                               fullname character varying(128) NOT NULL,
                               nickname character varying(64)
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
-- Name: forum_user forum_user_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.forum_user
    ADD CONSTRAINT forum_user_unique UNIQUE (forum, "user");


--
-- Name: user nickname_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public."user"
    ADD CONSTRAINT nickname_unique UNIQUE (nickname);


--
-- Name: post post_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.post
    ADD CONSTRAINT post_pkey PRIMARY KEY (id);


--
-- Name: forum slug_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.forum
    ADD CONSTRAINT slug_unique UNIQUE (slug);


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
-- Name: forum_lower_slug_key; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX forum_lower_slug_key ON public.forum USING btree (lower((slug)::text));


--
-- Name: forum_user_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX forum_user_idx ON public.forum_user USING btree (forum, "user");


--
-- Name: index_vote_thread_user; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX index_vote_thread_user ON public.vote USING btree (thread, "user");


--
-- Name: post_left_key_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX post_left_key_idx ON public.post USING btree (left_key);


--
-- Name: post_right_key_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX post_right_key_idx ON public.post USING btree (right_key);


--
-- Name: post_tree_left_key_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX post_tree_left_key_idx ON public.post USING btree (tree, left_key);


--
-- Name: thread_lower_slug_key; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX thread_lower_slug_key ON public.thread USING btree (lower((slug)::text)) WHERE ((slug IS NOT NULL) AND ((slug)::text <> ''::text));


--
-- Name: user_lower_email_key; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX user_lower_email_key ON public."user" USING btree (lower((email)::text));


--
-- Name: user_lower_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX user_lower_idx ON public."user" USING btree (lower((nickname)::text) COLLATE "C");


--
-- Name: user_lower_nickname_key; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX user_lower_nickname_key ON public."user" USING btree (lower((nickname)::text));


--
-- Name: vote_user_and_thread_key; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX vote_user_and_thread_key ON public.vote USING btree ("user", thread);


--
-- Name: post post_before_insert; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER post_before_insert BEFORE INSERT ON public.post FOR EACH ROW EXECUTE PROCEDURE public.post_before_insert_func();


--
-- Name: post post_user_add; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER post_user_add AFTER INSERT ON public.post FOR EACH ROW EXECUTE PROCEDURE public.post_user_add_func();


--
-- Name: post posts_add; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER posts_add AFTER INSERT OR UPDATE ON public.post FOR EACH ROW EXECUTE PROCEDURE public.posts_add_func();


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
-- Name: thread author_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.thread
    ADD CONSTRAINT author_fkey FOREIGN KEY (author) REFERENCES public."user"(nickname);


--
-- Name: post author_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.post
    ADD CONSTRAINT author_fkey FOREIGN KEY (author) REFERENCES public."user"(nickname);


--
-- Name: thread forum_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.thread
    ADD CONSTRAINT forum_fkey FOREIGN KEY (forum) REFERENCES public.forum(slug);


--
-- Name: post forum_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.post
    ADD CONSTRAINT forum_fkey FOREIGN KEY (forum) REFERENCES public.forum(slug);


--
-- Name: forum_user forum_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.forum_user
    ADD CONSTRAINT forum_fkey FOREIGN KEY (forum) REFERENCES public.forum(slug);


--
-- Name: post post_thread_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.post
    ADD CONSTRAINT post_thread_fkey FOREIGN KEY (thread) REFERENCES public.thread(id);


--
-- Name: forum user_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.forum
    ADD CONSTRAINT user_fkey FOREIGN KEY ("user") REFERENCES public."user"(nickname);


--
-- Name: forum_user user_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.forum_user
    ADD CONSTRAINT user_fkey FOREIGN KEY ("user") REFERENCES public."user"(nickname);


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
