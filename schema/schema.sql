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
  `user_group` int NOT NULL DEFAULT '1' COMMENT 'the group that the file belongs to, 0-public, 1-private',
  `fs_group_id` int NOT NULL DEFAULT '0' COMMENT 'id of fs_group',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'when the record is created',
  `create_by` varchar(255) NOT NULL DEFAULT '' COMMENT 'who created this record',
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'when the record is updated',
  `update_by` varchar(255) NOT NULL DEFAULT '' COMMENT 'who updated this record',
  `is_del` tinyint NOT NULL DEFAULT '0' COMMENT '0-normal, 1-deleted',
  `file_type` varchar(6) NOT NULL DEFAULT 'FILE' COMMENT 'file type: FILE, DIR',
  `parent_file` varchar(64) NOT NULL DEFAULT '' COMMENT 'parent file uuid',
  `fstore_file_id` varchar(32) NOT NULL DEFAULT '' COMMENT 'mini-fstore file id',
  `thumbnail` varchar(32) NOT NULL DEFAULT '' COMMENT 'thumbnail, mini-fstore file id',
  `uploader_no` varchar(32) NOT NULL DEFAULT '' COMMENT 'user no of uploader',
  `sensitive_mode` varchar(1) NOT NULL DEFAULT 'N' COMMENT 'sensitive file, Y/N',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uuid_uk` (`uuid`),
  KEY `parent_file_type_idx` (`parent_file`,`file_type`),
  KEY `uploader_id_idx` (`uploader_id`),
  KEY `name` (`name`(128)),
  KEY `uploader_no_idx` (`uploader_no`)
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
  `username` varchar(50) DEFAULT '' COMMENT 'username',
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

CREATE TABLE IF NOT EXISTS gallery (
  `id` int NOT NULL AUTO_INCREMENT,
  `gallery_no` varchar(32) NOT NULL DEFAULT '' COMMENT 'gallery no',
  `user_no` varchar(64) NOT NULL DEFAULT '' COMMENT 'user no',
  `name` varchar(255) NOT NULL COMMENT 'gallery name',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'when the record is created',
  `create_by` varchar(255) NOT NULL DEFAULT '' COMMENT 'who created this record',
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'when the record is updated',
  `update_by` varchar(255) NOT NULL DEFAULT '' COMMENT 'who updated this record',
  `is_del` tinyint NOT NULL DEFAULT '0' COMMENT '0-normal, 1-deleted',
  `dir_file_key` varchar(64) NOT NULL DEFAULT '' COMMENT 'directory file_key (vfm)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `gallery_no_uniq` (`gallery_no`),
  KEY `idx_dir_file_key` (`dir_file_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Gallery'

CREATE TABLE IF NOT EXISTS gallery_image (
  `id` int NOT NULL AUTO_INCREMENT,
  `gallery_no` varchar(32) NOT NULL DEFAULT '' COMMENT 'gallery no',
  `image_no` varchar(32) NOT NULL DEFAULT '' COMMENT 'image no',
  `name` varchar(255) NOT NULL COMMENT 'name of the file',
  `file_key` varchar(64) NOT NULL COMMENT 'file key (vfm)',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'when the record is created',
  `create_by` varchar(255) NOT NULL DEFAULT '' COMMENT 'who created this record',
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'when the record is updated',
  `update_by` varchar(255) NOT NULL DEFAULT '' COMMENT 'who updated this record',
  `is_del` tinyint NOT NULL DEFAULT '0' COMMENT '0-normal, 1-deleted',
  `status` varchar(20) NOT NULL DEFAULT 'NORMAL' COMMENT 'status',
  PRIMARY KEY (`id`),
  UNIQUE KEY `image_no_uniq` (`image_no`),
  UNIQUE KEY `gallery_no_file_key_uk` (`gallery_no`,`file_key`),
  KEY `gallery_no_idx` (`gallery_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Gallery Image';

CREATE TABLE IF NOT EXISTS gallery_user_access (
  `id` int NOT NULL AUTO_INCREMENT,
  `gallery_no` varchar(32) NOT NULL DEFAULT '' COMMENT 'gallery no',
  `user_no` varchar(64) NOT NULL DEFAULT '' COMMENT 'user''s no',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'when the record is created',
  `create_by` varchar(255) NOT NULL DEFAULT '' COMMENT 'who created this record',
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'when the record is updated',
  `update_by` varchar(255) NOT NULL DEFAULT '' COMMENT 'who updated this record',
  `is_del` tinyint NOT NULL DEFAULT '0' COMMENT '0-normal, 1-deleted',
  PRIMARY KEY (`id`),
  UNIQUE KEY `gallery_user` (`gallery_no`,`user_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='User access to gallery';