CREATE TABLE people (
    id INT AUTO_INCREMENT,
    name VARCHAR(64) NOT NULL UNIQUE,
    descr TEXT,
    source VARCHAR(64),
    gender VARCHAR(8),
    ref VARCHAR(128),

    PRIMARY KEY (id)
);

CREATE TABLE people_names (
    id INT AUTO_INCREMENT,
    person_id INT NOT NULL,
    name VARCHAR(64) NOT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (person_id) REFERENCES people(id) ON DELETE CASCADE
);

CREATE TABLE people_tags (
    id INT AUTO_INCREMENT,
    person_id INT NOT NULL,
    tag VARCHAR(64) NOT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (person_id) REFERENCES people(id) ON DELETE CASCADE,
    INDEX(tag)
);