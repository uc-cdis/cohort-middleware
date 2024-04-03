
-- ========================================================
DROP SCHEMA IF EXISTS results CASCADE;
CREATE SCHEMA results;
-- ========================================================

CREATE TABLE results.COHORT
(
    cohort_definition_id integer NOT NULL,
    subject_id integer NOT NULL,  -- this is person_id in cdm schema below!
    cohort_start_date date NOT NULL DEFAULT now(),
    cohort_end_date date NOT NULL DEFAULT DATE('2099-01-01')
);

-- This table can be present in future CDM schemas. Currently it is not filled by Atlas (per conversation with Andrew),
-- but might be used in the future, instead of atlas.cohort_definition table above.
-- CREATE TABLE results.COHORT_DEFINITION
-- (
--     cohort_definition_id integer NOT NULL,
--     cohort_definition_name varchar(255) NOT NULL,
--     cohort_definition_description TEXT NULL,
--     definition_type_concept_id integer NOT NULL DEFAULT 0,
--     cohort_definition_syntax TEXT NULL,
--     subject_concept_id integer NOT NULL DEFAULT 0,
--     cohort_initiation_date date NULL
-- );

-- ========================================================
DROP SCHEMA IF EXISTS omop CASCADE;
CREATE SCHEMA omop;
-- ========================================================

CREATE TABLE omop.person
(
    person_id integer NOT NULL,
    gender_concept_id integer NOT NULL DEFAULT 8507,
    year_of_birth integer NOT NULL DEFAULT 1970,
    month_of_birth integer,
    day_of_birth integer,
    birth_datetime timestamp without time zone,
    death_datetime timestamp without time zone,
    race_concept_id integer NOT NULL DEFAULT 8527,
    ethnicity_concept_id integer NOT NULL DEFAULT 0,
    location_id bigint,
    provider_id bigint,
    care_site_id bigint,
    person_source_value character varying(50) COLLATE pg_catalog."default",
    gender_source_value character varying(50) COLLATE pg_catalog."default",
    gender_source_concept_id integer NOT NULL DEFAULT 0,
    race_source_value character varying(50) COLLATE pg_catalog."default",
    race_source_concept_id integer NOT NULL DEFAULT 0,
    ethnicity_source_value character varying(50) COLLATE pg_catalog."default",
    ethnicity_source_concept_id integer NOT NULL DEFAULT 0
);

CREATE TABLE omop.observation
(
    observation_id bigint NOT NULL,
    person_id bigint NOT NULL,
    observation_concept_id integer NOT NULL DEFAULT 0,
    observation_date date DEFAULT now(),
    observation_datetime timestamp without time zone NOT NULL DEFAULT now(),
    observation_type_concept_id integer NOT NULL DEFAULT 38000276,
    value_as_number numeric,
    value_as_string character varying(60) COLLATE pg_catalog."default",
    value_as_concept_id integer,
    qualifier_concept_id integer,
    unit_concept_id integer,
    provider_id bigint,
    visit_occurrence_id bigint,
    visit_detail_id bigint,
    observation_source_value character varying(50) COLLATE pg_catalog."default",
    observation_source_concept_id integer NOT NULL  DEFAULT 0,
    unit_source_value character varying(50) COLLATE pg_catalog."default",
    qualifier_source_value character varying(50) COLLATE pg_catalog."default",
    observation_event_id bigint,
    obs_event_field_concept_id integer NOT NULL  DEFAULT 0,
    value_as_datetime timestamp without time zone
);
ALTER TABLE omop.observation  ADD CONSTRAINT xpk_observation PRIMARY KEY ( observation_id ) ;
drop sequence if exists observation_id_seq;
create sequence observation_id_seq start with 1;

CREATE TABLE omop.concept
(
    concept_id integer NOT NULL,
    concept_name character varying(255) COLLATE pg_catalog."default" NOT NULL,
    domain_id character varying(20) COLLATE pg_catalog."default" NOT NULL,
    vocabulary_id character varying(20) COLLATE pg_catalog."default" NOT NULL  DEFAULT 0,
    concept_class_id character varying(20) COLLATE pg_catalog."default" NOT NULL  DEFAULT 0,
    standard_concept character varying(1) COLLATE pg_catalog."default",
    concept_code character varying(50) COLLATE pg_catalog."default" NOT NULL,
    valid_start_date date NOT NULL  DEFAULT now(),
    valid_end_date date NOT NULL DEFAULT DATE('2099-01-01'),
    invalid_reason character varying(1) COLLATE pg_catalog."default"
);

CREATE VIEW omop.OBSERVATION_CONTINUOUS AS
SELECT ob.person_id, ob.observation_concept_id, ob.value_as_string, ob.value_as_number, ob.value_as_concept_id
FROM omop.observation ob
INNER JOIN omop.concept concept ON concept.CONCEPT_ID=ob.OBSERVATION_CONCEPT_ID
WHERE concept.CONCEPT_CLASS_ID='MVP Continuous' or concept.CONCEPT_ID=2000007027;

CREATE VIEW omop.DATA_DICTIONARY AS
SELECT c.vocabulary_id, c.concept_id, c.concept_code, c.concept_name, c.concept_class_id, COUNT(DISTINCT oc.person_id) as number_of_people_with_variable,
  (CASE WHEN c.concept_class_id = 'MVP Continuous' THEN (SELECT COUNT(DISTINCT person_id) FROM omop.OBSERVATION_CONTINUOUS oc1 WHERE oc1.observation_concept_id = c.concept_id  AND oc1.value_as_number IS NOT NULL) ELSE (SELECT COUNT(DISTINCT person_id) FROM omop.OBSERVATION_CONTINUOUS oc3 WHERE oc3.observation_concept_id = c.concept_id AND (oc3.value_as_concept_id IS NOT NULL and oc3.value_as_concept_id > 0)) END) as number_of_people_where_value_is_filled,
  (CASE WHEN c.concept_class_id  = 'MVP Continuous' THEN (SELECT COUNT(DISTINCT person_id) FROM omop.OBSERVATION_CONTINUOUS oc2 WHERE oc2.observation_concept_id = c.concept_id AND oc2.value_as_number IS NULL) ELSE (SELECT COUNT(DISTINCT person_id) FROM omop.OBSERVATION_CONTINUOUS oc4 WHERE oc4.observation_concept_id = c.concept_id AND (oc4.value_as_concept_id IS NULL or oc4.value_as_concept_id = 0)) END) as number_of_people_where_value_is_null,
  (CASE WHEN c.concept_class_id  = 'MVP Continuous' THEN 'Number' ELSE 'Concept Id' END) as value_stored_as,
  (CASE WHEN c.concept_class_id  = 'MVP Continuous' THEN MIN(value_as_number) ELSE NULL END) as min_value,
  (CASE WHEN c.concept_class_id  = 'MVP Continuous' THEN MAX(value_as_number) ELSE NULL END) as max_value,
  (CASE WHEN c.concept_class_id  = 'MVP Continuous' THEN AVG(value_as_number) ELSE NULL END) as mean_value,
  -- For sql server deployment, use stdev(value_as_number) instead of stddev(value_as_number)
  (CASE WHEN c.concept_class_id  = 'MVP Continuous' THEN stddev(value_as_number) ELSE NULL END) as standard_deviation,
  NULL as value_summary
  FROM omop.concept c
  INNER JOIN omop.OBSERVATION_CONTINUOUS oc on oc.observation_concept_id=c.concept_id
  GROUP BY c.vocabulary_id, c.concept_id, c.concept_code, c.concept_name, c.concept_class_id;
