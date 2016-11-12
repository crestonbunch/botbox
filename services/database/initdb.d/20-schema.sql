/* Permissions are simple keys that can be queried to decide if a user has
 * specific permissions to perform actions.
 */
CREATE TABLE permissions (
    "id"          text PRIMARY KEY,
    "description" text
);

INSERT INTO permissions (id, description) 
VALUES ('USER_AUTH', 'Allow user to authenticate via the API.');
INSERT INTO permissions (id, description) 
VALUES ('USER_PERMISSIONS', 'Allow user to change other user''s permissions.');
INSERT INTO permissions (id, description) 
VALUES ('AGENT_UPLOAD', 'Allow user to upload agents to compete.');
INSERT INTO permissions (id, description) 
VALUES ('LEAGUE_CREATE', 'Allow user to create new leagues.');
INSERT INTO permissions (id, description) 
VALUES ('SEASON_CREATE', 'Allow user to create new seasons.');

/* Permission sets are sets of permissions that can be assigned to a user to
 * distinguish them. E.g., ADMIN, MODERATOR, USER, etc.
 */
CREATE TABLE permission_sets (
    "id"          text PRIMARY KEY
);

INSERT INTO permission_sets (id) VALUES ('ADMIN');
INSERT INTO permission_sets (id) VALUES ('VERIFIED');
INSERT INTO permission_sets (id) VALUES ('UNVERIFIED');
INSERT INTO permission_sets (id) VALUES ('BANNED');

/* Map a set of permissions to a permission set to give all users in that
 * permission set ability to perform certain actions.
 */
CREATE TABLE permission_set_permissions (
    "id"      serial PRIMARY KEY,
    "permission_set" text REFERENCES permission_sets (id),
    "permission"     text REFERENCES permissions (id)
);

INSERT INTO permission_set_permissions (permission_set, permission)
VALUES ('ADMIN', 'USER_AUTH');
INSERT INTO permission_set_permissions (permission_set, permission)
VALUES ('ADMIN', 'USER_PERMISSIONS');
INSERT INTO permission_set_permissions (permission_set, permission)
VALUES ('ADMIN', 'AGENT_UPLOAD');
INSERT INTO permission_set_permissions (permission_set, permission)
VALUES ('ADMIN', 'LEAGUE_CREATE');
INSERT INTO permission_set_permissions (permission_set, permission)
VALUES ('ADMIN', 'SEASON_CREATE');
INSERT INTO permission_set_permissions (permission_set, permission)
VALUES ('VERIFIED', 'USER_AUTH');
INSERT INTO permission_set_permissions (permission_set, permission)
VALUES ('VERIFIED', 'AGENT_UPLOAD');
INSERT INTO permission_set_permissions (permission_set, permission)
VALUES ('UNVERIFIED', 'USER_AUTH');

/* A table for tracking user information.
 */
CREATE TABLE users (
    "id"          serial PRIMARY KEY,
    "username"    text UNIQUE NOT NULL,
    "fullname"    text NOT NULL DEFAULT '',
    "email"       text UNIQUE NOT NULL,
    "bio"         text NOT NULL DEFAULT '',
    "organization" text NOT NULL DEFAULT '',
    "location"    text NOT NULL DEFAULT '',
    "joined"      timestamptz NOT NULL DEFAULT now(),
    "permission_set" text REFERENCES permission_sets (id) NOT NULL DEFAULT 'UNVERIFIED'
);

/* User passwords are stored in a separate table to avoid SELECT
 * queries on the users from accidentally pulling in password hashes.
 */
CREATE TABLE passwords (
    "id"      serial PRIMARY KEY,
    "user"    integer UNIQUE REFERENCES users (id) ON DELETE CASCADE,
    "method"  text NOT NULL,
    "hash"    text NOT NULL,
    "salt"    text NOT NULL,
    "updated" timestamptz NOT NULL DEFAULT now()
);

/* A table to manage user sessions connected to the web client.
 */
CREATE TABLE user_sessions (
    "secret"      text PRIMARY KEY,
    "user"        integer REFERENCES users (id) ON DELETE CASCADE,
    "created"     timestamptz NOT NULL DEFAULT now(),
    "expires"     timestamptz NOT NULL DEFAULT now() + interval '30 minutes',
    "revoked"     boolean NOT NULL DEFAULT FALSE
);

/* A table to manage verification secrets sent to new users via email.
 */
CREATE TABLE user_verifications (
    "secret"      text PRIMARY KEY,
    "email"       text NOT NULL,
    "user"        integer REFERENCES users (id) ON DELETE CASCADE,
    "used"        boolean DEFAULT FALSE,
    "created"     timestamptz NOT NULL DEFAULT now(),
    "expires"     timestamptz NOT NULL DEFAULT now() + interval '30 days'
);


/* A table to manage password recovery requests. If a user makes a request with
 * the right secret, he/she will be able to recover the account password. */
CREATE TABLE password_recoveries (
    "secret"    text PRIMARY KEY,
    "user"      integer REFERENCES users (id) ON DELETE CASCADE,
    "used"      boolean DEFAULT FALSE,
    "created"   timestamptz NOT NULL DEFAULT now(),
    "expires"   timestamptz NOT NULL DEFAULT now() + interval '7 days'
);

/* Leagues represent a high-level game (e.g. Tron) and store the
 * information and server executable .zip file for running games.
 */
CREATE TABLE leagues (
    "id"          serial PRIMARY KEY,
    "name"        text NOT NULL,
    "description" text NOT NULL,
    "blob"        bytea NOT NULL
);

/* Seasons are time-delineated cycles of a league. Each season
 * will consist of x number of tournaments that are played
 * continuously from the start of the season to the end. At the
 * end of each season, the ranks of agents will be final. Each
 * agent plays for one season. To play in the next season, users
 * must upload a new agent.
 */
CREATE TABLE seasons (
    "id"      serial PRIMARY KEY,
    "league"  integer REFERENCES leagues (id) ON DELETE CASCADE,
    "preview" timestamptz,
    "start"   timestamptz,
    "end"     timestamptz
);

/* Tournaments are run using some system (e.g. Swiss, round-robin),
 * and track the number of rounds currently in the tournament.
 * Each match will refer to a tournament that it happened in. Each
 * tournament will refer to season that it was played for.
 * Every time a tournament is started, the current pool of agents
 * is automatically enlisted in the tournament. Tournaments will
 * run until they are over, and the next one begins with the new
 * pool of agents.
 */
CREATE TABLE tournaments (
    "id"      serial PRIMARY KEY,
    "season"  integer REFERENCES seasons (id) ON DELETE CASCADE,
    "rounds"  integer
);

/* Agents belong to the user who uploaded them. This table tracks
 * agents and their executable .zip file. Each agent competes in a certain
 * league for a certain season. Agents can be deactivated to avoid 
 * future participation in tournaments.
 */
CREATE TABLE agents (
    "id"       serial PRIMARY KEY,
    "user"     integer REFERENCES users (id) ON DELETE CASCADE,
    "name"     text,
    "season"   integer REFERENCES seasons (id) ON DELETE CASCADE,
    "active"   boolean,
    "uploaded" timestamptz,
    "blob"     bytea
);

/* A match is a game between some number of agents in a tournament.
 * The history of the match is stored as JSON so it can be played
 * back to a viewer. Each match will take place in a certain round
 * of the tournament.
 */
CREATE TABLE matches (
    "id"          serial PRIMARY KEY,
    "complete"    boolean,
    "history"     jsonb,
    "duration"    integer,
    "tournament"  integer REFERENCES tournaments (id) ON DELETE CASCADE, 
    "round"       integer
);

/* A mapping table of agents and the matches they played in. Also
 * tracks the result (win, loss, tie) of the agent, and any logs
 * to STDIN/STDOUT printed by the agent during the match.
 */
CREATE TABLE match_agents (
    "id"       serial,
    "match"    integer REFERENCES matches (id) ON DELETE CASCADE,
    "agent"    integer REFERENCES agents (id) ON DELETE CASCADE,
    "result"   integer,
    "logs"     text
);

/* A record of agents who have failed to connect during some match.
 * If an agent accumulates enough of these strikes, it may be
 * banned from playing again.
 */
CREATE TABLE agent_strikes (
    "id"       serial,
    "match"    integer REFERENCES matches (id) ON DELETE CASCADE,
    "agent"    integer REFERENCES agents (id) ON DELETE CASCADE
);
