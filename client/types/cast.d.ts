type Cast = {
  id: number;
  first_name: string;
  last_name: string;
  photo_url?: string;
  bio?: string;
  is_director: boolean;
  crews: [{ movie: CastMovie }]
  movies?: number | [CastMovie];
  movies_aggregate?: {
    aggregate: {
      count: number;
    };
  };
  crews_aggregate?: {
    aggregate: {
      count: number;
    };
  };
}

type CastMovie = {
  published_at: string;
  title: string;
  id?: number;
}
