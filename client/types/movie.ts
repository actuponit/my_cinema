type MovieThumbnail = {
  image_url?: string;
};

export type Rating = {
  feedback?: string;
  rating?: number;
  userByUser?: {
    first_name?: string;
    last_name?: string;
    id?: number;
  };
};

export type Schedule = {
  format?: string;
  hall?: string;
  id?: number;
  movie?: string;
  price?: number;
  start_time?: string;
};

export type Movie = {
  average_rating?: number;
  description?: string;
  director?: string;
  duration?: number;
  featured_image?: string;
  genre?: string;
  id?: number;
  movie_thumbnails?: MovieThumbnail[];
  my_director?: Cast;
  crews?: [{cast: Cast}];
  published_at?: string;
  ratings?: Rating[];
  title?: string;
  total_rating?: number;
  schedules?: Schedule[];
};