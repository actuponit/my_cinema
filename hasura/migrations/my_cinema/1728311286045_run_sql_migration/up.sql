CREATE OR REPLACE FUNCTION bookmark_movie(movie_row movies, user_id int)
RETURNS boolean AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1
        FROM bookmarks A
        WHERE A.user = user_id AND A.movie_id = movie_row.id
    );
END;
$$ LANGUAGE plpgsql STABLE;
