--
-- PostgreSQL database dump
--

-- Dumped from database version 15.6
-- Dumped by pg_dump version 15.13 (Debian 15.13-0+deb12u1)

CREATE TYPE public.daily_roster_position AS ENUM (
    'PG',
    'SG',
    'SF',
    'PF',
    'C'
);


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: daily_roster; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.daily_roster (
    guild_id text NOT NULL,
    discord_player_id text NOT NULL,
    nba_player_uid uuid NOT NULL,
    date date NOT NULL,
    nickname text NOT NULL,
    "position" public.daily_roster_position NOT NULL
);


--
-- Name: discord_player_guilds; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.discord_player_guilds (
    discord_player_id text NOT NULL,
    guild_id text NOT NULL
);


--
-- Name: TABLE discord_player_guilds; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.discord_player_guilds IS 'Tracks servers the user has ever played it';


--
-- Name: is_locked; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.is_locked (
    date date NOT NULL,
    lock_time timestamp with time zone NOT NULL
);


--
-- Name: nba_players; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nba_players (
    nba_player_uid uuid NOT NULL,
    nba_player_id integer NOT NULL,
    name character varying NOT NULL,
    date character varying NOT NULL,
    against_team integer,
    dollar_value integer,
    fantasy_score double precision,
    team_id integer,
    "position" text NOT NULL,
    avg_pts real,
    avg_reb real,
    avg_stl real,
    avg_ast real,
    avg_tov real,
    avg_blk real,
    jersey_num text,
    status text,
    reb integer,
    ast smallint,
    pts smallint,
    blk smallint,
    stl smallint,
    tov smallint
);


--
-- Name: teams; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.teams (
    team_id integer NOT NULL,
    name character varying NOT NULL,
    abbr text
);


--
-- Name: teams_team_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.teams_team_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: teams_team_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.teams_team_id_seq OWNED BY public.teams.team_id;


--
-- Name: teams team_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teams ALTER COLUMN team_id SET DEFAULT nextval('public.teams_team_id_seq'::regclass);


--
-- Name: daily_roster daily_roster_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.daily_roster
    ADD CONSTRAINT daily_roster_pkey PRIMARY KEY (discord_player_id, date, "position");


--
-- Name: discord_player_guilds discord_player_guilds_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.discord_player_guilds
    ADD CONSTRAINT discord_player_guilds_pkey PRIMARY KEY (discord_player_id, guild_id);


--
-- Name: is_locked is_locked_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.is_locked
    ADD CONSTRAINT is_locked_pkey PRIMARY KEY (date);


--
-- Name: nba_players nba_players_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nba_players
    ADD CONSTRAINT nba_players_pkey PRIMARY KEY (nba_player_uid);


--
-- Name: teams teams_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT teams_pkey PRIMARY KEY (team_id);


--
-- Name: nba_players unique_player_per_day; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nba_players
    ADD CONSTRAINT unique_player_per_day UNIQUE (nba_player_id, date);


--
-- Name: daily_roster daily_roster_nba_player_uid_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.daily_roster
    ADD CONSTRAINT daily_roster_nba_player_uid_fkey FOREIGN KEY (nba_player_uid) REFERENCES public.nba_players(nba_player_uid);


--
-- Name: nba_players nba_players_against_team_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nba_players
    ADD CONSTRAINT nba_players_against_team_fkey FOREIGN KEY (against_team) REFERENCES public.teams(team_id);


--
-- Name: nba_players nba_players_team_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nba_players
    ADD CONSTRAINT nba_players_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(team_id) ON UPDATE CASCADE ON DELETE SET NULL;


--
-- Name: discord_player_guilds; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.discord_player_guilds ENABLE ROW LEVEL SECURITY;

--
-- Name: is_locked; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.is_locked ENABLE ROW LEVEL SECURITY;

--
-- PostgreSQL database dump complete
--

