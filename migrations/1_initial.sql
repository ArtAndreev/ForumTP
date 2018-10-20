-- +migrate Up
CREATE TABLE IF NOT EXISTS forum_user (
    forum_user_id serial PRIMARY KEY,
    nickname varchar(64) UNIQUE,
    fullname varchar(128) NOT NULL,
    email varchar(64) UNIQUE NOT NULL,
    about text
);

CREATE TABLE IF NOT EXISTS forum (
    forum_id serial PRIMARY KEY,
    title varchar(128) NOT NULL,
    slug varchar(64) UNIQUE NOT NULL,
    forum_user integer REFERENCES forum_user NOT NULL,
    threads integer DEFAULT 0,
    posts integer DEFAULT 0
);

CREATE TABLE IF NOT EXISTS thread (
    thread_id serial PRIMARY KEY,
    forum integer REFERENCES forum NOT NULL,
    slug varchar(64) UNIQUE,
    title varchar(128) NOT NULL,
    author integer REFERENCES forum_user NOT NULL,
    created timestamp with time zone DEFAULT now(),
    message text NOT NULL,
    votes integer DEFAULT 0
);

CREATE TABLE IF NOT EXISTS post (
    post_id serial PRIMARY KEY,
    forum integer REFERENCES forum NOT NULL,
    thread integer REFERENCES thread NOT NULL,
    parent integer DEFAULT 0,
    author integer REFERENCES forum_user NOT NULL,
    created timestamp with time zone DEFAULT now(),
    is_edited boolean DEFAULT FALSE NOT NULL,
    message text NOT NULL
);

CREATE TABLE IF NOT EXISTS vote (
    vote_id serial PRIMARY KEY,
    nickname integer REFERENCES forum_user UNIQUE NOT NULL,
    voice integer NOT NULL,
    CONSTRAINT vote_constraint CHECK (voice IN (-1, 1))
);

ALTER DATABASE docker SET timezone TO 'UTC-3';

-- +migrate Down
DROP TABLE vote;
DROP TABLE post;
DROP TABLE thread;
DROP TABLE forum;
DROP TABLE forum_user;

ALTER DATABASE docker SET timezone TO 'UTC';
