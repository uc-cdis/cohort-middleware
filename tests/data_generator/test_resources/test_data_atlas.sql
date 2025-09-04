-- TODO - improve this by reusing the ../../setup_local_db/*.sql version which is basically a copy of this

-- ========================================================
-- Populate atlas2 schema
-- ========================================================

insert into atlas2.source
(source_id,source_name,source_connection,source_dialect,username,password)
values
    (1,'results_and_cdm_DATABASE', 'jdbc:postgresql://localhost:5434/postgres', 'postgres', 'postgres', 'mysecretpassword') -- pragma: allowlist secret
;

insert into atlas2.source_daimon
(source_daimon_id,source_id,daimon_type,table_qualifier,priority)
values
    (1,1,0, 'OMOP2', 1),
    (2,1,1, 'OMOP2', 1),
    (3,1,2, 'RESULTS2', 1),
    (4,1,5, 'TEMP', 1)
;
