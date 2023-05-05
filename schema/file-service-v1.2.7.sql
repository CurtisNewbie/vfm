-- schema copied from github.com/curtisnewbie/file-server v1.2.7, not all tables included because some of them are not supported in vfm
CREATE TABLE IF NOT EXISTS file_info (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL COMMENT 'name of the file',
  `uuid` varchar(64) NOT NULL COMMENT 'file''s uuid',
  `is_logic_deleted` int NOT NULL DEFAULT '0' COMMENT 'whether the file is logically deleted, 0-normal, 1-deleted',
  `is_physic_deleted` int NOT NULL DEFAULT '0' COMMENT 'whether the file is physically deleted, 0-normal, 1-deleted',
  `size_in_bytes` bigint NOT NULL COMMENT 'size of file in bytes',
  `uploader_id` int NOT NULL DEFAULT '0' COMMENT 'uploader id, i.e., user.id',
  `uploader_name` varchar(255) NOT NULL DEFAULT '' COMMENT 'uploader name',
  `upload_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'upload time',
  `logic_delete_time` datetime DEFAULT NULL COMMENT 'when the file is logically deleted',
  `physic_delete_time` datetime DEFAULT NULL COMMENT 'when the file is physically deleted',
  `user_group` int NOT NULL COMMENT 'the group that the file belongs to, 0-public, 1-private',
  `fs_group_id` int NOT NULL DEFAULT '0' COMMENT 'id of fs_group',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'when the record is created',
  `create_by` varchar(255) NOT NULL DEFAULT '' COMMENT 'who created this record',
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'when the record is updated',
  `update_by` varchar(255) NOT NULL DEFAULT '' COMMENT 'who updated this record',
  `is_del` tinyint NOT NULL DEFAULT '0' COMMENT '0-normal, 1-deleted',
  `file_type` varchar(6) NOT NULL DEFAULT 'FILE' COMMENT 'file type: FILE, DIR',
  `parent_file` varchar(64) NOT NULL DEFAULT '' COMMENT 'parent file uuid',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uuid_uk` (`uuid`),
  KEY `parent_file_type_idx` (`parent_file`,`file_type`),
  KEY `uploader_id_idx` (`uploader_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS file_tag (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  `file_id` int unsigned NOT NULL COMMENT 'id of file_info',
  `tag_id` int unsigned NOT NULL COMMENT 'id of tag',
  `user_id` int unsigned NOT NULL COMMENT 'id of user who created this file_tag relation',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'when the record is created',
  `create_by` varchar(255) NOT NULL DEFAULT '' COMMENT 'who created this record',
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'when the record is updated',
  `update_by` varchar(255) NOT NULL DEFAULT '' COMMENT 'who updated this record',
  `is_del` tinyint NOT NULL DEFAULT '0' COMMENT '0-normal, 1-deleted',
  PRIMARY KEY (`id`),
  KEY `user_id_file_id_idx` (`user_id`,`file_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='join table between file_info and tag'

CREATE TABLE IF NOT EXISTS tag (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  `name` varchar(50) NOT NULL COMMENT 'name of tag',
  `user_id` int unsigned NOT NULL COMMENT 'user who owns this tag (tags are isolated between different users)',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'when the record is created',
  `create_by` varchar(255) NOT NULL DEFAULT '' COMMENT 'who created this record',
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'when the record is updated',
  `update_by` varchar(255) NOT NULL DEFAULT '' COMMENT 'who updated this record',
  `is_del` tinyint NOT NULL DEFAULT '0' COMMENT '0-normal, 1-deleted',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_tag` (`user_id`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='tag';

CREATE TABLE IF NOT EXISTS file_sharing (
  `id` int NOT NULL AUTO_INCREMENT,
  `file_id` int NOT NULL COMMENT 'id of file_info',
  `user_id` int NOT NULL COMMENT 'user who now have access to the file',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'when the record is created',
  `create_by` varchar(255) NOT NULL DEFAULT '' COMMENT 'who created this record',
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'when the record is updated',
  `update_by` varchar(255) NOT NULL DEFAULT '' COMMENT 'who updated this record',
  `is_del` tinyint NOT NULL DEFAULT '0' COMMENT 'is deleted, 0: normal, 1: deleted',
  PRIMARY KEY (`id`),
  UNIQUE KEY `file_id` (`file_id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='file''s sharing information';

CREATE TABLE IF NOT EXISTS vfolder (
  `id` int NOT NULL AUTO_INCREMENT,
  `folder_no` varchar(64) NOT NULL COMMENT 'folder no',
  `name` varchar(255) NOT NULL COMMENT 'name of the folder',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'when the record is created',
  `create_by` varchar(255) NOT NULL DEFAULT '' COMMENT 'who created this record',
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'when the record is updated',
  `update_by` varchar(255) NOT NULL DEFAULT '' COMMENT 'who updated this record',
  `is_del` tinyint NOT NULL DEFAULT '0' COMMENT '0-normal, 1-deleted',
  PRIMARY KEY (`id`),
  UNIQUE KEY `folder_no_uk` (`folder_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Virtual folder';

CREATE TABLE IF NOT EXISTS user_vfolder (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_no` varchar(64) NOT NULL COMMENT 'user no',
  `folder_no` varchar(64) NOT NULL COMMENT 'folder no',
  `ownership` varchar(15) NOT NULL DEFAULT 'OWNER' COMMENT 'ownership',
  `granted_by` varchar(64) NOT NULL COMMENT 'granted by (user_no)',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'when the record is created',
  `create_by` varchar(255) NOT NULL DEFAULT '' COMMENT 'who created this record',
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'when the record is updated',
  `update_by` varchar(255) NOT NULL DEFAULT '' COMMENT 'who updated this record',
  `is_del` tinyint NOT NULL DEFAULT '0' COMMENT '0-normal, 1-deleted',
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_folder_uk` (`user_no`,`folder_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='User and Virtual folder join table';

CREATE TABLE IF NOT EXISTS file_vfolder (
  `id` int NOT NULL AUTO_INCREMENT,
  `folder_no` varchar(64) NOT NULL COMMENT 'folder no',
  `uuid` varchar(64) NOT NULL COMMENT 'file''s uuid',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'when the record is created',
  `create_by` varchar(255) NOT NULL DEFAULT '' COMMENT 'who created this record',
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'when the record is updated',
  `update_by` varchar(255) NOT NULL DEFAULT '' COMMENT 'who updated this record',
  `is_del` tinyint NOT NULL DEFAULT '0' COMMENT '0-normal, 1-deleted',
  PRIMARY KEY (`id`),
  UNIQUE KEY `folder_file_uk` (`folder_no`,`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='File and vfolder join table';

CREATE TABLE IF NOT EXISTS file_task (
  id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT "primary key",
  task_no varchar(32) NOT NULL DEFAULT '' COMMENT 'task no',
  user_no VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'user no',
  type varchar(20) NOT NULL DEFAULT '' COMMENT 'task type',
  status varchar(20) NOT NULL DEFAULT '' COMMENT 'task status',
  description varchar(100) NOT NULL DEFAULT '' COMMENT 'task description',
  file_key VARCHAR(32) NOT NULL DEFAULT '' comment 'file key',
  start_time TIMESTAMP NULL COMMENT 'start time',
  end_time TIMESTAMP NULL COMMENT 'end time',
  remark VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'remark',
  create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'when the record is created',
  create_by VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'who created this record',
  update_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'when the record is updated',
  update_by VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'who updated this record',
  is_del TINYINT NOT NULL DEFAULT '0' COMMENT '0-normal, 1-deleted',
  KEY (task_no),
  KEY (file_key)
) ENGINE=InnoDB COMMENT 'File Task';

CREATE TABLE IF NOT EXISTS fs_group (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL COMMENT 'group name',
  `base_folder` varchar(255) NOT NULL COMMENT 'base folder',
  `mode` int NOT NULL DEFAULT '2' COMMENT '1-read, 2-read/write',
  `type` varchar(32) NOT NULL DEFAULT 'USER' COMMENT 'FsGroup Type',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'when the record is created',
  `create_by` varchar(255) NOT NULL DEFAULT '' COMMENT 'who created this record',
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'when the record is updated',
  `update_by` varchar(255) NOT NULL DEFAULT '' COMMENT 'who updated this record',
  `is_del` tinyint NOT NULL DEFAULT '0' COMMENT '0-normal, 1-deleted',
  `size` bigint NOT NULL DEFAULT '0' COMMENT 'size in bytes',
  `scan_time` timestamp NULL DEFAULT NULL COMMENT 'previous scan time',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='FileSystem group, used to differentiate which base folder or mounted folder should be used';
