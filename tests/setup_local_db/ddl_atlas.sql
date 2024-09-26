
-- ========================================================
DROP SCHEMA IF EXISTS atlas CASCADE;
CREATE SCHEMA atlas;
-- ========================================================

CREATE TABLE atlas.source
(
    source_id integer NOT NULL,
    source_name character varying(100) NOT NULL,
    source_connection character varying(100) NOT NULL,
    source_dialect character varying(100),
    username character varying(100),
    password character varying(100),
    PRIMARY KEY (source_id)
);

CREATE TABLE atlas.source_daimon
(
    source_daimon_id int NOT NULL,
    source_id int NOT NULL,
    daimon_type int NOT NULL DEFAULT 0,
    table_qualifier  VARCHAR (255) NOT NULL DEFAULT '?',
    priority int NOT NULL DEFAULT 0,
    CONSTRAINT PK_source_daimon PRIMARY KEY (source_daimon_id)
);

CREATE TABLE atlas.cohort_definition
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

CREATE TABLE atlas.cohort_definition_details
(
    id integer NOT NULL ,
    expression text NOT NULL,
    hash_code integer,
    CONSTRAINT PK_cohort_definition_details PRIMARY KEY (id),
    CONSTRAINT fk_cohort_definition_details_to_cohort_definition FOREIGN KEY (id)
        REFERENCES atlas.cohort_definition (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
);

CREATE TABLE atlas.sec_role
(
    id integer NOT NULL,
    name varchar(255) ,
    system_role boolean NOT NULL DEFAULT false,
    CONSTRAINT pk_sec_role PRIMARY KEY (id),
    CONSTRAINT sec_role_name_uq UNIQUE (name, system_role)
);

CREATE TABLE atlas.sec_permission
(
    id integer NOT NULL,
    value varchar(255) NOT NULL,
    description varchar(255),
    CONSTRAINT pk_sec_permission PRIMARY KEY (id),
    CONSTRAINT permission_unique UNIQUE (value)
);

CREATE TABLE atlas.sec_role_permission
(
    id integer NOT NULL,
    role_id integer NOT NULL,
    permission_id integer NOT NULL,
    status varchar(255),
    CONSTRAINT pk_sec_role_permission PRIMARY KEY (id),
    CONSTRAINT role_permission_unique UNIQUE (role_id, permission_id),
    CONSTRAINT fk_role_permission_to_permission FOREIGN KEY (permission_id)
        REFERENCES atlas.sec_permission (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT fk_role_permission_to_role FOREIGN KEY (role_id)
        REFERENCES atlas.sec_role (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

CREATE TABLE atlas.schema_version
(
    installed_rank integer NOT NULL,
    version varchar(50),
    description varchar(200) NOT NULL,
    type varchar(20) NOT NULL,
    script varchar(1000) NOT NULL,
    checksum int,
    installed_by varchar(100) NOT NULL,
    installed_on timestamp(3) NOT NULL,
    execution_time int NOT NULL,
    success boolean NOT NULL,
    CONSTRAINT pk_schema_version PRIMARY KEY (installed_rank)
);


CREATE VIEW atlas.COHORT_DEFINITION_SEC_ROLE AS
  select
    distinct cast(regexp_replace(sec_permission.value,
         '^cohortdefinition:([0-9]+):.*','\1') as integer) as cohort_definition_id,
    sec_role.name as sec_role_name
  from
    atlas.sec_role
    inner join atlas.sec_role_permission on sec_role.id = sec_role_permission.role_id
    inner join atlas.sec_permission on sec_role_permission.permission_id = sec_permission.id
  where
    sec_permission.value ~ 'cohortdefinition:[0-9]+'
;
