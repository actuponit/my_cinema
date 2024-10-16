SET check_function_bodies = false;
CREATE TABLE public.movies (
    id integer NOT NULL,
    title text NOT NULL,
    duration integer NOT NULL,
    description text NOT NULL,
    published_at date NOT NULL,
    director integer,
    genre text NOT NULL,
    featured_image text,
    average_rating numeric DEFAULT '0'::numeric NOT NULL,
    total_rating integer DEFAULT 0 NOT NULL,
    CONSTRAINT duration_must_be_greater_than_15_minutes CHECK ((duration >= 900))
);
COMMENT ON TABLE public.movies IS 'the movie collections';
COMMENT ON COLUMN public.movies.duration IS 'duration in second';
CREATE FUNCTION public.bookmark_movie(movie_row public.movies, user_id integer) RETURNS boolean
    LANGUAGE plpgsql STABLE
    AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1
        FROM bookmarks A
        WHERE user_id = A.user AND A.movie_id = movie_row.id
    );
END;
$$;
CREATE FUNCTION public.check_and_update_featured_thumbnail() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Step 1: If the new thumbnail is featured, set all other thumbnails for the same movie to not featured
    IF NEW.featured = TRUE THEN
        UPDATE movie_thumbnails
        SET featured = FALSE
        WHERE movie == NEW.movie and image_url != NEW.image_url;
    ELSE
        -- Step 2: If the new thumbnail is not featured, check if any other thumbnail is featured
        IF NOT EXISTS (
            SELECT 1
            FROM movie_thumbnails
            WHERE movie = NEW.movie AND featured = TRUE
        ) THEN
            -- Step 3: If no other thumbnail is featured, set the current one to featured
            NEW.featured = TRUE;
        END IF;
    END IF;
    RETURN NEW;
END;
$$;
CREATE FUNCTION public.check_director() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Check if the director_id in the movies table corresponds to a user with is_director = true
    IF NOT EXISTS (
        SELECT *
        FROM casts
        WHERE id = NEW.director AND is_director = TRUE
    ) THEN
        RAISE EXCEPTION 'The assigned user is not a valid director';
    END IF;
    RETURN NEW;
END;
$$;
CREATE FUNCTION public.check_schedule_conflict() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Check if the new start_time conflicts with any existing schedule
    IF EXISTS (
        SELECT *
        FROM schedules
        WHERE NEW.hall = hall AND NEW.start_time BETWEEN start_time AND start_time + INTERVAL '2 hours'
    ) THEN
        RAISE EXCEPTION 'New schedule conflicts with an existing schedule in the 2-hour window';
    END IF;
    RETURN NEW;
END;
$$;
CREATE FUNCTION public.delete_featured_image() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Check if the deleted image is currently set as the featured_image in the movies table
    IF EXISTS (SELECT 1 FROM movies WHERE featured_image = OLD.image_url) THEN
        -- Update the featured_image to NULL or choose another valid thumbnail
        UPDATE movies
        SET featured_image = (
            SELECT image_url FROM movie_thumbnails
            WHERE movie_id = OLD.movie_id AND image_url != OLD.image_url
            LIMIT 1
        )
        WHERE featured_image = OLD.image_url;
    END IF;
    RETURN OLD;
END;
$$;
CREATE FUNCTION public.set_current_timestamp_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
  _new record;
BEGIN
  _new := NEW;
  _new."updated_at" = NOW();
  RETURN _new;
END;
$$;
CREATE FUNCTION public.update_featured_image() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    UPDATE movies
    SET featured_image = NEW.image_url
    WHERE id = NEW.movie_id AND featured_image IS NULL;
    RETURN NEW;
END;
$$;
CREATE FUNCTION public.update_movie_ratings() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  -- Calculate the new average rating and total rating for the movie
  UPDATE movies
  SET average_rating = (
      SELECT AVG(rating) FROM ratings WHERE movie = NEW.movie
  ),
  total_rating = (
      SELECT COUNT(*) FROM ratings WHERE movie = NEW.movie
  )
  WHERE id = NEW.movie;
  RETURN NEW;
END;
$$;
CREATE FUNCTION public.update_movie_ratings_for_delete() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  -- Calculate the new average rating and total rating for the movie
  UPDATE movies
  SET average_rating = (
      SELECT COALESCE(AVG(rating), 0)  FROM ratings WHERE movie = OLD.movie
  ),
  total_rating = (
      SELECT COALESCE(AVG(rating), 0)  FROM ratings WHERE movie = OLD.movie
  )
  WHERE id = OLD.movie;
  RETURN NEW;
END;
$$;
CREATE TABLE public.bookmarks (
    movie_id integer NOT NULL,
    "user" integer NOT NULL
);
CREATE TABLE public.casts (
    id integer NOT NULL,
    first_name text NOT NULL,
    last_name text NOT NULL,
    is_director boolean DEFAULT false NOT NULL,
    bio text,
    photo_url text
);
COMMENT ON TABLE public.casts IS 'The movie crew';
CREATE SEQUENCE public.casts_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.casts_id_seq OWNED BY public.casts.id;
CREATE TABLE public.crew (
    movie_id integer NOT NULL,
    cast_id integer NOT NULL
);
COMMENT ON TABLE public.crew IS 'A relationship table from a movie to multiple casts';
CREATE TABLE public.formats (
    name text NOT NULL
);
CREATE TABLE public.genra (
    name text NOT NULL
);
COMMENT ON TABLE public.genra IS 'Enum genra';
CREATE VIEW public.group_movie_by_category AS
 SELECT count(movies.id) AS count,
    movies.genre
   FROM public.movies
  GROUP BY movies.genre;
CREATE TABLE public.halls (
    name text NOT NULL
);
COMMENT ON TABLE public.halls IS 'Cinema halls';
CREATE TABLE public.movie_thumbnails (
    movie_id integer NOT NULL,
    image_url text NOT NULL
);
ALTER TABLE public.movies ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.movies_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);
CREATE TABLE public.ratings (
    movie integer NOT NULL,
    "user" integer NOT NULL,
    rating numeric NOT NULL,
    feedback text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT check_rating_between_1_and_5 CHECK (((rating > (0)::numeric) AND (rating < (6)::numeric)))
);
CREATE TABLE public.schedules (
    movie integer NOT NULL,
    id integer NOT NULL,
    start_time timestamp with time zone NOT NULL,
    hall text NOT NULL,
    format text NOT NULL,
    price numeric NOT NULL,
    CONSTRAINT start_in_future CHECK ((start_time > now()))
);
CREATE SEQUENCE public.schedules_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.schedules_id_seq OWNED BY public.schedules.id;
ALTER TABLE public.schedules ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.schedules_id_seq1
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);
CREATE TABLE public.tickets (
    schedule_id integer NOT NULL,
    user_id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    quantity integer,
    id uuid DEFAULT gen_random_uuid()
);
CREATE VIEW public.tickets_graph AS
 SELECT date(t.created_at) AS t_date,
    count(t.id) AS count,
    sum(s.price) AS total_price
   FROM (public.tickets t
     JOIN public.schedules s ON ((t.schedule_id = s.id)))
  GROUP BY (date(t.created_at))
  ORDER BY (date(t.created_at));
CREATE TABLE public.users (
    first_name text NOT NULL,
    last_name text NOT NULL,
    id integer NOT NULL,
    role text DEFAULT 'user'::text NOT NULL,
    password text NOT NULL,
    email text NOT NULL,
    CONSTRAINT email_format_check CHECK ((email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9]+\.[A-Za-z]{2,}$'::text))
);
ALTER TABLE public.users ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);
ALTER TABLE ONLY public.casts ALTER COLUMN id SET DEFAULT nextval('public.casts_id_seq'::regclass);
ALTER TABLE ONLY public.bookmarks
    ADD CONSTRAINT bookmarks_pkey PRIMARY KEY (movie_id, "user");
ALTER TABLE ONLY public.casts
    ADD CONSTRAINT casts_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.crew
    ADD CONSTRAINT crew_pkey PRIMARY KEY (movie_id, cast_id);
ALTER TABLE ONLY public.formats
    ADD CONSTRAINT formats_pkey PRIMARY KEY (name);
ALTER TABLE ONLY public.genra
    ADD CONSTRAINT genra_pkey PRIMARY KEY (name);
ALTER TABLE ONLY public.halls
    ADD CONSTRAINT halls_pkey PRIMARY KEY (name);
ALTER TABLE ONLY public.movie_thumbnails
    ADD CONSTRAINT movie_thumbnails_pkey PRIMARY KEY (image_url);
ALTER TABLE ONLY public.movies
    ADD CONSTRAINT movies_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.movies
    ADD CONSTRAINT movies_title_key UNIQUE (title);
ALTER TABLE ONLY public.ratings
    ADD CONSTRAINT ratings_pkey PRIMARY KEY (movie, "user");
ALTER TABLE ONLY public.schedules
    ADD CONSTRAINT schedules_id_key UNIQUE (id);
ALTER TABLE ONLY public.schedules
    ADD CONSTRAINT schedules_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.tickets
    ADD CONSTRAINT tickets_pkey PRIMARY KEY (schedule_id, user_id);
ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);
ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);
CREATE UNIQUE INDEX unique_ticket_id ON public.tickets USING btree (id);
CREATE TRIGGER rating_delete_trigger AFTER DELETE ON public.ratings FOR EACH ROW EXECUTE FUNCTION public.update_movie_ratings_for_delete();
CREATE TRIGGER rating_insert_update_trigger AFTER INSERT OR UPDATE ON public.ratings FOR EACH ROW EXECUTE FUNCTION public.update_movie_ratings();
CREATE TRIGGER set_public_ratings_updated_at BEFORE UPDATE ON public.ratings FOR EACH ROW EXECUTE FUNCTION public.set_current_timestamp_updated_at();
COMMENT ON TRIGGER set_public_ratings_updated_at ON public.ratings IS 'trigger to set value of column "updated_at" to current timestamp on row update';
CREATE TRIGGER set_public_tickets_updated_at BEFORE UPDATE ON public.tickets FOR EACH ROW EXECUTE FUNCTION public.set_current_timestamp_updated_at();
COMMENT ON TRIGGER set_public_tickets_updated_at ON public.tickets IS 'trigger to set value of column "updated_at" to current timestamp on row update';
CREATE TRIGGER update_featured_image_on_delete BEFORE DELETE ON public.movie_thumbnails FOR EACH ROW EXECUTE FUNCTION public.delete_featured_image();
CREATE TRIGGER update_featured_image_trigger BEFORE INSERT OR UPDATE ON public.movie_thumbnails FOR EACH ROW EXECUTE FUNCTION public.update_featured_image();
CREATE TRIGGER validate_director BEFORE INSERT OR UPDATE ON public.movies FOR EACH ROW WHEN ((new.director IS NOT NULL)) EXECUTE FUNCTION public.check_director();
CREATE TRIGGER validate_schedule_conflict BEFORE INSERT OR UPDATE ON public.schedules FOR EACH ROW EXECUTE FUNCTION public.check_schedule_conflict();
ALTER TABLE ONLY public.bookmarks
    ADD CONSTRAINT bookmarks_movie_id_fkey FOREIGN KEY (movie_id) REFERENCES public.movies(id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE ONLY public.bookmarks
    ADD CONSTRAINT bookmarks_user_fkey FOREIGN KEY ("user") REFERENCES public.users(id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE ONLY public.crew
    ADD CONSTRAINT crew_cast_id_fkey FOREIGN KEY (cast_id) REFERENCES public.casts(id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE ONLY public.crew
    ADD CONSTRAINT crew_movie_id_fkey FOREIGN KEY (movie_id) REFERENCES public.movies(id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE ONLY public.movie_thumbnails
    ADD CONSTRAINT movie_thumbnails_movie_id_fkey FOREIGN KEY (movie_id) REFERENCES public.movies(id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE ONLY public.movies
    ADD CONSTRAINT movies_director_fkey FOREIGN KEY (director) REFERENCES public.casts(id) ON UPDATE CASCADE ON DELETE SET NULL;
ALTER TABLE ONLY public.movies
    ADD CONSTRAINT movies_genra_fkey FOREIGN KEY (genre) REFERENCES public.genra(name) ON UPDATE RESTRICT ON DELETE RESTRICT;
ALTER TABLE ONLY public.ratings
    ADD CONSTRAINT ratings_movie_fkey FOREIGN KEY (movie) REFERENCES public.movies(id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE ONLY public.ratings
    ADD CONSTRAINT ratings_user_fkey FOREIGN KEY ("user") REFERENCES public.users(id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE ONLY public.schedules
    ADD CONSTRAINT schedules_format_fkey FOREIGN KEY (format) REFERENCES public.formats(name) ON UPDATE RESTRICT ON DELETE RESTRICT;
ALTER TABLE ONLY public.schedules
    ADD CONSTRAINT schedules_hall_fkey FOREIGN KEY (hall) REFERENCES public.halls(name) ON UPDATE RESTRICT ON DELETE RESTRICT;
ALTER TABLE ONLY public.schedules
    ADD CONSTRAINT schedules_movie_fkey FOREIGN KEY (movie) REFERENCES public.movies(id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE ONLY public.tickets
    ADD CONSTRAINT tickets_schedule_id_fkey FOREIGN KEY (schedule_id) REFERENCES public.schedules(id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE ONLY public.tickets
    ADD CONSTRAINT tickets_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON UPDATE CASCADE ON DELETE CASCADE;
