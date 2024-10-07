CREATE OR REPLACE FUNCTION public.bookmark_movie(movie_row movies, hasura_session json)
 RETURNS boolean
 LANGUAGE plpgsql
 STABLE
AS $function$
BEGIN
    RETURN EXISTS (
        SELECT 1
        FROM bookmarks A
        WHERE A.movie_id = movie_row.id
    );
END;
$function$;
