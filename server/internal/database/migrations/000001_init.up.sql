-- Enums ----------------------------------------------------------------------

CREATE TYPE tag_category AS ENUM ('language', 'framework', 'tool', 'domain');
CREATE TYPE guestbook_status AS ENUM ('visible', 'hidden');
CREATE TYPE analytics_event_type AS ENUM ('page_view', 'project_click', 'resume_download', 'outbound_click');
CREATE TYPE device_type AS ENUM ('desktop', 'mobile', 'tablet');

-- Projects & tags ------------------------------------------------------------

CREATE TABLE projects (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    slug            text        NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    title           text        NOT NULL CHECK (char_length(title) BETWEEN 1 AND 120),
    summary         text        NOT NULL CHECK (char_length(summary) BETWEEN 1 AND 160),
    body            text        NOT NULL DEFAULT '',          -- markdown
    cover_image_url text,
    repo_url        text,
    live_url        text,
    featured        boolean     NOT NULL DEFAULT false,
    published       boolean     NOT NULL DEFAULT false,
    sort_order      integer     NOT NULL DEFAULT 0,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

-- Public listing only ever reads published rows in display order.
CREATE INDEX projects_published_order_idx ON projects (sort_order, created_at DESC) WHERE published;

CREATE TABLE tags (
    id       uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    slug     text         NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    name     text         NOT NULL CHECK (char_length(name) BETWEEN 1 AND 40),
    category tag_category NOT NULL
);

CREATE TABLE project_tags (
    project_id uuid NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    tag_id     uuid NOT NULL REFERENCES tags (id) ON DELETE CASCADE,
    PRIMARY KEY (project_id, tag_id)
);

-- The PK covers project -> tags; this covers tag -> projects (tag filtering).
CREATE INDEX project_tags_tag_id_idx ON project_tags (tag_id);

-- Guestbook ------------------------------------------------------------------

CREATE TABLE guestbook_entries (
    id                uuid             PRIMARY KEY DEFAULT gen_random_uuid(),
    author_github_id  bigint           NOT NULL,
    author_name       text             NOT NULL CHECK (char_length(author_name) BETWEEN 1 AND 100),
    author_avatar_url text,
    message           text             NOT NULL CHECK (char_length(message) BETWEEN 1 AND 280),
    status            guestbook_status NOT NULL DEFAULT 'visible',
    created_at        timestamptz      NOT NULL DEFAULT now()
);

-- Keyset pagination over visible entries, newest first.
CREATE INDEX guestbook_entries_visible_idx ON guestbook_entries (created_at DESC, id DESC) WHERE status = 'visible';
CREATE INDEX guestbook_entries_author_idx ON guestbook_entries (author_github_id);

-- Contact --------------------------------------------------------------------

CREATE TABLE contact_messages (
    id         uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    name       text        NOT NULL CHECK (char_length(name) BETWEEN 1 AND 80),
    email      text        NOT NULL CHECK (char_length(email) BETWEEN 3 AND 254),
    message    text        NOT NULL CHECK (char_length(message) BETWEEN 10 AND 2000),
    ip_hash    text        NOT NULL,                          -- for abuse checks; never the raw IP
    read_at    timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX contact_messages_created_at_idx ON contact_messages (created_at DESC);

-- Analytics ------------------------------------------------------------------

CREATE TABLE analytics_events (
    id            bigint               GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    type          analytics_event_type NOT NULL,
    path          text                 NOT NULL CHECK (char_length(path) BETWEEN 1 AND 200),
    target        text                 CHECK (char_length(target) <= 200), -- project slug or outbound URL
    referrer_host text                 CHECK (char_length(referrer_host) <= 255),
    device        device_type          NOT NULL,
    country       text                 CHECK (country ~ '^[A-Z]{2}$'),
    visitor_hash  text                 NOT NULL,              -- daily-salted; not linkable across days
    created_at    timestamptz          NOT NULL DEFAULT now()
);

CREATE INDEX analytics_events_type_created_at_idx ON analytics_events (type, created_at);
CREATE INDEX analytics_events_path_created_at_idx ON analytics_events (path, created_at);
