DROP TABLE IF EXISTS projects;
CREATE TABLE projects (
	id INT(11) NOT NULL AUTO_INCREMENT,
	name VARCHAR(255) NOT NULL DEFAULT '',
	PRIMARY KEY (id)
) ENGINE=MyISAM DEFAULT CHARSET=UTF8;

DROP TABLE IF EXISTS types;
CREATE TABLE types (
	id INT(11) NOT NULL AUTO_INCREMENT,
	project_id INT(11) NOT NULL DEFAULT 0,
	name VARCHAR(255) NOT NULL DEFAULT '',
	
	PRIMARY KEY (id),
	KEY (project_id)
) ENGINE=MyISAM DEFAULT CHARSET=UTF8;

DROP TABLE IF EXISTS stats;
CREATE TABLE stats (
	id INT(11) NOT NULL AUTO_INCREMENT,
	project_id INT(11) NOT NULL DEFAULT 0,
	type_id INT(11) NOT NULL DEFAULT 0,

	value DECIMAL(10,5) NOT NULL DEFAULT 0,

	meta_type_id INT(11) NOT NULL DEFAULT 0,
	meta_value VARCHAR(255) NOT NULL DEFAULT '',

	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

	PRIMARY KEY (id),
	
	KEY t (type_id),
	KEY p (project_id),
	KEY mt (meta_type_id),
	KEY c (created_at),
	
	KEY p_t (project_id, type_id),
	KEY p_mt (project_id, meta_type_id),
	KEY p_t_mt (project_id, type_id, meta_type_id),
	
	KEY p_t_c (project_id, type_id, created_at),
	KEY p_mt_c (project_id, meta_type_id, created_at),
	KEY p_t_mt_c (project_id, type_id, meta_type_id, created_at)
	
) ENGINE=MyISAM DEFAULT CHARSET=UTF8;

