import type { object } from "yup";
import { MOVIE_INSERT } from "~/graphql/mutations/movie";

export function useInsertMovie() {
  const { mutate, loading, onDone } = useMutation(MOVIE_INSERT, {
    // refetchQueries: [{query: CAST_QUERY}]
  });

  const executeInsert = async (values: any, thumbnails: any, schedules: any) => {
    try {
      const crew_data = values.actors.map((crew: number) => ({cast_id: crew}));
      
      const object: {
        title: any;
        description: any;
        genre: any;
        duration: any;
        published_at: any;
        director: any;
        crews: { data: any };
        movie_thumbnails: { data: any };
        schedules?: { data: any };
      } = { 
        title: values.title,
        description: values.description,
        genre: values.genre.toLowerCase().trim(),
        duration: values.duration,
        published_at: values.published_at,
        director: values.director,
        crews: {data: crew_data}, 
        movie_thumbnails: {data: thumbnails},
      };
      if (schedules && schedules.length > 0) {
        object.schedules = {data: schedules};
      }

      console.log('object', object);
      const response = await mutate({ object });
      // console.log('response', response);
    } catch (err) {
      throw createError((err as Error).message);
    }
  }
  return { loading, executeInsert, onDone };
}