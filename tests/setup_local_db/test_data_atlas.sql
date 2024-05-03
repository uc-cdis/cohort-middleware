-- ========================================================
-- Populate atlas schema
-- ========================================================

insert into atlas.source
(source_id,source_name,source_connection,source_dialect,username,password)
values
    (1,'results_and_cdm_DATABASE', 'jdbc:postgresql://localhost:5434;databaseName=postgres;user=postgres;password=mysecretpassword', 'postgres', 'postgres', 'mysecretpassword') -- pragma: allowlist secret
;

insert into atlas.source_daimon
(source_daimon_id,source_id,daimon_type,table_qualifier,priority)
values
    (1,1,0, 'OMOP', 1),
    (2,1,1, 'OMOP', 1),
    (3,1,2, 'RESULTS', 1),
    (4,1,5, 'TEMP', 1),
    (5,1,6, 'MISC', 1)
;

insert into atlas.cohort_definition
(id,name,description)
values
    (1,'Test cohort1','Small cohort'),
    (2,'Test cohort2','Medium cohort'),
    (3,'Test cohort3','Larger cohort'),
    (32,'Test cohort3b','Copy of Larger cohort'),
    (4,'Test cohort4','Extra Larger cohort')
;

insert into atlas.sec_role
    (id, name, system_role)
values
    (1,'public',true),
    (1005,'teamprojectX',false),
    (1009,'teamprojectY',false),
    (3000,'someotherrole',false),
    (4000,'defaultteamproject',false),
    (5000,'dummyGlobalReaderRole',false)
;

insert into atlas.sec_permission
    (id, value, description)
values
    (1181, 'cohortdefinition:2:check:post', 'Fix Cohort Definition with ID = 2'),
    (1182, 'cohortdefinition:2:put', 'Update Cohort Definition with ID = 2'),
    (1183, 'cohortdefinition:2:delete', 'Delete Cohort Definition with ID = 2'),
    (1184, 'cohortdefinition:2:version:*:get', 'Get cohort version'),
    (1185, 'cohortdefinition:2:info:get', 'no description'),
    (1186, 'cohortdefinition:2:get', 'Get Cohort Definition by ID'),
    (1187, 'cohortdefinition:2:version:get', 'Get list of cohort versions'),
    (1188, 'cohortdefinition:4:check:post', 'Fix Cohort Definition with ID = 4'),
    (1189, 'cohortdefinition:4:put', 'Update Cohort Definition with ID = 4'),
    (1190, 'cohortdefinition:4:delete', 'Delete Cohort Definition with ID = 4'),
    (1191, 'cohortdefinition:4:version:*:get', 'Get cohort version'),
    (1192, 'cohortdefinition:4:info:get', 'no description'),
    (1193, 'cohortdefinition:4:get', 'Get Cohort Definition by ID'),
    (1194, 'cohortdefinition:4:version:get', 'Get list of cohort versions'),
    (2193, 'cohortdefinition:1:get', 'Get Cohort Definition by ID'),
    (3193, 'cohortdefinition:3:get', 'Get Cohort Definition by ID'),
    (4193, 'cohortdefinition:32:get', 'Get Cohort Definition by ID')
;

insert into atlas.sec_role_permission
    (id, role_id, permission_id)
values
    (1454, 1005, 1181),
    (1455, 1005, 1182),
    (1456, 1005, 1183),
    (1457, 1005, 1184),
    (1458, 1005, 1185),
    (1459, 1005, 1186),
    (1460, 1005, 1187),
    (1461, 1005, 4193), -- 1005 teamprojectX has access to cohorts 2 and 32
    (2461, 1009, 1188),
    (2462, 1009, 1189),
    (2463, 1009, 1190),
    (2464, 1009, 1191),
    (2465, 1009, 1192),
    (2466, 1009, 1193),
    (2467, 1009, 1194), -- 1009 teamprojectY has access to cohort 4
    (4454, 4000, 1181),
    (4455, 4000, 1182),
    (4456, 4000, 1183),
    (4457, 4000, 1184),
    (4458, 4000, 1185),
    (4459, 4000, 1186),
    (4460, 4000, 1187),
    (4461, 4000, 1188),
    (4462, 4000, 1189),
    (4463, 4000, 1190),
    (4464, 4000, 1191),
    (4465, 4000, 1192),
    (4466, 4000, 1193),
    (4467, 4000, 1194),
    (4468, 4000, 2193),
    (4469, 4000, 3193),
    (4470, 4000, 4193), -- 4000 a "default" teamproject that has access to all cohorts - not really used in practice...but a possible kind of scenario.
    (5464, 5000, 1191),
    (5465, 5000, 1192),
    (5466, 5000, 1193),
    (5467, 5000, 1194) -- 5000 dummyGlobalReaderRole has READ ONLY access to cohort 4
;
