-- ========================================================
-- Populate cdm schema
-- ========================================================

insert into omop.concept
(concept_id,concept_name,domain_id,vocabulary_id,concept_class_id,standard_concept,concept_code,valid_start_date,valid_end_date,invalid_reason)
values
    (2000006885,'Average height ','Measurement','Measurement','Measurement_contin','S','F','1970-01-01','2099-12-31',NULL),
    (2000000323,'MVP Age Group','Person','Person','Person_contin','S','F','1970-01-01','2099-12-31',NULL),
    (2000000324,'Sex, indicated by the subject','Person','Person','Observ_categ',NULL,'OMOP4822310','1970-01-01','2099-12-31',NULL),
    (2000000280,'BMI at enrollment','Measurement','Measurement','Measurement_contin','S','2','1970-01-01','2099-12-31',NULL)
;

-- These are the concepts we are looking for in the demo:
-- 2000006885 - Average height (VALUE_AS_NUMBER)
-- 2000000323 - MVP Age Group (VALUE_AS_STRING)
-- 2000000324 - Sex indicated by the subject (VALUE_AS_STRING)
-- 2000000280 - BMI at enrollment (VALUE_AS_NUMBER)

insert into omop.domain
(domain_id,domain_name)
values
	('Measurement', 'Measurement'),
	('Person', 'Person')
;

insert into omop.person
(person_id,gender_concept_id,year_of_birth,month_of_birth,day_of_birth,birth_datetime,death_datetime,race_concept_id,ethnicity_concept_id,location_id,provider_id,care_site_id,person_source_value,gender_source_value,gender_source_concept_id,race_source_value,race_source_concept_id,ethnicity_source_value,ethnicity_source_concept_id)
values
    (1,2000000324,1981,1,26,'1981-01-26 00:00:00',NULL,8527,0,NULL,NULL,NULL,'61735069-d238-1e52-1fac-bfc49c4b6325','F',0,'white',0,'nonhispanic',0),
    (2,2000000324,1971,12,6,'1971-12-06 00:00:00',NULL,8527,0,NULL,NULL,NULL,'8c66bd81-9588-69bd-6f39-5acc8242bfac','F',0,'white',0,'nonhispanic',0),
    (3,2000000324,1942,9,26,'1942-09-26 00:00:00',NULL,8527,0,NULL,NULL,NULL,'a0ebf5bf-0009-20af-bc00-04c256717664','F',0,'white',0,'nonhispanic',0),
    (4,2000000324,1993,5,22,'1993-05-22 00:00:00',NULL,8516,0,NULL,NULL,NULL,'d90c07cc-e303-298a-6d28-fbac7ff3f282','F',0,'black',0,'nonhispanic',0),
    (5,2000000324,1953,11,11,'1953-11-11 00:00:00',NULL,8515,0,NULL,NULL,NULL,'e6b6627f-4e38-dfc8-078c-11406151c521','F',0,'asian',0,'hispanic',0)
;

insert into omop.observation
(observation_id,person_id,observation_concept_id,observation_date,observation_datetime,observation_type_concept_id,value_as_number,value_as_string,value_as_concept_id,qualifier_concept_id,unit_concept_id,provider_id,visit_occurrence_id,visit_detail_id,observation_source_value,observation_source_concept_id,unit_source_value,qualifier_source_value,observation_event_id,obs_event_field_concept_id,value_as_datetime)
values
    (1,1,2000000324,'2019-03-29','2019-03-29 00:00:00',38000276,NULL,'F',0,0,0,NULL,26,0,'43878008',0,NULL,NULL,NULL,0,NULL),
	(22,1,2000000324,'2013-04-15','2013-04-15 00:00:00',38000276,NULL,'F',0,0,0,NULL,9,0,'302870006',0,NULL,NULL,NULL,0,NULL),
	(23,2,2000000324,'2014-02-05','2014-02-05 00:00:00',38000276,NULL,'A value with , comma!',0,0,0,NULL,52,0,'278860009',0,NULL,NULL,NULL,0,NULL),
	(35,2,0,'2017-06-13','2017-06-13 00:00:00',38000276,NULL,NULL,0,0,0,NULL,60,0,'444814009',0,NULL,NULL,NULL,0,NULL),
	(47,3,2000000324,'1993-10-24','1993-10-24 00:00:00',38000276,NULL,'M',0,0,0,NULL,81,0,'713197008',0,NULL,NULL,NULL,0,NULL),
	(48,3,0,'1967-12-02','1967-12-02 00:00:00',38000276,NULL,NULL,0,0,0,NULL,114,0,'53741008',0,NULL,NULL,NULL,0,NULL),
	(57,4,2000000324,'2019-02-16','2019-02-16 00:00:00',38000276,NULL,'F',0,0,0,NULL,162,0,'198992004',0,NULL,NULL,NULL,0,NULL),
	(58,4,0,'2012-06-06','2012-06-06 00:00:00',38000276,NULL,NULL,0,0,0,NULL,170,0,'403191005',0,NULL,NULL,NULL,0,NULL),
	(64,5,2000000324,'1993-11-17','1993-11-17 00:00:00',38000276,NULL,NULL,0,0,0,NULL,179,0,'162864005',0,NULL,NULL,NULL,0,NULL),
	(65,5,0,'2014-01-31','2014-01-31 00:00:00',38000276,NULL,NULL,0,0,0,NULL,197,0,'278860009',0,NULL,NULL,NULL,0,NULL),
	(66,5,0,'2020-03-16','2020-03-16 00:00:00',38000276,NULL,NULL,0,0,0,NULL,184,0,'84229001',0,NULL,NULL,NULL,0,NULL),
    -- 2000006885 mock "Average height "
    (1,1,2000006885,'2019-03-29','2019-03-29 00:00:00',38000276,5.4,NULL,0,0,0,NULL,26,0,'43878008',0,NULL,NULL,NULL,0,NULL),
	(22,1,2000006885,'2013-04-15','2013-04-15 00:00:00',38000276,5.5,NULL,0,0,0,NULL,9,0,'302870006',0,NULL,NULL,NULL,0,NULL),
	(23,2,2000006885,'2014-02-05','2014-02-05 00:00:00',38000276,6.2,NULL,0,0,0,NULL,52,0,'278860009',0,NULL,NULL,NULL,0,NULL),
	(35,2,0,'2017-06-13','2017-06-13 00:00:00',38000276,NULL,NULL,0,0,0,NULL,60,0,'444814009',0,NULL,NULL,NULL,0,NULL),
	(47,3,0,'1993-10-24','1993-10-24 00:00:00',38000276,NULL,'M',0,0,0,NULL,81,0,'713197008',0,NULL,NULL,NULL,0,NULL),
	(48,3,0,'1967-12-02','1967-12-02 00:00:00',38000276,NULL,NULL,0,0,0,NULL,114,0,'53741008',0,NULL,NULL,NULL,0,NULL),
	(58,4,0,'2012-06-06','2012-06-06 00:00:00',38000276,NULL,NULL,0,0,0,NULL,170,0,'403191005',0,NULL,NULL,NULL,0,NULL),
	(64,5,2000006885,'1993-11-17','1993-11-17 00:00:00',38000276,NULL,NULL,0,0,0,NULL,179,0,'162864005',0,NULL,NULL,NULL,0,NULL),
	(65,5,0,'2014-01-31','2014-01-31 00:00:00',38000276,NULL,NULL,0,0,0,NULL,197,0,'278860009',0,NULL,NULL,NULL,0,NULL)
;

-- ========================================================
-- Populate results schema
-- ========================================================

insert into results.COHORT
(cohort_definition_id,subject_id)
values
-- small cohort: 1 person:
    (1,1),
-- medium cohort: 2 persons:
    (2,2),
    (2,3),
-- large cohort: 5 persons:
    (3,1),
    (3,2),
    (3,3),
    (3,4),
    (3,5)
;
