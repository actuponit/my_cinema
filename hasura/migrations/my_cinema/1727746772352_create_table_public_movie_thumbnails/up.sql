CREATE TABLE "public"."movie_thumbnails" ("movie_id" integer NOT NULL, "image_url" text NOT NULL, PRIMARY KEY ("image_url") , FOREIGN KEY ("movie_id") REFERENCES "public"."movies"("id") ON UPDATE cascade ON DELETE cascade, UNIQUE ("image_url"));