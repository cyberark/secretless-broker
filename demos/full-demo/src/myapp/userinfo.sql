DROP TABLE IF EXISTS userinfo;

CREATE TABLE userinfo
(
    uid serial NOT NULL,
    username text NOT NULL,
    department text NOT NULL,
    created timestamp default current_timestamp,
    CONSTRAINT userinfo_pkey PRIMARY KEY (uid)
)
;
