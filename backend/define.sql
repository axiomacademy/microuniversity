-- clear up old stuff
DROP TABLE IF EXISTS learner CASCADE;
DROP TABLE IF EXISTS module CASCADE;
DROP TABLE IF EXISTS cohort CASCADE;
DROP TABLE IF EXISTS learner_cohort CASCADE;
DROP TABLE IF EXISTS lesson CASCADE;
DROP TABLE IF EXISTS tutorial CASCADE;
DROP TABLE IF EXISTS flashcard CASCADE;
DROP TABLE IF EXISTS learner_flashcard CASCADE;
DROP TABLE IF EXISTS learner_lesson CASCADE;

--UUID support
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- learner table hols the data about the users
CREATE TABLE learner (
  email VARCHAR UNIQUE NOT NULL,
  first_name VARCHAR,
  last_name VARCHAR,
  last_completed TIMESTAMPTZ DEFAULT NOW(),
  streak INT NOT NULL DEFAULT 0,

  PRIMARY KEY (email)
);

-- module ID is codes like CS0001 etc, based on the university
CREATE TABLE module (
  module_id VARCHAR UNIQUE NOT NULL,
  title VARCHAR NOT NULL,
  description TEXT NOT NULL,

  PRIMARY KEY (module_id)
);

-- cohort_start_date is always a Monday
-- weekly tutorial starts the week before the cohort_start_date with the module orientation session
-- weekly_tutorial_day starts at 0 for Monday and 6 for Sunday
CREATE TABLE cohort (
  cohort_id uuid DEFAULT uuid_generate_v4 (),
  module VARCHAR NOT NULL,
  completed BOOL NOT NULL DEFAULT FALSE,
  start_date DATE NOT NULL,
  weekly_tutorial_day INT NOT NULL CHECK (weekly_tutorial_day >= 0 AND weekly_tutorial_day <=6),
  weekly_tutorial_time TIMETZ NOT NULL
);

CREATE TABLE learner_cohort (
  learner VARCHAR,
  cohort uuid,

  PRIMARY KEY (learner, cohort),

  CONSTRAINT fk_learner
    FOREIGN KEY (learner) REFERENCES(email),
  CONSTRAINT fk_cohort
    FOREIGN KEY (cohort) REFERENCES(cohort_id)
);

-- lesson table holds the lessons in a module
CREATE TABLE lesson (
  lesson_id uuid DEFAULT uuid_generate_v4 (),
  title VARCHAR NOT NULL,
  description TEXT NOT NULL,
  video_link VARCHAR NOT NULL,
  date_offset INT NOT NULL,
  module VARCHAR NOT NULL,

  PRIMARY KEY (lesson_id),
  CONSTRAINT fk_module
    FOREIGN KEY (module) REFERENCES module(module_id)
);

CREATE TABLE learner_lesson (
  learner VARCHAR,
  lesson uuid,
  scheduled_date DATE NOT NULL,
  completed bool,
  
  PRIMARY KEY (learner, lesson),
  
  CONSTRAINT fk_learner
    FOREIGN KEY (learner) REFERENCES learner(email),
  CONSTRAINT fk_lesson
    FOREIGN KEY (lesson) REFERENCES lesson(lesson_id)
);

-- tutorial table holds the tutorials for a module
CREATE TABLE tutorial (
  tutorial_id SERIAL,
  title VARCHAR NOT NULL,
  description TEXT NOT NULL,
  module VARCHAR NOT NULL,

  PRIMARY KEY (tutorial_id),
  CONSTRAINT fk_module
    FOREIGN KEY (module) REFERENCES module(module_id)
);

-- flashcards table holds the associated flashcards for a module
CREATE TABLE flashcard (
  flashcard_id uuid DEFAULT uuid_generate_v4 (),
  top_side TEXT NOT NULL,
  bottom_side TEXT NOT NULL,
  lesson uuid NOT NULL,

  PRIMARY KEY (flashcard_id),
  CONSTRAINT fk_lesson
    FOREIGN KEY (lesson) REFERENCES lesson(lesson_id)
);

-- Relationship between flashcard and learner
CREATE TABLE learner_flashcard (
  learner VARCHAR,
  flashcard uuid,
  repeat INT DEFAULT 0 NOT NULL,

  PRIMARY KEY (learner, flashcard),
  
  CONSTRAINT fk_learner
    FOREIGN KEY (learner) REFERENCES learner(email),
  CONSTRAINT fk_flashcard
    FOREIGN KEY (flashcard) REFERENCES flashcard(flashcard_id)
);
