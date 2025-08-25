
-- TODO - improve this by reusing the ../../setup_local_db/*.sql version which is basically a copy of this

-- ========================================================
DROP SCHEMA IF EXISTS atlas2 CASCADE;
CREATE SCHEMA atlas2;
-- ========================================================

CREATE TABLE atlas2.source
(
    source_id integer NOT NULL,
    source_name character varying(100) NOT NULL,
    source_connection character varying(100) NOT NULL,
    source_dialect character varying(100),
    username character varying(100),
    password character varying(100),
    deleted_date timestamp(3),
    PRIMARY KEY (source_id)
);

CREATE TABLE atlas2.source_daimon
(
    source_daimon_id int NOT NULL,
    source_id int NOT NULL,
    daimon_type int NOT NULL DEFAULT 0,
    table_qualifier  VARCHAR (255) NOT NULL DEFAULT '?',
    priority int NOT NULL DEFAULT 0,
    CONSTRAINT PK_source_daimon PRIMARY KEY (source_daimon_id)
);

CREATE TABLE atlas2.cohort_definition
(
    id integer NOT NULL ,
    name varchar(255) NOT NULL,
    description varchar(1000) NULL,
    expression_type varchar(50) NULL,
    created_date    timestamp(3),
    modified_date   timestamp(3),
    created_by_id   integer,
    modified_by_id  integer,
    CONSTRAINT PK_cohort_definition PRIMARY KEY (id)
);
