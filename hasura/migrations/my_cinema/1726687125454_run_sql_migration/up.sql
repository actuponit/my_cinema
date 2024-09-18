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
