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
