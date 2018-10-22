-- +migrate Up
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS forum_user (
    forum_user_id serial PRIMARY KEY,
    nickname citext UNIQUE,
    fullname varchar(128) NOT NULL,
    email citext UNIQUE NOT NULL,
    about text
);

CREATE TABLE IF NOT EXISTS forum (
    forum_id serial PRIMARY KEY,
    title varchar(128) NOT NULL,
    slug citext UNIQUE NOT NULL,
    forum_user integer REFERENCES forum_user NOT NULL,
    threads integer DEFAULT 0,
    posts integer DEFAULT 0
);

CREATE TABLE IF NOT EXISTS thread (
    thread_id serial PRIMARY KEY,
    forum integer REFERENCES forum NOT NULL,
    slug citext UNIQUE,
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
    nickname integer REFERENCES forum_user UNIQUE NOT NULL,
    thread integer REFERENCES thread NOT NULL,
    voice integer NOT NULL,
    CONSTRAINT vote_constraint CHECK (voice IN (-1, 1)),
    CONSTRAINT vote_unique_all UNIQUE (nickname, thread)
);

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION recount_vote_value() RETURNS TRIGGER AS $recount_vote_value$
    BEGIN
        IF (TG_OP = 'INSERT') THEN
            UPDATE thread SET votes = votes + NEW.voice WHERE thread_id = NEW.thread;
            RETURN NEW;
        ELSIF (TG_OP = 'UPDATE') THEN
            IF OLD.voice <> NEW.voice THEN 
                UPDATE thread SET votes = votes + NEW.voice * 2 WHERE thread_id = NEW.thread;
            END IF;
            RETURN NEW;
        END IF;
        RETURN NULL;
        UPDATE thread SET votes = votes + 1 WHERE thread_id = NEW.thread;
    END;
$recount_vote_value$ LANGUAGE plpgsql;
-- +migrate StatementEnd

CREATE TRIGGER recount_vote_value AFTER INSERT OR UPDATE ON vote 
FOR EACH ROW EXECUTE PROCEDURE recount_vote_value();

ALTER DATABASE docker SET timezone TO 'UTC-3';

-- +migrate Down
DROP TABLE IF EXISTS vote;
DROP TABLE IF EXISTS post;
DROP TABLE IF EXISTS thread;
DROP TABLE IF EXISTS forum;
DROP TABLE IF EXISTS forum_user;

DROP EXTENSION IF EXISTS citext;

DROP TRIGGER IF EXISTS increment_vote_value ON vote;

DROP FUNCTION IF EXISTS recount_vote_value();

ALTER DATABASE docker SET timezone TO 'UTC';
