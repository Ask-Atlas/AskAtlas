-- ASK-215: AI edit audit table.
--
-- Records every AI-driven rewrite the user requested on a study
-- guide selection: the instruction, the selected span the user
-- highlighted, and the replacement the model generated. Tracks the
-- user's accept/reject decision asynchronously via PATCH so we have
-- a long-term signal for "are AI edits actually useful?" -- the
-- ratio of `accepted=true` to total rows is the eval metric we'll
-- watch in the AI-features dashboard.
--
-- Retention: indefinite. Rows are bounded (200-2000 chars typical
-- for selection + replacement); future eval value outweighs storage.

CREATE TABLE study_guide_edits (
  id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  study_guide_id  UUID         NOT NULL REFERENCES study_guides(id) ON DELETE CASCADE,
  user_id         UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,

  instruction     TEXT         NOT NULL,
  selection_start INTEGER      NOT NULL CHECK (selection_start >= 0),
  selection_end   INTEGER      NOT NULL CHECK (selection_end >= selection_start),
  original_span   TEXT         NOT NULL,
  replacement     TEXT         NOT NULL,

  model           TEXT         NOT NULL,
  input_tokens    BIGINT       NOT NULL DEFAULT 0 CHECK (input_tokens >= 0),
  output_tokens   BIGINT       NOT NULL DEFAULT 0 CHECK (output_tokens >= 0),

  -- Accept/reject recorded asynchronously by PATCH after the user
  -- resolves the diff. NULL = still pending. Some rows stay NULL
  -- forever (user navigated away); still useful for cost + eval.
  accepted        BOOLEAN,
  accepted_at     TIMESTAMPTZ,

  created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),

  CHECK ((accepted IS NULL) = (accepted_at IS NULL))
);

CREATE INDEX idx_study_guide_edits_guide_created
  ON study_guide_edits (study_guide_id, created_at DESC);

CREATE INDEX idx_study_guide_edits_user_created
  ON study_guide_edits (user_id, created_at DESC);
