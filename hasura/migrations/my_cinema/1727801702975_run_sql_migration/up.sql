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

CREATE OR REPLACE FUNCTION update_movie_ratings_for_delete()
RETURNS TRIGGER AS $$
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
EXECUTE PROCEDURE update_movie_ratings_for_delete();

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
        WHERE NEW.hall = hall AND NEW.start_time BETWEEN start_time AND start_time + INTERVAL '2 hours'
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


-- Create a trigger on the
CREATE OR REPLACE FUNCTION update_featured_image()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE movies
    SET featured_image = NEW.image_url
    WHERE id = NEW.movie_id AND featured_image IS NULL;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION delete_featured_image()
RETURNS TRIGGER AS $$
BEGIN
    -- Check if the deleted image is currently set as the featured_image in the movies table
    IF EXISTS (SELECT 1 FROM movies WHERE featured_image = OLD.image_url) THEN
        -- Update the featured_image to NULL or choose another valid thumbnail
        UPDATE movies
        SET featured_image = (
            SELECT id FROM movie_thumbnails
            WHERE id = OLD.movie_id AND image_url != OLD.image_url
            LIMIT 1
        )
        WHERE featured_image = OLD.image_url;
    END IF;

    RETURN OLD;
END;
$$ LANGUAGE plpgsql;


CREATE OR REPLACE TRIGGER update_featured_image_trigger
BEFORE INSERT OR UPDATE ON movie_thumbnails
FOR EACH ROW
EXECUTE PROCEDURE update_featured_image();

CREATE OR REPLACE TRIGGER update_featured_image_on_delete
BEFORE DELETE ON movie_thumbnails
FOR EACH ROW
EXECUTE FUNCTION delete_featured_image();
