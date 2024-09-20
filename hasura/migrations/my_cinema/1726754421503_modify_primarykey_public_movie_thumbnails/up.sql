BEGIN TRANSACTION;
ALTER TABLE "public"."movie_thumbnails" DROP CONSTRAINT "movie_thumbnails_pkey";

ALTER TABLE "public"."movie_thumbnails"
    ADD CONSTRAINT "movie_thumbnails_pkey" PRIMARY KEY ("image_url");
COMMIT TRANSACTION;
