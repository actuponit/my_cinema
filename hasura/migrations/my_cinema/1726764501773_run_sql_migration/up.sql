CREATE OR REPLACE FUNCTION update_movie_ratings()
RETURNS TRIGGER AS $$
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
$$ LANGUAGE plpgsql;

-- Trigger for INSERT or UPDATE
CREATE OR REPLACE TRIGGER rating_insert_update_trigger
AFTER INSERT OR UPDATE ON ratings
FOR EACH ROW
EXECUTE PROCEDURE update_movie_ratings();

-- Trigger for DELETE (since we need to update totals if a rating is removed)
CREATE OR REPLACE TRIGGER rating_delete_trigger
AFTER DELETE ON ratings
FOR EACH ROW
EXECUTE PROCEDURE update_movie_ratings();

-- Check director
CREATE OR REPLACE FUNCTION check_director()
RETURNS TRIGGER AS $$
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
$$ LANGUAGE plpgsql;


-- Create a trigger on the movies table
CREATE OR REPLACE TRIGGER validate_director
BEFORE INSERT OR UPDATE ON movies
FOR EACH ROW
WHEN (NEW.director IS NOT NULL) -- Only trigger when director_id is set
EXECUTE PROCEDURE check_director();

CREATE OR REPLACE FUNCTION check_schedule_conflict()
RETURNS TRIGGER AS $$
BEGIN
    -- Check if the new start_time conflicts with any existing schedule
    IF EXISTS (
        SELECT *
        FROM scheduleS
        WHERE NEW.start_time BETWEEN start_time AND start_time + INTERVAL '2 hours'
    ) THEN
        RAISE EXCEPTION 'New schedule conflicts with an existing schedule in the 2-hour window';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- Create a trigger on the schedule table
CREATE OR REPLACE TRIGGER validate_schedule_conflict
BEFORE INSERT OR UPDATE ON scheduleS
FOR EACH ROW
EXECUTE PROCEDURE check_schedule_conflict();

-- Only one thumbnail of a movie is featured
CREATE OR REPLACE FUNCTION check_and_update_featured_thumbnail()
RETURNS TRIGGER AS $$
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
$$ LANGUAGE plpgsql;

CREATE TRIGGER validate_featured_thumbnail
BEFORE INSERT OR UPDATE ON movie_thumbnails
FOR EACH ROW
EXECUTE FUNCTION check_and_update_featured_thumbnail();
