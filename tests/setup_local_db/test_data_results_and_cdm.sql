-- ========================================================
-- Populate cdm schema
-- ========================================================

insert into omop.concept
(concept_id,concept_name,domain_id,vocabulary_id,concept_class_id,standard_concept,concept_code,valid_start_date,valid_end_date,invalid_reason)
values
    (2000006885,'Average height ','Measurement','Measurement','MVP Continuous','S','F','1970-01-01','2099-12-31',NULL),
    (2000000323,'MVP Age Group','Person','Person','MVP Continuous','S','F','1970-01-01','2099-12-31',NULL),
    (2000000324,'Some continuous value, indicated by the subject','Person','Person','MVP Continuous',NULL,'OMOP4822310','1970-01-01','2099-12-31',NULL),
    (2000000280,'BMI at enrollment','Measurement','Measurement','MVP Continuous','S','2','1970-01-01','2099-12-31',NULL)
;

-- TODO - add concept_type_class "concept_class_id" "MVP Continuous" to better reflect real queries?

-- These are the concepts we are looking for in the demo:
-- 2000006885 - Average height (VALUE_AS_NUMBER)
-- 2000000323 - MVP Age Group (VALUE_AS_STRING)
-- 2000000324 - Sex indicated by the subject (VALUE_AS_STRING)
-- 2000000280 - BMI at enrollment (VALUE_AS_NUMBER)

insert into omop.person
(person_id,gender_concept_id,year_of_birth,month_of_birth,day_of_birth,birth_datetime,death_datetime,race_concept_id,ethnicity_concept_id,location_id,provider_id,care_site_id,person_source_value,gender_source_value,gender_source_concept_id,race_source_value,race_source_concept_id,ethnicity_source_value,ethnicity_source_concept_id)
values
    (1,2000000324,1981,1,26,'1981-01-26 00:00:00',NULL,8527,0,NULL,NULL,NULL,'61735069-d238-1e52-1fac-bfc49c4b6325','F',0,'white',0,'',0),
    (2,2000000324,1971,12,6,'1971-12-06 00:00:00',NULL,8527,0,NULL,NULL,NULL,'8c66bd81-9588-69bd-6f39-5acc8242bfac','F',0,'white',0,'',0),
    (3,2000000324,1942,9,26,'1942-09-26 00:00:00',NULL,8527,0,NULL,NULL,NULL,'a0ebf5bf-0009-20af-bc00-04c256717664','F',0,'white',0,'',0),
    (4,2000000324,1993,5,22,'1993-05-22 00:00:00',NULL,8516,0,NULL,NULL,NULL,'d90c07cc-e303-298a-6d28-fbac7ff3f282','F',0,'black',0,'',0),
    (5,2000000324,1953,11,11,'1953-11-11 00:00:00',NULL,8515,0,NULL,NULL,NULL,'e6b6627f-4e38-dfc8-078c-11406151c521','F',0,'asian',0,'',0),
    (6,2000000324,1958,11,11,'1958-11-11 00:00:00',NULL,8515,0,NULL,NULL,NULL,'e6b6627f-4e38-dfc8-078c-11406151c522','F',0,'asian',0,'',0),
    -- additional people for historgram data testing
    (7,2000000324,1961,11,11,'1961-11-11 00:00:00',NULL,8515,0,NULL,NULL,NULL,'e6b6627f-4e38-dfc8-078c-11406151c523','F',0,'asian',0,'',0),
    (8,2000000324,1962,11,11,'1962-11-11 00:00:00',NULL,8515,0,NULL,NULL,NULL,'e6b6627f-4e38-dfc8-078c-11406151c524','F',0,'asian',0,'',0),
    (9,2000000324,1963,11,11,'1963-11-11 00:00:00',NULL,8515,0,NULL,NULL,NULL,'e6b6627f-4e38-dfc8-078c-11406151c525','F',0,'asian',0,'',0),
    (10,2000000324,1964,11,11,'1964-11-11 00:00:00',NULL,8515,0,NULL,NULL,NULL,'e6b6627f-4e38-dfc8-078c-11406151c536','F',0,'asian',0,'',0),
    (11,2000000324,1965,11,11,'1965-11-11 00:00:00',NULL,8515,0,NULL,NULL,NULL,'e6b6627f-4e38-dfc8-078c-11406151c528','F',0,'asian',0,'',0),
    (12,2000000324,1966,11,11,'1966-11-11 00:00:00',NULL,8515,0,NULL,NULL,NULL,'e6b6627f-4e38-dfc8-078c-11406151c529','F',0,'asian',0,'',0),
    (13,2000000324,1967,11,11,'1967-11-11 00:00:00',NULL,8515,0,NULL,NULL,NULL,'e6b6627f-4e38-dfc8-078c-11406151c599','F',0,'asian',0,'',0),
    (14,2000000324,1968,11,11,'1968-11-11 00:00:00',NULL,8515,0,NULL,NULL,NULL,'e6b6627f-4e38-dfc8-078c-11406151c598','F',0,'asian',0,'',0),
    (15,2000000324,1969,11,11,'1969-11-11 00:00:00',NULL,8515,0,NULL,NULL,NULL,'e6b6627f-4e38-dfc8-078c-11406151c597','F',0,'asian',0,'',0),
    (16,2000000324,1970,11,11,'1970-11-11 00:00:00',NULL,8515,0,NULL,NULL,NULL,'e6b6627f-4e38-dfc8-078c-11406151c596','F',0,'asian',0,'',0),
    (17,2000000324,1971,11,11,'1971-11-11 00:00:00',NULL,8515,0,NULL,NULL,NULL,'e6b6627f-4e38-dfc8-078c-11406151c595','F',0,'asian',0,'',0),
    (18,2000000324,1972,11,11,'1972-11-11 00:00:00',NULL,8515,0,NULL,NULL,NULL,'e6b6627f-4e38-dfc8-078c-11406151c594','F',0,'asian',0,'',0)
;

-- add a mix of:
--  - good observation records with a real `observation_concept_id` and a real value in `value_as_string` or `value_as_number`
--  - bad observation records, where `observation_concept_id` is missing or the `value_as_string` or `value_as_number` are both NULL:
insert into omop.observation
(observation_id,person_id,observation_concept_id,observation_date,observation_datetime,observation_type_concept_id,value_as_number,value_as_string,value_as_concept_id,qualifier_concept_id,unit_concept_id,provider_id,visit_occurrence_id,visit_detail_id,observation_source_value,observation_source_concept_id,unit_source_value,qualifier_source_value,observation_event_id,obs_event_field_concept_id,value_as_datetime)
values
    (nextval('observation_id_seq'), 1,2000000324,'2019-03-29','2019-03-29 00:00:00',38000276,NULL,'F',0,0,0,NULL,26,0,'43878008',0,NULL,NULL,NULL,0,NULL),
	(nextval('observation_id_seq'), 1,2000000324,'2013-04-15','2013-04-15 00:00:00',38000276,NULL,'F',0,0,0,NULL,9,0,'302870006',0,NULL,NULL,NULL,0,NULL),
	(nextval('observation_id_seq'), 2,2000000324,'2014-02-05','2014-02-05 00:00:00',38000276,NULL,'A value with , comma!',0,0,0,NULL,52,0,'278860009',0,NULL,NULL,NULL,0,NULL),
	(nextval('observation_id_seq'), 2,0,'2017-06-13','2017-06-13 00:00:00',38000276,NULL,NULL,0,0,0,NULL,60,0,'444814009',0,NULL,NULL,NULL,0,NULL),
	(nextval('observation_id_seq'), 3,2000000324,'1993-10-24','1993-10-24 00:00:00',38000276,NULL,'M',0,0,0,NULL,81,0,'713197008',0,NULL,NULL,NULL,0,NULL),
	(nextval('observation_id_seq'), 3,0,'1967-12-02','1967-12-02 00:00:00',38000276,NULL,NULL,0,0,0,NULL,114,0,'53741008',0,NULL,NULL,NULL,0,NULL),
	(nextval('observation_id_seq'), 4,2000000324,'2019-02-16','2019-02-16 00:00:00',38000276,NULL,'F',0,0,0,NULL,162,0,'198992004',0,NULL,NULL,NULL,0,NULL),
	(nextval('observation_id_seq'), 4,0,'2012-06-06','2012-06-06 00:00:00',38000276,NULL,NULL,0,0,0,NULL,170,0,'403191005',0,NULL,NULL,NULL,0,NULL),
	(nextval('observation_id_seq'), 5,2000000324,'1993-11-17','1993-11-17 00:00:00',38000276,NULL,NULL,0,0,0,NULL,179,0,'162864005',0,NULL,NULL,NULL,0,NULL),
	(nextval('observation_id_seq'), 5,0,'2014-01-31','2014-01-31 00:00:00',38000276,NULL,NULL,0,0,0,NULL,197,0,'278860009',0,NULL,NULL,NULL,0,NULL),
	(nextval('observation_id_seq'), 5,0,'2020-03-16','2020-03-16 00:00:00',38000276,NULL,NULL,0,0,0,NULL,184,0,'84229001',0,NULL,NULL,NULL,0,NULL),
    -- 2000006885 mock "Average height "
    (nextval('observation_id_seq'), 1,2000006885,'2019-03-29','2019-03-29 00:00:00',38000276,5.4,NULL,0,0,0,NULL,26,0,'43878008',0,NULL,NULL,NULL,0,NULL),
	(nextval('observation_id_seq'), 1,2000006885,'2013-04-15','2013-04-15 00:00:00',38000276,5.5,NULL,0,0,0,NULL,9,0,'302870006',0,NULL,NULL,NULL,0,NULL),
	(nextval('observation_id_seq'), 2,2000006885,'2014-02-05','2014-02-05 00:00:00',38000276,6.2,NULL,0,0,0,NULL,52,0,'278860009',0,NULL,NULL,NULL,0,NULL),
	(nextval('observation_id_seq'), 2,0,'2017-06-13','2017-06-13 00:00:00',38000276,NULL,NULL,0,0,0,NULL,60,0,'444814009',0,NULL,NULL,NULL,0,NULL),
	(nextval('observation_id_seq'), 3,0,'1993-10-24','1993-10-24 00:00:00',38000276,NULL,'M',0,0,0,NULL,81,0,'713197008',0,NULL,NULL,NULL,0,NULL),
	(nextval('observation_id_seq'), 3,0,'1967-12-02','1967-12-02 00:00:00',38000276,NULL,NULL,0,0,0,NULL,114,0,'53741008',0,NULL,NULL,NULL,0,NULL),
	(nextval('observation_id_seq'), 4,0,'2012-06-06','2012-06-06 00:00:00',38000276,NULL,NULL,0,0,0,NULL,170,0,'403191005',0,NULL,NULL,NULL,0,NULL),
    -- person 5 has null in value_as_number here:
	(nextval('observation_id_seq'), 5,2000006885,'1993-11-17','1993-11-17 00:00:00',38000276,NULL,NULL,0,0,0,NULL,179,0,'162864005',0,NULL,NULL,NULL,0,NULL),
	(nextval('observation_id_seq'), 5,0,'2014-01-31','2014-01-31 00:00:00',38000276,NULL,NULL,0,0,0,NULL,197,0,'278860009',0,NULL,NULL,NULL,0,NULL),
   	(nextval('observation_id_seq'), 6,2000006885,'2014-01-31','2014-01-31 00:00:00',38000276,5.41,NULL,0,0,0,NULL,197,0,'278860009',0,NULL,NULL,NULL,0,NULL),
    -- additional Average height to test histogram
    (nextval('observation_id_seq'), 7,2000006885,'2014-01-31','2014-01-31 00:00:00',38000276,7.23,NULL,0,0,0,NULL,197,0,'278860009',0,NULL,NULL,NULL,0,NULL),
    (nextval('observation_id_seq'), 8,2000006885,'2014-01-31','2014-01-31 00:00:00',38000276,7.44,NULL,0,0,0,NULL,197,0,'278860009',0,NULL,NULL,NULL,0,NULL),
    (nextval('observation_id_seq'), 9,2000006885,'2014-01-31','2014-01-31 00:00:00',38000276,8.21,NULL,0,0,0,NULL,197,0,'278860009',0,NULL,NULL,NULL,0,NULL),
    (nextval('observation_id_seq'), 10,2000006885,'2014-01-31','2014-01-31 00:00:00',38000276,6.14,NULL,0,0,0,NULL,197,0,'278860009',0,NULL,NULL,NULL,0,NULL),
    (nextval('observation_id_seq'), 11,2000006885,'2014-01-31','2014-01-31 00:00:00',38000276,3.14,NULL,0,0,0,NULL,197,0,'278860009',0,NULL,NULL,NULL,0,NULL),
    (nextval('observation_id_seq'), 12,2000006885,'2014-01-31','2014-01-31 00:00:00',38000276,1.16,NULL,0,0,0,NULL,197,0,'278860009',0,NULL,NULL,NULL,0,NULL),
    (nextval('observation_id_seq'), 13,2000006885,'2014-01-31','2014-01-31 00:00:00',38000276,9.52,NULL,0,0,0,NULL,197,0,'278860009',0,NULL,NULL,NULL,0,NULL),
    (nextval('observation_id_seq'), 14,2000006885,'2014-01-31','2014-01-31 00:00:00',38000276,7.83,NULL,0,0,0,NULL,197,0,'278860009',0,NULL,NULL,NULL,0,NULL),
    (nextval('observation_id_seq'), 15,2000006885,'2014-01-31','2014-01-31 00:00:00',38000276,4.52,NULL,0,0,0,NULL,197,0,'278860009',0,NULL,NULL,NULL,0,NULL),
    (nextval('observation_id_seq'), 16,2000006885,'2014-01-31','2014-01-31 00:00:00',38000276,8.73,NULL,0,0,0,NULL,197,0,'278860009',0,NULL,NULL,NULL,0,NULL),
    (nextval('observation_id_seq'), 17,2000006885,'2014-01-31','2014-01-31 00:00:00',38000276,2.33,NULL,0,0,0,NULL,197,0,'278860009',0,NULL,NULL,NULL,0,NULL),
    (nextval('observation_id_seq'), 18,2000006885,'2014-01-31','2014-01-31 00:00:00',38000276,6.32,NULL,0,0,0,NULL,197,0,'278860009',0,NULL,NULL,NULL,0,NULL)

;

/**
HARE DUMMY DATA
For HARE info, see https://pubmed.ncbi.nlm.nih.gov/31564439/.
    HIS - Hispanic
    ASN - non-Hispanic Asian
    EUR - non-Hispanic White
    AFR - non-Hispanic Black
    NA - Missing
**/
insert into omop.concept
(concept_id,concept_code,concept_name,domain_id,vocabulary_id,concept_class_id,standard_concept,valid_start_date,valid_end_date,invalid_reason)
values
    (2000007027,'HARE_CODE','HARE',        'Person','Person','MVP Nominal','S','1970-01-01','2099-12-31',NULL),
    (2000007028,'HIS', 'Hispanic',         'Person','Person','MVP Nominal','S','1970-01-01','2099-12-31',NULL),
    (2000007029,'ASN','non-Hispanic Asian','Person','Person','MVP Nominal','S','1970-01-01','2099-12-31',NULL),
    (2000007030,'AFR','non-Hispanic Black','Person','Person','MVP Nominal','S','1970-01-01','2099-12-31',NULL),
    (2000007031,'EUR','non-Hispanic White','Person','Person','MVP Nominal','S','1970-01-01','2099-12-31',NULL),
    (2000007032,'OTH','Other',             'Person','Person','MVP Nominal','S','1970-01-01','2099-12-31',NULL)
;

-- insert `observation` records:
insert into omop.observation
(observation_id,                        person_id, observation_concept_id, value_as_concept_id, value_as_string, observation_source_value, observation_type_concept_id)
values
    (nextval('observation_id_seq'),          1,           2000007027,           2000007028,            'HIS',           'HIS',                    38000276),
    (nextval('observation_id_seq'),          2,           2000007027,           2000007029,            'ASN',           'ASN',                    38000276),
    (nextval('observation_id_seq'),          3,           2000007027,           2000007031,            'EUR',           'EUR',                    38000276),
    (nextval('observation_id_seq'),          4,           2000007027,           2000007030,            'AFR',           'AFR',                    38000276),
    (nextval('observation_id_seq'),          5,           2000007027,           NULL,             NULL,            NULL,                    38000276),
    (nextval('observation_id_seq'),          6,           2000007027,           2000007029,            'ASN',           'ASN',                    38000276),
    (nextval('observation_id_seq'),          2,           2000007027,           2000007030,            'AFR',           'AFR',                    38000276),
    (nextval('observation_id_seq'),          7,           2000007027,           2000007028,            'HIS',           'HIS',                    38000276),
    (nextval('observation_id_seq'),          8,           2000007027,           2000007029,            'ASN',           'ASN',                    38000276),
    (nextval('observation_id_seq'),          9,           2000007027,           2000007031,            'EUR',           'EUR',                    38000276),
    (nextval('observation_id_seq'),         10,           2000007027,           2000007030,            'AFR',           'AFR',                    38000276),
    (nextval('observation_id_seq'),         11,           2000007027,           2000007030,            'AFR',           'AFR',                    38000276)
;

-- ========================================================
-- Populate results schema
-- ========================================================

insert into results.COHORT
(cohort_definition_id,subject_id)
values
-- small cohort: 1 person: (aka "smallestCohort" in models_test.go script)
    (1,1),
-- medium cohort: 2 persons: (aka "thirdLargestCohort" in models_test.go script)
    (2,2),
    (2,3),
-- large cohort: 6 persons: (aka "secondLargestCohort" in models_test.go script)
    (3,1),
    (3,2),
    (3,3),
    (3,4),
    (3,5),
    (3,6),
-- copy of large cohort above plus some extra persons: (aka "extendedCopyOfSecondLargestCohort" in models_test.go script)
    (32,1),
    (32,2),
    (32,3),
    (32,4),
    (32,5),
    (32,6),
    (32,7),
    (32,8),
    (32,9),
    (32,10),
-- extra large cohort for testing histogram: (aka "largestCohort" in models_test.go script)
    (4,5),
    (4,6),
    (4,7),
    (4,8),
    (4,9),
    (4,10),
    (4,11),
    (4,12),
    (4,13),
    (4,14),
    (4,15),
    (4,16),
    (4,17),
    (4,18)
;

WITH cte_counts AS (SELECT observation_concept_id,
                           COUNT(DISTINCT person_id) AS number_of_people_with_variable,
                           COUNT(DISTINCT CASE WHEN value_as_number IS NOT NULL THEN person_id END) AS number_of_people_where_value_is_filled_number,
                           COUNT(DISTINCT CASE WHEN value_as_concept_id IS NOT NULL AND value_as_concept_id > 0 THEN person_id END) AS number_of_people_where_value_is_filled_concept,
                           COUNT(DISTINCT CASE WHEN value_as_number IS NULL THEN person_id END) AS number_of_people_where_value_is_null_number,
                           COUNT(DISTINCT CASE WHEN value_as_concept_id IS NULL OR value_as_concept_id = 0 THEN person_id END) AS number_of_people_where_value_is_null_concept
                    FROM omop.OBSERVATION
                    GROUP BY observation_concept_id),
cte_observation_class AS (select c.concept_id FROM OMOP.concept c, omop.observation ob
Where c.concept_class_id in ('MVP Year', 'MVP_v20_1')
And ob.OBSERVATION_CONCEPT_ID = c.CONCEPT_ID
and ob.VALUE_AS_CONCEPT_ID in (4306768, 2000007053, 2000007054, 2000007055, 2000007056, 2000007059))
SELECT c.vocabulary_id,
       c.concept_id,
       c.concept_code,
       c.concept_name,
       c.concept_class_id,
       cc.number_of_people_with_variable,
       CASE
           WHEN c.concept_class_id IN ('MVP Continuous', 'MVP DF', 'MVP Discrete', 'MVP Year') THEN cc.number_of_people_where_value_is_filled_number
           ELSE cc.number_of_people_where_value_is_filled_concept END AS number_of_people_where_value_is_filled,
       CASE
           WHEN  c.concept_class_id IN ('MVP Continuous', 'MVP DF', 'MVP Discrete', 'MVP Year') THEN cc.number_of_people_where_value_is_null_number
           ELSE cc.number_of_people_where_value_is_null_concept END  AS number_of_people_where_value_is_null,
       CASE WHEN c.concept_class_id IN ('MVP Continuous', 'MVP DF', 'MVP Discrete', 'MVP Year') OR c.CONCEPT_ID in (select concept_id from cte_observation_class coc) THEN 'Number' ELSE 'Concept Id' END AS value_stored_as,
       MIN(oc.value_as_number) AS min_value,
       MAX(oc.value_as_number) AS max_value,
       AVG(oc.value_as_number) AS mean_value,
-- For sql server deployment, use stdev(value_as_number) instead of stddev(value_as_number)
       STDDEV(oc.value_as_number) AS standard_deviation,
       NULL AS value_summary
INTO misc.DATA_DICTIONARY
FROM omop.CONCEPT c
         JOIN omop.OBSERVATION oc ON oc.observation_concept_id = c.concept_id
         JOIN cte_counts cc ON cc.observation_concept_id = c.concept_id
         LEFT JOIN cte_observation_class coc on coc.concept_id = c.concept_id
WHERE c.CONCEPT_CLASS_ID IN ('MVP Continuous', 'MVP DF', 'MVP Discrete', 'MVP Year', 'MVP Dichotomous', 'MVP Nominal', 'MVP Answer', 'MVP_v20_1')
GROUP BY c.vocabulary_id, c.concept_id, c.concept_code, c.concept_name, c.concept_class_id,
         cc.number_of_people_with_variable,
         cc.number_of_people_where_value_is_filled_number,
         cc.number_of_people_where_value_is_filled_concept,
         cc.number_of_people_where_value_is_null_number,
         cc.number_of_people_where_value_is_null_concept;
