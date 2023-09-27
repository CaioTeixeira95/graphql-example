CREATE TABLE IF NOT EXISTS public.developers (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    first_name text NOT NULL,
    last_name text NOT NULL,
    github_url text, 
    stack text[]
);
