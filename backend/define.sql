-- clear up old stuff
DROP TABLE IF EXISTS learner CASCADE;
DROP TABLE IF EXISTS module CASCADE;
DROP TABLE IF EXISTS lesson CASCADE;
DROP TABLE IF EXISTS tutorial CASCADE;
DROP TABLE IF EXISTS flashcard CASCADE;
DROP TABLE IF EXISTS learner_flashcard CASCADE;
DROP TABLE IF EXISTS learner_lesson CASCADE;

--UUID support
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- learner table hols the data about the users
CREATE TABLE learner (
  learner_id uuid DEFAULT uuid_generate_v4 (),
  username VARCHAR NOT NULL,
  password_hash VARCHAR NOT NULL,
  first_name VARCHAR NOT NULL,
  last_completed TIMESTAMPTZ DEFAULT NOW(),
  streak INT NOT NULL DEFAULT 0,

  PRIMARY KEY (learner_id)
);

-- module ID is codes like CS0001 etc, based on the university
CREATE TABLE module (
  module_id VARCHAR UNIQUE NOT NULL,
  title VARCHAR NOT NULL,
  description TEXT NOT NULL,

  PRIMARY KEY (module_id)
);

-- lesson table holds the lessons in a module
CREATE TABLE lesson (
  lesson_id uuid DEFAULT uuid_generate_v4 (),
  title VARCHAR NOT NULL,
  description TEXT NOT NULL,
  video_link VARCHAR NOT NULL,
  scheduled_date DATE NOT NULL,
  module VARCHAR NOT NULL,

  PRIMARY KEY (lesson_id),
  CONSTRAINT fk_module
    FOREIGN KEY (module) REFERENCES module(module_id)
);

-- tutorial table holds the tutorials for a module
CREATE TABLE tutorial (
  tutorial_id uuid DEFAULT uuid_generate_v4 (),
  title VARCHAR NOT NULL,
  description TEXT NOT NULL,
  scheduled_datetime TIMESTAMPTZ NOT NULL,
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
  learner uuid,
  flashcard uuid,
  repeat INT DEFAULT 0 NOT NULL,

  CONSTRAINT fk_learner
    FOREIGN KEY (learner) REFERENCES learner(learner_id),
  CONSTRAINT fk_flashcard
    FOREIGN KEY (flashcard) REFERENCES flashcard(flashcard_id)
);

CREATE TABLE learner_lesson (
  learner uuid,
  lesson uuid,
  completed bool,
  
  CONSTRAINT fk_learner
    FOREIGN KEY (learner) REFERENCES learner(learner_id),
  CONSTRAINT fk_lesson
    FOREIGN KEY (lesson) REFERENCES flashcard(lesson_id)
);
